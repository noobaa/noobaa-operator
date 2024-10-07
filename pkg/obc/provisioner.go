package obc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner"
	obAPI "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api"
	obErrors "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	allNamespaces = ""
)

// Provisioner implements lib-bucket-provisioner callbacks
type Provisioner struct {
	client    client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	Logger    *logrus.Entry
	Namespace string
}

// RunProvisioner will run OBC provisioner
func RunProvisioner(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) error {

	provisionerName := options.ObjectBucketProvisionerName()
	log := logrus.WithField("provisioner", provisionerName)
	log.Info("OBC Provisioner - start..")

	config := util.KubeConfig()

	p := &Provisioner{
		client:    client,
		scheme:    scheme,
		recorder:  recorder,
		Logger:    log,
		Namespace: options.Namespace,
	}

	// Create and run the s3 provisioner controller.
	// It implements the Provisioner interface expected by the bucket
	// provisioning lib.
	libProv, err := provisioner.NewProvisioner(config, provisionerName, p, allNamespaces)
	if err != nil {
		log.Error(err, "failed to create noobaa provisioner")
		return err
	}

	errStrings := libProv.SetLabels(map[string]string{
		"app":           "noobaa",
		"noobaa-domain": options.SubDomainNS(),
	})
	if errStrings != nil {
		util.Panic(fmt.Errorf("SetLabels errors: %+v", errStrings))
	}

	log.Info("running noobaa provisioner ", provisionerName)
	stopChan := make(chan struct{})
	go func() {
		util.Panic(libProv.Run(stopChan))
	}()

	return nil
}

// GenerateUserID implements lib-bucket-provisioner callback to generate a user ID
func (p Provisioner) GenerateUserID(obc *nbv1.ObjectBucketClaim, ob *nbv1.ObjectBucket) (string, error) {
	// We do not implement this
	return "", nil
}

// Provision implements lib-bucket-provisioner callback to create a new bucket
func (p *Provisioner) Provision(bucketOptions *obAPI.BucketOptions) (*nbv1.ObjectBucket, error) {

	log := p.Logger
	log.Infof("Provision: got request to provision bucket %q", bucketOptions.BucketName)

	err := ValidateOBC(bucketOptions.ObjectBucketClaim, false)
	if err != nil {
		return nil, err
	}

	r, err := NewBucketRequest(p, nil, bucketOptions)
	if err != nil {
		return nil, err
	}

	if r.SysClient.NooBaa.DeletionTimestamp != nil {
		finalizersArray := r.SysClient.NooBaa.GetFinalizers()
		if util.Contains(finalizersArray, nbv1.GracefulFinalizer) {
			msg := "NooBaa is in deleting state, new requests will be ignored"
			log.Error(msg)
			return nil, obErrors.NewBucketExistsError(msg)
		}
	}
	// TODO: we need to better handle the case that a bucket was created, but Provision failed
	// right now we will fail on create bucket when Provision is called the second time
	err = r.CreateAndUpdateBucket(p, bucketOptions)
	if err != nil {
		return nil, err
	}

	// create account and give permissions for bucket
	err = r.CreateAccount()
	if err != nil {
		return nil, err
	}

	err = r.putBucketTagging()
	if err != nil {
		logrus.Warnf("failed executing putBucketTagging on bucket: %v, %v", r.BucketName, err)
	}
	return r.OB, nil
}

// Grant implements lib-bucket-provisioner callback to use an existing bucket
func (p *Provisioner) Grant(bucketOptions *obAPI.BucketOptions) (*nbv1.ObjectBucket, error) {

	log := p.Logger
	log.Infof("Grant: got request to grant access to bucket %q", bucketOptions.BucketName)

	r, err := NewBucketRequest(p, nil, bucketOptions)
	if err != nil {
		return nil, err
	}

	if r.SysClient.NooBaa.DeletionTimestamp != nil {
		finalizersArray := r.SysClient.NooBaa.GetFinalizers()
		if util.Contains(finalizersArray, nbv1.GracefulFinalizer) {
			msg := "NooBaa is in deleting state, new requests will be ignored"
			log.Error(msg)
			return nil, obErrors.NewBucketExistsError(msg)
		}
	}

	// create account and give permissions for bucket
	err = r.CreateAccount()
	if err != nil {
		return nil, err
	}

	return r.OB, nil
}

