package obc

import (
	"fmt"
	"strconv"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/nb"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/system"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"

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
	err = r.CreateBucket()
	if err != nil {
		return nil, err
	}

	// create account and give permissions for bucket
	err = r.CreateAccount()
	if err != nil {
		return nil, err
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

	if ob.Spec.ReclaimPolicy != nil &&
		(*ob.Spec.ReclaimPolicy == corev1.PersistentVolumeReclaimDelete ||
			*ob.Spec.ReclaimPolicy == corev1.PersistentVolumeReclaimRecycle) {
		err = r.DeleteBucket()
		if err != nil {
			return err
		}
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
	}

	return r, nil
}

// CreateBucket creates the obc bucket
func (r *BucketRequest) CreateBucket() error {

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
	createBucketParams := &nb.CreateBucketParams{
		Name: r.BucketName,
		BucketClaim: &nb.BucketClaimInfo{
			BucketClass: r.BucketClass.Name,
			Namespace:   r.OBC.Namespace,
		},
	}

	if r.BucketClass.Spec.PlacementPolicy != nil {
		tierName := fmt.Sprintf("%s.%x", r.BucketName, time.Now().Unix())
		tiers := []nb.TierItem{}

		for i := range r.BucketClass.Spec.PlacementPolicy.Tiers {
			tier := r.BucketClass.Spec.PlacementPolicy.Tiers[i]
			name := fmt.Sprintf("%s.%d", tierName, i)
			tiers = append(tiers, nb.TierItem{Order: int64(i), Tier: name})
			// we assume either mirror or spread but no mix and the bucket class controller rejects mixed classes.
			placement := "SPREAD"
			if tier.Placement == nbv1.TierPlacementMirror {
				placement = "MIRROR"
			}
			err := r.SysClient.NBClient.CreateTierAPI(nb.CreateTierParams{
				Name:          name,
				AttachedPools: tier.BackingStores,
				DataPlacement: placement,
			})
			if err != nil {
				return fmt.Errorf("Failed to create tier %q with error: %v", name, err)
			}
		}

		err = r.SysClient.NBClient.CreateTieringPolicyAPI(nb.TieringPolicyInfo{
			Name:  tierName,
			Tiers: tiers,
		})
		if err != nil {
			return fmt.Errorf("Failed to create tier %q with error: %v", tierName, err)
		}
		createBucketParams.Tiering = tierName
	}

	// create NS bucket
	if r.BucketClass.Spec.NamespacePolicy != nil {
		namespacePolicyType := r.BucketClass.Spec.NamespacePolicy.Type
		var readResources []string
		createBucketParams.Namespace = &nb.NamespaceBucketInfo{}
		if namespacePolicyType == nbv1.NSBucketClassTypeSingle {
			createBucketParams.Namespace.WriteResource = r.BucketClass.Spec.NamespacePolicy.Single.Resource
			createBucketParams.Namespace.ReadResources = append(readResources, r.BucketClass.Spec.NamespacePolicy.Single.Resource)
		} else if namespacePolicyType == nbv1.NSBucketClassTypeMulti {
			createBucketParams.Namespace.WriteResource = r.BucketClass.Spec.NamespacePolicy.Multi.WriteResource
			createBucketParams.Namespace.ReadResources = r.BucketClass.Spec.NamespacePolicy.Multi.ReadResources
		} else if namespacePolicyType == nbv1.NSBucketClassTypeCache {
			createBucketParams.Namespace.WriteResource = r.BucketClass.Spec.NamespacePolicy.Cache.HubResource
			createBucketParams.Namespace.ReadResources = append(readResources, r.BucketClass.Spec.NamespacePolicy.Cache.HubResource)
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

	log.Infof("✅ Successfully created bucket %q", r.BucketName)
	return nil
}

// CreateAccount creates the obc account
func (r *BucketRequest) CreateAccount() error {

	log := r.Provisioner.Logger
	var defaultPool string
	if r.BucketClass.Spec.PlacementPolicy != nil {
		defaultPool = r.BucketClass.Spec.PlacementPolicy.Tiers[0].BackingStores[0]
	}
	accountInfo, err := r.SysClient.NBClient.CreateAccountAPI(nb.CreateAccountParams{
		Name:              r.AccountName,
		Email:             r.AccountName,
		DefaultPool:       defaultPool,
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

	log := r.Provisioner.Logger
	log.Infof("deleting bucket %q", r.BucketName)
	err := r.SysClient.NBClient.DeleteBucketAndObjectsAPI(nb.DeleteBucketParams{Name: r.BucketName})

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
