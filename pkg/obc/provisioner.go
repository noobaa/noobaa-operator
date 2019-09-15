package obc

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner"
	obAPI "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api"
	obErrors "github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	allNamespaces = ""
)

// Provisioner implements lib-bucket-provisioner callbacks
type Provisioner struct {
	clientset *kubernetes.Clientset
	client    client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	Logger    *logrus.Entry
}

// RunProvisioner will run OBC provisioner
func RunProvisioner(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) error {
	util.Logger().Info("OBC Provisioner - start..")

	config := util.KubeConfig()
	clientset, err := kubernetes.NewForConfig(config)
	util.Panic(err)

	provisionerName := options.ObjectBucketProvisionerName()
	log := logrus.WithField("provisioner", provisionerName)

	p := &Provisioner{
		clientset: clientset,
		client:    client,
		scheme:    scheme,
		recorder:  recorder,
		Logger:    log,
	}

	// Create and run the s3 provisioner controller.
	// It implements the Provisioner interface expected by the bucket
	// provisioning lib.
	libProv, err := provisioner.NewProvisioner(config, provisionerName, p, allNamespaces)
	if err != nil {
		log.Error(err, "failed to create noobaa provisioner")
		return err
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
	SysClient   *system.Client
	BucketName  string
	AccountName string
	TierName    string
	PoolName    string
	OB          *nbv1.ObjectBucket
}

// NewBucketRequest initializes an obc bucket request
func NewBucketRequest(
	p *Provisioner,
	ob *nbv1.ObjectBucket,
	bucketOptions *obAPI.BucketOptions,
) (*BucketRequest, error) {

	sysClient, err := system.Connect()
	if err != nil {
		return nil, err
	}

	s3URL, err := url.Parse(sysClient.S3Client.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse s3 endpoint %q. got error: %v", sysClient.S3Client.Endpoint, err)
	}
	s3Hostname := s3URL.Hostname()
	s3Port, err := strconv.Atoi(s3URL.Port())
	if err != nil {
		return nil, fmt.Errorf("failed to parse s3 port %q. got error: %v", sysClient.S3Client.Endpoint, err)
	}

	var bucketName string
	var accountName string
	var tierName string
	var bsName string

	if ob == nil {
		bucketName = bucketOptions.BucketName
		accountName = fmt.Sprintf("obc-account.%s.%x@noobaa.io", bucketName, time.Now().Unix())
		bsName = bucketOptions.ObjectBucketClaim.Spec.AdditionalConfig["backingstore"]
		tierName = fmt.Sprintf("%s-%x", bucketName, time.Now().Unix())
		if bsName == "" {
			bsName = bucketOptions.Parameters["backingstore"]
			if bsName == "" {
				// TODO list pools instead of read system
				backingStoreList := &nbv1.BackingStoreList{}
				util.KubeList(backingStoreList, &client.ListOptions{Namespace: sysClient.NooBaa.Namespace})
				for i := range backingStoreList.Items {
					bs := &backingStoreList.Items[i]
					if bs.Status.Phase == nbv1.BackingStorePhaseReady {
						bsName = bs.Name
						break
					}
				}
				if bsName == "" {
					return nil, fmt.Errorf("no ready backing stores %+v", backingStoreList)
				}
			}
		}
		ob = &nbv1.ObjectBucket{
			Spec: nbv1.ObjectBucketSpec{
				Connection: &nbv1.ObjectBucketConnection{
					Endpoint: &nbv1.ObjectBucketEndpoint{
						BucketHost: s3Hostname,
						BucketPort: s3Port,
						BucketName: bucketOptions.BucketName,
						SSL:        s3URL.Scheme == "https:",
					},
					AdditionalState: map[string]string{
						"AccountName": accountName, // needed for delete flow
					},
				},
			},
		}
	} else {
		if ob.Spec.Connection == nil || ob.Spec.Connection.Endpoint == nil {
			return nil, fmt.Errorf("ObjectBucket has no connection/endpoint info %+v", ob)
		}
		bucketName = ob.Spec.Connection.Endpoint.BucketName
		accountName = ob.Spec.AdditionalState["AccountName"]
	}

	return &BucketRequest{
		Provisioner: p,
		SysClient:   sysClient,
		BucketName:  bucketName,
		AccountName: accountName,
		TierName:    tierName,
		PoolName:    bsName,
		OB:          ob,
	}, nil
}

// CreateBucket creates the obc bucket
func (r *BucketRequest) CreateBucket() error {

	log := r.Provisioner.Logger

	err := r.SysClient.NBClient.CreateTierAPI(nb.CreateTierParams{
		Name:          r.TierName,
		AttachedPools: []string{r.PoolName},
	})
	if err != nil {
		return fmt.Errorf("Failed to create tier %q with error: %v", r.TierName, err)
	}

	err = r.SysClient.NBClient.CreateTieringPolicyAPI(nb.CreateTieringPolicyParams{
		Name:  r.TierName,
		Tiers: []nb.TierItem{{Order: 0, Tier: r.TierName}},
	})
	if err != nil {
		return fmt.Errorf("Failed to create tier %q with error: %v", r.TierName, err)
	}

	err = r.SysClient.NBClient.CreateBucketAPI(nb.CreateBucketParams{
		Name:    r.BucketName,
		Tiering: r.TierName,
	})
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

	accountInfo, err := r.SysClient.NBClient.CreateAccountAPI(nb.CreateAccountParams{
		Name:              r.AccountName,
		Email:             r.AccountName,
		DefaultPool:       r.PoolName,
		HasLogin:          false,
		S3Access:          true,
		AllowBucketCreate: false,
		AllowedBuckets: nb.AccountAllowedBuckets{
			FullPermission: false,
			PermissionList: []string{r.BucketName},
		},
	})
	if err != nil {
		return err
	}

	if len(accountInfo.AccessKeys) == 0 {
		return fmt.Errorf("Create account did not return access keys")
	}

	r.OB.Spec.Authentication = &nbv1.ObjectBucketAuthentication{
		AccessKeys: &nbv1.ObjectBucketAccessKeys{
			AccessKeyID:     accountInfo.AccessKeys[0].AccessKey,
			SecretAccessKey: accountInfo.AccessKeys[0].SecretKey,
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
		return fmt.Errorf("failed to delete account %q. got error: %v", r.AccountName, err)
	}

	log.Infof("✅ Successfully deleted account %q", r.AccountName)
	return nil
}

// DeleteBucket deletes the obc bucket **including data**
func (r *BucketRequest) DeleteBucket() error {

	log := r.Provisioner.Logger
	log.Infof("deleting bucket %q", r.BucketName)

	// TODO delete bucket data!!!

	err := r.SysClient.NBClient.DeleteBucketAPI(nb.DeleteBucketParams{Name: r.BucketName})
	if err != nil {
		return fmt.Errorf("failed to delete bucket %q. got error: %v", r.BucketName, err)
	}

	log.Infof("✅ Successfully deleted bucket %q", r.BucketName)
	return nil
}