// Delete implements lib-bucket-provisioner callback to delete a bucket
func (p *Provisioner) Delete(ob *nbv1.ObjectBucket) error {

	log := p.Logger

	r, err := NewBucketRequest(p, ob, nil)
	if err != nil {
		return err
	}

	log.Infof("Delete: got request to delete bucket %q and account %q", r.BucketName, r.AccountName)

	if ob.Spec.ReclaimPolicy != nil && *ob.Spec.ReclaimPolicy != corev1.PersistentVolumeReclaimDelete {
		// if reclaim policy is not delete, just warn and continue with deletion.
		// we still want to delete because this could be part of resources cleanup after failed provisioning
		log.Warnf("got delete request but reclaim policy is not Delete. assuming this is cleanup after error. ob.Spec.ReclaimPolicy=%q", *ob.Spec.ReclaimPolicy)
	}

	err = r.DeleteBucket()
	if err != nil {
		return err
	}

	err = r.DeleteAccount()
	if err != nil {
		return err
	}

	return nil
}

// Revoke implements lib-bucket-provisioner callback to stop using an existing bucket
func (p *Provisioner) Revoke(ob *nbv1.ObjectBucket) error {

	log := p.Logger

	r, err := NewBucketRequest(p, ob, nil)
	if err != nil {
		return err
	}

	log.Infof("Revoke: got request to revoke access to bucket %q for account %q", r.BucketName, r.AccountName)

	err = r.DeleteAccount()
	if err != nil {
		return err
	}

	return nil
}

// Update implements lib-bucket-provisioner callback to stop using an existing bucket additional config of OBC
func (p *Provisioner) Update(ob *nbv1.ObjectBucket) error {
	log := p.Logger
	log.Infof("Update: got request to Update bucket %q", ob.Name)

	err := ValidateOB(ob, false)
	if err != nil {
		return err
	}

	r, err := NewBucketRequest(p, ob, nil)
	if err != nil {
		return err
	}

	if err = r.updateReplicationPolicy(ob); err != nil {
		return err
	}

	if err = r.UpdateBucket(); err != nil {
		return err
	}

	return nil
}

// BucketRequest is the context of handling a single bucket request
type BucketRequest struct {
	Provisioner *Provisioner
	OB          *nbv1.ObjectBucket
	OBC         *nbv1.ObjectBucketClaim
	BucketName  string
	AccountName string
	SysClient   *system.Client
	BucketClass *nbv1.BucketClass
}

