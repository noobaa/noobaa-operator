package obc

import (
	"encoding/json"
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

// Provision implements lib-bucket-provisioner callback to create a new bucket
func (p *Provisioner) Provision(bucketOptions *obAPI.BucketOptions) (*nbv1.ObjectBucket, error) {

	log := p.Logger
	log.Infof("Provision: got request to provision bucket %q", bucketOptions.BucketName)

	r, err := NewBucketRequest(p, nil, bucketOptions)
	if err != nil {
		return nil, err
	}

	if r.SysClient.NooBaa.DeletionTimestamp != nil {
		finalizersArray := r.SysClient.NooBaa.GetFinalizers()
		if util.Contains(nbv1.GracefulFinalizer, finalizersArray) {
			msg := "NooBaa is in deleting state, new requests will be ignored"
			log.Errorf(msg)
			return nil, obErrors.NewBucketExistsError(msg)
		}
	}
	// TODO: we need to better handle the case that a bucket was created, but Provision failed
	// right now we will fail on create bucket when Provision is called the second time
	err = r.CreateBucket(p, bucketOptions)
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
		if util.Contains(nbv1.GracefulFinalizer, finalizersArray) {
			msg := "NooBaa is in deleting state, new requests will be ignored"
			log.Errorf(msg)
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
	log.Infof("Update: got request to Update access to bucket %q", ob.Name)

	r, err := NewBucketRequest(p, ob, nil)
	if err != nil {
		return err
	}

	if err = r.updateReplicationPolicy(ob); err != nil {
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

		r.BucketClass = &nbv1.BucketClass{
			TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      bucketClassName,
				Namespace: p.Namespace,
			},
		}
		if !util.KubeCheck(r.BucketClass) {
			msg := fmt.Sprintf("BucketClass %q not found in provisioner namespace %q", bucketClassName, p.Namespace)
			p.recorder.Event(r.OBC, "Warning", "MissingBucketClass", msg)
			return nil, fmt.Errorf(msg)
		}
		if r.BucketClass.Status.Phase != nbv1.BucketClassPhaseReady {
			msg := fmt.Sprintf("BucketClass %q is not ready", bucketClassName)
			p.recorder.Event(r.OBC, "Warning", "BucketClassNotReady", msg)
			return nil, fmt.Errorf(msg)
		}
		r.OB = &nbv1.ObjectBucket{
			Spec: nbv1.ObjectBucketSpec{
				Connection: &nbv1.ObjectBucketConnection{
					Endpoint: &nbv1.ObjectBucketEndpoint{
						BucketHost:           s3Hostname,
						BucketPort:           s3Port,
						BucketName:           r.BucketName,
						AdditionalConfigData: map[string]string{},
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
		r.BucketClass = &nbv1.BucketClass{
			TypeMeta: metav1.TypeMeta{Kind: "BucketClass"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      bucketClassName,
				Namespace: p.Namespace,
			},
		}
		if !util.KubeCheck(r.BucketClass) {
			msg := fmt.Sprintf("BucketClass %q not found in provisioner namespace %q", bucketClassName, p.Namespace)
			p.recorder.Event(r.OBC, "Warning", "MissingBucketClass", msg)
			return nil, fmt.Errorf(msg)
		}
	}

	return r, nil
}

// CreateBucket creates the obc bucket
func (r *BucketRequest) CreateBucket(
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
		if replicationParams, _, err = r.prepareReplicationParams(replicationPolicy, false); err != nil {
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
		namespacePolicyType := r.BucketClass.Spec.NamespacePolicy.Type
		var readResources []nb.NamespaceResourceFullConfig
		createBucketParams.Namespace = &nb.NamespaceBucketInfo{}

		if namespacePolicyType == nbv1.NSBucketClassTypeSingle {
			createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Single.Resource,
				Path:     r.OBC.Spec.AdditionalConfig["path"],
			}
			createBucketParams.Namespace.ReadResources = append(readResources, nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Single.Resource})
		} else if namespacePolicyType == nbv1.NSBucketClassTypeMulti {
			createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Multi.WriteResource,
				Path:     r.OBC.Spec.AdditionalConfig["path"],
			}
			for i := range r.BucketClass.Spec.NamespacePolicy.Multi.ReadResources {
				rr := r.BucketClass.Spec.NamespacePolicy.Multi.ReadResources[i]
				readResources = append(readResources, nb.NamespaceResourceFullConfig{Resource: rr})
			}
			createBucketParams.Namespace.ReadResources = readResources
		} else if namespacePolicyType == nbv1.NSBucketClassTypeCache {
			createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Cache.HubResource}
			createBucketParams.Namespace.ReadResources = append(readResources, nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Cache.HubResource})
			createBucketParams.Namespace.Caching = &nb.CacheSpec{TTLMs: r.BucketClass.Spec.NamespacePolicy.Cache.Caching.TTL}
			//cachePrefix := r.BucketClass.Spec.NamespacePolicy.Cache.Prefix
		}
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
	return nil
}

// CreateAccount creates the obc account
func (r *BucketRequest) CreateAccount() error {

	log := r.Provisioner.Logger
	var defaultResource string
	if r.BucketClass.Spec.PlacementPolicy != nil {
		defaultResource = r.BucketClass.Spec.PlacementPolicy.Tiers[0].BackingStores[0]
	}
	accountInfo, err := r.SysClient.NBClient.CreateAccountAPI(nb.CreateAccountParams{
		Name:              r.AccountName,
		Email:             r.AccountName,
		DefaultResource:   defaultResource,
		HasLogin:          false,
		S3Access:          true,
		AllowBucketCreate: false,
		AllowedBuckets: nb.AccountAllowedBuckets{
			FullPermission: false,
			PermissionList: []string{r.BucketName},
		},
		BucketClaimOwner: r.BucketName,
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
			AccessKeyID:     accessKeys.AccessKey,
			SecretAccessKey: accessKeys.SecretKey,
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
		if !util.Contains(key, []string{"app", "noobaa-domain", "bucket-provisioner"}) {
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

// prepareReplicationParams validates and prepare the replication params
func (r *BucketRequest) prepareReplicationParams(replicationPolicy string, update bool) (*nb.BucketReplicationParams, *nb.DeleteBucketReplicationParams, error) {
	log := r.Provisioner.Logger
	deleteReplicationParams := &nb.DeleteBucketReplicationParams{
		Name: r.BucketName,
	}

	if replicationPolicy == "" && update {
		return nil, deleteReplicationParams, nil
	}

	var replicationRules []interface{}
	err := json.Unmarshal([]byte(replicationPolicy), &replicationRules)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to parse replication json %q: %v", replicationRules, err)
	}
	log.Infof("prepareReplicationParams: newReplication %+v", replicationRules)

	if len(replicationRules) == 0 {
		if update {
			return nil, deleteReplicationParams, nil
		}
		return nil, nil, fmt.Errorf("replication rules array of bucket %q is empty %q", r.BucketName, replicationRules)
	}

	replicationParams := &nb.BucketReplicationParams{
		Name:              r.BucketName,
		ReplicationPolicy: replicationRules,
	}

	log.Infof("prepareReplicationParams: validating replication: replicationParams: %+v", replicationParams)
	err = r.SysClient.NBClient.ValidateReplicationAPI(*replicationParams)
	if err != nil {
		rpcErr, isRPCErr := err.(*nb.RPCError)
		if isRPCErr && rpcErr.RPCCode == "INVALID_REPLICATION_POLICY" {
			return nil, nil, fmt.Errorf("Bucket replication configuration is invalid")
		}
		return nil, nil, fmt.Errorf("Provisioner Failed to validate replication of bucket %q with error: %v", r.BucketName, err)
	}
	log.Infof("prepareReplicationParams: validated replication successfully")
	return replicationParams, nil, nil
}

// updateReplicationPolicy validates and prepare the replication params
func (r *BucketRequest) updateReplicationPolicy(ob *nbv1.ObjectBucket) error {
	log := r.Provisioner.Logger
	newReplication := ob.Spec.Endpoint.AdditionalConfigData["replicationPolicy"]
	log.Infof("updateReplicationPolicy: new Replication %q detected on ob: %v", newReplication, ob.Name)

	updateReplicationParams, deleteReplicationParams, err := r.prepareReplicationParams(newReplication, true)
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