// NewBucketRequest initializes an obc bucket request
func NewBucketRequest(
	p *Provisioner,
	ob *nbv1.ObjectBucket,
	bucketOptions *obAPI.BucketOptions,
) (*BucketRequest, error) {

	sysClient, err := system.Connect(false)
	if err != nil {
		return nil, err
	}

	s3Hostname := sysClient.S3URL.Hostname()
	s3Port, err := strconv.Atoi(sysClient.S3URL.Port())
	if err != nil {
		return nil, fmt.Errorf("failed to parse s3 port %q. got error: %v", sysClient.S3URL, err)
	}

	r := &BucketRequest{
		Provisioner: p,
		OB:          ob,
		SysClient:   sysClient,
	}

	if r.OB == nil {
		r.OBC = bucketOptions.ObjectBucketClaim
		r.BucketName = bucketOptions.BucketName
		r.AccountName = fmt.Sprintf("obc-account.%s.%x@noobaa.io", r.BucketName, time.Now().Unix())

		bucketClassName := r.OBC.Spec.AdditionalConfig["bucketclass"]
		if bucketClassName == "" {
			bucketClassName = bucketOptions.Parameters["bucketclass"]
		}
		if bucketClassName == "" {
			return nil, fmt.Errorf("failed to find bucket class in OBC %s or storage class %s",
				r.OBC.Name,
				r.OBC.Spec.StorageClassName,
			)
		}

		bucketClass, exists := getBucketClass(r.OBC, bucketOptions, p.Namespace, util.KubeCheck)
		r.BucketClass = bucketClass
		if !exists {
			var msg string
			if bucketClass.GetName() == "" {
				msg = fmt.Sprintf(
					"failed to find bucket class name in OBC %s or storage class %s",
					r.OBC.Name,
					r.OBC.Spec.StorageClassName,
				)
			} else {
				msg = fmt.Sprintf("BucketClass %q not found in namespace %q", bucketClassName, p.Namespace)
			}

			p.recorder.Event(r.OBC, "Warning", "MissingBucketClass", msg)
			return nil, errors.New(msg)
		}
		if r.BucketClass.Status.Phase != nbv1.BucketClassPhaseReady {
			msg := fmt.Sprintf("BucketClass %q is not ready", bucketClassName)
			p.recorder.Event(r.OBC, "Warning", "BucketClassNotReady", msg)
			return nil, errors.New(msg)
		}
		additionalConfig := r.OBC.Spec.AdditionalConfig
		if additionalConfig == nil {
			additionalConfig = map[string]string{}
		}

		r.OB = &nbv1.ObjectBucket{
			Spec: nbv1.ObjectBucketSpec{
				Connection: &nbv1.ObjectBucketConnection{
					Endpoint: &nbv1.ObjectBucketEndpoint{
						BucketHost:           s3Hostname,
						BucketPort:           s3Port,
						BucketName:           r.BucketName,
						AdditionalConfigData: additionalConfig,
					},
					AdditionalState: map[string]string{
						"account":               r.AccountName, // needed for delete flow
						"bucketclass":           bucketClassName,
						"bucketclassgeneration": fmt.Sprintf("%d", r.BucketClass.Generation),
					},
				},
			},
		}
	} else {
		if ob.Spec.Connection == nil || ob.Spec.Connection.Endpoint == nil {
			return nil, fmt.Errorf("ObjectBucket has no connection/endpoint info %+v", ob)
		}
		r.BucketName = ob.Spec.Connection.Endpoint.BucketName
		r.AccountName = ob.Spec.AdditionalState["account"]
		bucketClassName := ob.Spec.AdditionalState["bucketclass"]

		bucketClass, exists := getBucketClass(r.OBC, bucketOptions, p.Namespace, util.KubeCheck)
		if !exists {
			p.Logger.Warnf("BucketClass %q not found in namespace %q", bucketClassName, p.Namespace)
		}
		r.BucketClass = bucketClass
		if r.OB.Spec.Connection.Endpoint.AdditionalConfigData == nil {
			r.OB.Spec.Connection.Endpoint.AdditionalConfigData = map[string]string{}
		}
	}

	return r, nil
}

// CreateAndUpdateBucket creates the obc bucket and update
func (r *BucketRequest) CreateAndUpdateBucket(
	p *Provisioner,
	bucketOptions *obAPI.BucketOptions,
) error {

	log := r.Provisioner.Logger

	_, err := r.SysClient.NBClient.ReadBucketAPI(nb.ReadBucketParams{Name: r.BucketName})
	if err == nil {
		msg := fmt.Sprintf("Bucket %q already exists", r.BucketName)
		log.Error(msg)
		return obErrors.NewBucketExistsError(msg)
	}
	if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode != "NO_SUCH_BUCKET" {
		return err
	}

	if r.BucketClass == nil {
		return fmt.Errorf("BucketClass not loaded %#v", r)
	}

	log.Infof("Provisioner: replication policy %s", r.BucketClass.Spec.ReplicationPolicy)
	var replicationParams *nb.BucketReplicationParams
	replicationPolicy := r.BucketClass.Spec.ReplicationPolicy
	// if OBC has replication policy set it to replication policy instead of the bucketclass
	if r.OBC.Spec.AdditionalConfig["replicationPolicy"] != "" {
		replicationPolicy = r.OBC.Spec.AdditionalConfig["replicationPolicy"]
	}
	if replicationPolicy != "" {
		if replicationParams, _, err = PrepareReplicationParams(r.BucketName, replicationPolicy, false); err != nil {
			return err
		}
	}

	createBucketParams := &nb.CreateBucketParams{
		Name: r.BucketName,
		BucketClaim: &nb.BucketClaimInfo{
			BucketClass: r.BucketClass.Name,
			Namespace:   r.OBC.Namespace,
		},
	}
	if r.BucketClass.Spec.PlacementPolicy != nil {
		if r.OBC.Spec.AdditionalConfig["path"] != "" {
			return fmt.Errorf("Could not create OBC %q with inner path while missing namespace bucketclass", r.OBC.Name)
		}

		tierName, err := bucketclass.CreateTieringStructure(*r.BucketClass.Spec.PlacementPolicy, r.BucketName, r.SysClient.NBClient)
		if err != nil {
			return fmt.Errorf("CreateTieringStructure for PlacementPolicy failed to create policy %q with error: %v", tierName, err)
		}
		createBucketParams.Tiering = tierName
	}

	// create NS bucket
	if r.BucketClass.Spec.NamespacePolicy != nil {
		createBucketParams.Namespace = bucketclass.CreateNamespaceBucketInfoStructure(*r.BucketClass.Spec.NamespacePolicy, r.OBC.Spec.AdditionalConfig["path"])
	}

	err = r.SysClient.NBClient.CreateBucketAPI(*createBucketParams)

	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok {
			if nbErr.RPCCode == "BUCKET_ALREADY_EXISTS" {
				msg := fmt.Sprintf("Bucket %q already exists", r.BucketName)
				log.Error(msg)
				return obErrors.NewBucketExistsError(msg)
			}
		}
		return fmt.Errorf("Failed to create bucket %q with error: %v", r.BucketName, err)
	}

	log.Infof("PutBucketReplicationAPI params: %v", replicationParams)

	// update replication policy
	if replicationParams != nil {
		err = r.SysClient.NBClient.PutBucketReplicationAPI(*replicationParams)
		if err != nil {
			return fmt.Errorf("Provisioner Failed to update replication on bucket %q with error: %v", r.BucketName, err)
		}
	}

	log.Infof("✅ Successfully created bucket %q", r.BucketName)

	return r.UpdateBucket()
}

// UpdateBucket update obc bucket
func (r *BucketRequest) UpdateBucket() error {

	log := r.Provisioner.Logger

	if r.BucketClass == nil {
		return r.LogAndGetError("BucketClass not loaded %#v", r)
	}

	bucket, err := r.SysClient.NBClient.ReadBucketAPI(nb.ReadBucketParams{Name: r.BucketName})
	if err != nil {
		return r.LogAndGetError("Bucket %q doesn't exist", r.BucketName)
	}

	quotaConfig, err := GetQuotaConfig(r.BucketName, &r.BucketClass.Spec, r.OB.Spec.Endpoint.AdditionalConfigData, log)
	if err != nil {
		return err
	}
	if quotaConfig.IsEqual(bucket.Quota) {
		r.Provisioner.Logger.Infof("UpdateBucket: no changes in quota config")
	} else {
		createBucketParams := &nb.CreateBucketParams{
			Name:  r.BucketName,
			Quota: quotaConfig,
		}

		err = r.SysClient.NBClient.UpdateBucketAPI(*createBucketParams)
		if err != nil {
			return r.LogAndGetError("failed to update bucket %q with error: %v", r.BucketName, err)
		}
	}

	log.Infof("✅ Successfully update bucket %q", r.BucketName)
	return nil
}

// GetQuotaConfig Gets minimum QuotaConfig based on OBC config and BucketClass
func GetQuotaConfig(bucketName string, BucketClassSpec *nbv1.BucketClassSpec, obAdditionalConfig map[string]string, log *logrus.Entry) (*nb.QuotaConfig, error) {
	var obMaxSize, obMaxObjects, bcMaxSize, bcMaxObjects string
	var minMaxSize, minMaxObjects int64

	// get quota config from ob
	if obAdditionalConfig != nil {
		obMaxSize = obAdditionalConfig["maxSize"]
		obMaxObjects = obAdditionalConfig["maxObjects"]
	}
	// get quota config from bucketclass
	if BucketClassSpec.Quota != nil {
		bcMaxSize = BucketClassSpec.Quota.MaxSize
		bcMaxObjects = BucketClassSpec.Quota.MaxObjects
	}

	log.Debugf("getQuotaConfig: bucket %q, obMaxSize %q, obMaxObjects %q, bcMaxSize %q, bcMaxObjects %q",
		bucketName, obMaxSize, obMaxObjects, bcMaxSize, bcMaxObjects)

	quota := nb.QuotaConfig{}
	// In order to remove the bucket quota, returns empty quota config if bucketclass and ob have no quota config
	if bcMaxSize == "" && bcMaxObjects == "" && obMaxSize == "" && obMaxObjects == "" {
		return &quota, nil
	}

	//Parse bucketclass quota and transform to quotaConfig
	if bcMaxSize != "" {
		// Validator catchs parsing error
		quantity, _ := resource.ParseQuantity(bcMaxSize)
		minMaxSize = quantity.Value()
	}
	if bcMaxObjects != "" {
		// Validator catchs parsing error
		num, _ := strconv.ParseInt(bcMaxObjects, 10, 32)
		minMaxObjects = num
	}

	//Parse obc quota transform to quotaConfig
	if obMaxSize != "" {
		// Validator catchs parsing error
		quantity, _ := resource.ParseQuantity(obMaxSize)
		obcMaxSizeInt := quantity.Value()
		//Calculate min maxSize
		if minMaxSize == 0 || obcMaxSizeInt < minMaxSize {
			minMaxSize = obcMaxSizeInt
		}
	}

	if obMaxObjects != "" {
		// Validator catchs parsing error
		obcMaxObjectsInt, _ := strconv.ParseInt(obMaxObjects, 10, 32)
		//Calculate min maxObjects
		if minMaxObjects == 0 || obcMaxObjectsInt < minMaxObjects {
			minMaxObjects = obcMaxObjectsInt
		}
	}

	if minMaxSize > 0 {
		f, u := nb.GetBytesAndUnits(minMaxSize, 2)
		quota.Size = &nb.SizeQuotaConfig{Value: f, Unit: u}
	}
	if minMaxObjects > 0 {
		quota.Quantity = &nb.QuantityQuotaConfig{Value: int(minMaxObjects)}
	}

	return &quota, nil
}

// LogAndGetError error handler. prints error message to log and returns error
func (r *BucketRequest) LogAndGetError(format string, a ...interface{}) error {
	log := r.Provisioner.Logger
	msg := fmt.Sprintf(format, a...)
	log.Error(msg)
	return errors.New(msg)
}

// CreateAccount creates the obc account
func (r *BucketRequest) CreateAccount() error {

	log := r.Provisioner.Logger
	var defaultResource string
	if r.BucketClass.Spec.PlacementPolicy != nil {
		defaultResource = r.BucketClass.Spec.PlacementPolicy.Tiers[0].BackingStores[0]
	}

	var nsfsAccountConfig *nbv1.AccountNsfsConfig
	// Validation is already performed as part of ValidateOBC before CreateAccount is ever called
	// ...but we revalidate to satisfy the linter.
	if r.OBC.Spec.AdditionalConfig["nsfsAccountConfig"] != "" {
		nsfsAccountConfig = &nbv1.AccountNsfsConfig{}
		err := json.Unmarshal([]byte(r.OBC.Spec.AdditionalConfig["nsfsAccountConfig"]), nsfsAccountConfig)
		if err != nil {
			return fmt.Errorf("failed to parse NSFS config %q: %w", r.OBC.Spec.AdditionalConfig["nsfsAccountConfig"], err)
		}
		// We prefer to make sure this account is only used for its appropriate NSFS operations
		nsfsAccountConfig.NewBucketsPath = ""
		nsfsAccountConfig.NsfsOnly = true
		// -1 is the default CLI value which we use to indicate that the UID/GID should not be set
		// 0 cannot be used since it is a valid GID/UID value
		var IDNullifier = -1
		if nsfsAccountConfig.UID == &IDNullifier {
			nsfsAccountConfig.UID = nil
			nsfsAccountConfig.GID = nil
		}
	}

	accountInfo, err := r.SysClient.NBClient.CreateAccountAPI(nb.CreateAccountParams{
		Name:  r.AccountName,
		Email: r.AccountName,
		// defaultResource is left as-is only because AllowBucketCreate is false
		DefaultResource: defaultResource,
		HasLogin:        false,
		S3Access:        true,
		// If this field is to be changed, DefaultResource above will need to be modified as well
		AllowBucketCreate: false,
		BucketClaimOwner:  r.BucketName,
		NsfsAccountConfig: nsfsAccountConfig,
	})
	if err != nil {
		return err
	}

	var accessKeys nb.S3AccessKeys
	// if we didn't get the access keys in the create_account reply we might be talking to an older noobaa version (prior to 5.1)
	// in that case try to get it using read account
	if len(accountInfo.AccessKeys) == 0 {
		log.Info("CreateAccountAPI did not return access keys. calling ReadAccountAPI to get keys..")
		readAccountReply, err := r.SysClient.NBClient.ReadAccountAPI(nb.ReadAccountParams{Email: r.AccountName})
		if err != nil {
			return err
		}
		accessKeys = readAccountReply.AccessKeys[0]
	} else {
		accessKeys = accountInfo.AccessKeys[0]
	}

	r.OB.Spec.Authentication = &nbv1.ObjectBucketAuthentication{
		AccessKeys: &nbv1.ObjectBucketAccessKeys{
			AccessKeyID:     string(accessKeys.AccessKey),
			SecretAccessKey: string(accessKeys.SecretKey),
		},
	}

	log.Infof("✅ Successfully created account %q with access to bucket %q", r.AccountName, r.BucketName)
	return nil
}

// DeleteAccount deletes the obc account
func (r *BucketRequest) DeleteAccount() error {

	log := r.Provisioner.Logger
	log.Infof("deleting account %q", r.AccountName)
	err := r.SysClient.NBClient.DeleteAccountAPI(nb.DeleteAccountParams{Email: r.AccountName})

	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode == "NO_SUCH_ACCOUNT" {
			log.Warnf("Account to delete was not found %q", r.AccountName)
		} else {
			return fmt.Errorf("failed to delete account %q. got error: %v", r.AccountName, err)
		}
	} else {
		log.Infof("✅ Successfully deleted account %q", r.AccountName)
	}

	return nil
}

// DeleteBucket deletes the obc bucket **including data**
func (r *BucketRequest) DeleteBucket() error {

	// TODO delete bucket data!!!
	var err error
	log := r.Provisioner.Logger
	log.Infof("deleting bucket %q", r.BucketName)
	if r.BucketClass.Spec.NamespacePolicy != nil {
		err = r.SysClient.NBClient.DeleteBucketAPI(nb.DeleteBucketParams{Name: r.BucketName})
	} else {
		err = r.SysClient.NBClient.DeleteBucketAndObjectsAPI(nb.DeleteBucketParams{Name: r.BucketName})
	}

	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode == "NO_SUCH_BUCKET" {
			log.Warnf("Bucket to delete was not found %q", r.BucketName)
		} else {
			return fmt.Errorf("failed to delete bucket %q. got error: %v", r.BucketName, err)
		}
	} else {
		log.Infof("✅ Successfully deleted bucket %q", r.BucketName)
	}

	return nil
}

// putBucketTagging calls s3 putBucketTagging on the created noobaa bucket
func (r *BucketRequest) putBucketTagging() error {

	client := &http.Client{Transport: util.InsecureHTTPTransport}
	s3Status := &r.SysClient.NooBaa.Status.Services.ServiceS3
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			r.SysClient.SecretAdmin.StringData["AWS_ACCESS_KEY_ID"],
			r.SysClient.SecretAdmin.StringData["AWS_SECRET_ACCESS_KEY"],
			"",
		),
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(s3Status.InternalDNS[0]),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       client,
	}
	s3Session, err := session.NewSession(s3Config)
	if err != nil {
		return err
	}
	s3Client := s3.New(s3Session)

	// convert labels to tagging array
	taggingArray := []*s3.Tag{}
	for key, value := range r.OBC.Labels {
		// no need to put tagging of these labels
		if !util.Contains([]string{"app", "noobaa-domain", "bucket-provisioner"}, key) {
			keyPointer := key
			valuePointer := value
			taggingArray = append(taggingArray, &s3.Tag{Key: &keyPointer, Value: &valuePointer})
		}
	}
	logrus.Infof("put bucket tagging on bucket: %s tagging: %+v ", r.BucketName, taggingArray)
	if len(taggingArray) == 0 {
		return nil
	}
	_, err = s3Client.PutBucketTagging(&s3.PutBucketTaggingInput{
		Bucket: &r.BucketName,
		Tagging: &s3.Tagging{
			TagSet: taggingArray,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// PrepareReplicationParams validates and prepare the replication params
func PrepareReplicationParams(bucketName string, replicationPolicy string, update bool) (*nb.BucketReplicationParams, *nb.DeleteBucketReplicationParams, error) {

	var replicationRules nb.ReplicationPolicy

	if replicationPolicy == "" && update {
		deleteReplicationParams := &nb.DeleteBucketReplicationParams{
			Name: bucketName,
		}
		return nil, deleteReplicationParams, nil
	}

	err := json.Unmarshal([]byte(replicationPolicy), &replicationRules)
	if err != nil {
		return nil, nil, fmt.Errorf("PrepareReplicationParams: Failed to parse replication json %q: %v", replicationRules, err)
	}

	replicationParams := &nb.BucketReplicationParams{
		Name:              bucketName,
		ReplicationPolicy: replicationRules,
	}

	return replicationParams, nil, nil
}

// updateReplicationPolicy validates and prepare the replication params
func (r *BucketRequest) updateReplicationPolicy(ob *nbv1.ObjectBucket) error {
	log := r.Provisioner.Logger
	newReplication := ob.Spec.Endpoint.AdditionalConfigData["replicationPolicy"]
	log.Infof("updateReplicationPolicy: new Replication %q detected on ob: %v", newReplication, ob.Name)

	updateReplicationParams, deleteReplicationParams, err := PrepareReplicationParams(r.BucketName, newReplication, true)
	if err != nil {
		return err
	}

	// delete bucket replication
	if deleteReplicationParams != nil {
		log.Infof("updateReplicationPolicy: deleting replication of bucket %q", ob.Name)
		err = r.SysClient.NBClient.DeleteBucketReplicationAPI(*deleteReplicationParams)
		if err != nil {
			return fmt.Errorf("Provisioner Failed to remove replication of bucket %q with error: %v", ob.Name, err)
		}
		return nil
	}

	// update replication policy
	if updateReplicationParams != nil {
		log.Infof("updateReplicationPolicy: updating replication on ob: %q replicationParams: %+v", ob.Name, updateReplicationParams)
		err = r.SysClient.NBClient.PutBucketReplicationAPI(*updateReplicationParams)
		if err != nil {
			return fmt.Errorf("Provisioner Failed to update replication on bucket %q with error: %v", ob.Name, err)
		}
	}
	log.Infof("updateReplicationPolicy: updated replication successfully")
	return nil
}

// getBucketClass takes an OBC, bucketoptions and provisioner namespace and returns the bucketClass
//
// If BucketClass name is not specified in the OBC, then the empty string is returned with exists=false
// If BucketClass name is specified in the OBC, then:
// - if the bucketclass is found in the obc namespace, then that bucketclass is returned
// with exists=true
// - if the bucketclass is found in the provisioner namespace, then that buckeclass is
// returned with exists=true
// - if the bucketclass is not found in the obc namespace or the provisioner namespace, then the
// bucketclass with namespace set to provisioner namespace is returned with exists=false
func getBucketClass(
	obc *nbv1.ObjectBucketClaim,
	bucketOptions *obAPI.BucketOptions,
	provisionerNS string,
	checkExists func(client.Object) bool,
) (bc *nbv1.BucketClass, exists bool) {
	bucketClass := &nbv1.BucketClass{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "",
			Namespace: "",
		},
	}

	if obc == nil {
		return bucketClass, false
	}

	bucketclassName := obc.Spec.AdditionalConfig["bucketclass"]
	if bucketclassName == "" && bucketOptions != nil {
		bucketclassName = bucketOptions.Parameters["bucketclass"]
	}
	if bucketclassName == "" {
		return bucketClass, false
	}

	bucketClass.SetName(bucketclassName)

	// Find the bucketclass in the same namespace as the OBC
	bucketClass.SetNamespace(obc.Namespace)
	if checkExists(bucketClass) {
		return bucketClass, true
	}

	// Find the bucketclass in the provisioner namespace
	bucketClass.SetNamespace(provisionerNS)
	if checkExists(bucketClass) {
		return bucketClass, true
	}

	return bucketClass, false
}
