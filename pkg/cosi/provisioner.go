package cosi

import (
	"context"
	"encoding/json"
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/container-object-storage-interface-provisioner-sidecar/pkg/provisioner"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Provisioner implements lib-bucket-provisioner callbacks
type Provisioner struct {
	client    client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	Logger    *logrus.Entry
	Namespace string
}

// RunProvisioner will run the COSI provisioner
func RunProvisioner(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) error {

	provisionerName := options.COSIProvisionerName()
	log := logrus.WithField("provisioner", provisionerName)
	log.Info("COSI Provisioner - start..")

	p := &Provisioner{
		client:    client,
		scheme:    scheme,
		recorder:  recorder,
		Logger:    log,
		Namespace: options.Namespace,
	}

	cosiProv, err := provisioner.NewDefaultCOSIProvisionerServer(options.CosiDriverAddress,
		&IdentityServer{
			provisioner: provisionerName,
		}, p);
	if err != nil {
		return err
	}

	log.Info("running noobaa cosi provisioner ", provisionerName)
	go func() {
		util.Panic(cosiProv.Run(util.Context()))
	}()

	return nil
}

// ProvisionerCreateBucket is an idempotent method for creating buckets
// It is expected to create the same bucket given a bucketName and protocol
// If the bucket already exists, then it MUST return codes.AlreadyExists
// Return values
//    nil -                   Bucket successfully created
//    codes.AlreadyExists -   Bucket already exists. No more retries
//    non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (p *Provisioner) ProvisionerCreateBucket(ctx context.Context,
	req *cosi.ProvisionerCreateBucketRequest) (*cosi.ProvisionerCreateBucketResponse, error) {
	log := p.Logger

	protocol := req.GetProtocol()
	if protocol == nil {
		msg := "Protocol is nil"
		log.Errorf(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}
	s3 := protocol.GetS3()
	if s3 == nil {
		msg :=  "S3 protocol is missing, only S3 is supported"
		log.Errorf(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}
	bucketName := req.GetName()
	log.Infof("ProvisionerCreateBucket: got request to provision bucket %q", req.Name)
	r, err := NewBucketRequest(p, req, nil)
	if err != nil {
		return nil, err
	}

	if r.SysClient.NooBaa.DeletionTimestamp != nil {
		finalizersArray := r.SysClient.NooBaa.GetFinalizers()
		if util.Contains(nbv1.GracefulFinalizer, finalizersArray) {
			msg := "NooBaa is in deleting state, new requests will be ignored"
			log.Errorf(msg)
			return nil, status.Error(codes.Internal, msg)
		}
	}
	// TODO: we need to better handle the case that a bucket was created, but Provision failed
	// right now we will fail on create bucket when Provision is called the second time
	err = r.CreateBucket(p, req)
	if err != nil {
		return nil, err
	}
	log.Infof("ProvisionerCreateBucket: Successfully created backend Bucket %q", bucketName)

	return &cosi.ProvisionerCreateBucketResponse{
		BucketId: bucketName,
	}, nil
}

// ProvisionerDeleteBucket is an idempotent method for deleting buckets
// It is expected to delete the same bucket given a bucketName
// Return values
//    nil -                   Bucket successfully deleted
//    non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (p *Provisioner) ProvisionerDeleteBucket(ctx context.Context,
	req *cosi.ProvisionerDeleteBucketRequest) (*cosi.ProvisionerDeleteBucketResponse, error) {
	log := p.Logger
	log.Infof("ProvisionerDeleteBucket: got request to delete bucket %q", req.GetBucketId())

	r, err := NewBucketRequest(p, nil, req)
	if err != nil {
		return nil, err
	}
	err = r.DeleteBucket()
	if err != nil {
		return nil, err
	}
	log.Infof("ProvisionerDeleteBucket: Successfully deleted Bucket %q", req.GetBucketId())

	return &cosi.ProvisionerDeleteBucketResponse{}, nil
}

// ProvisionerGrantBucketAccess is an idempotent method for granting bucket access
// It is expected to grant access to the given bucketName
// Return values
//    nil -                   Access was successfully granted
//    non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (p *Provisioner) ProvisionerGrantBucketAccess(ctx context.Context,
	req *cosi.ProvisionerGrantBucketAccessRequest) (*cosi.ProvisionerGrantBucketAccessResponse, error) {
	log := p.Logger
	log.Infof("ProvisionerGrantBucketAccess: got request to grant user %q access to bucket %q", req.AccountName, req.BucketId)
	r, err := NewAccountRequest(p, req, nil)
	if err != nil {
		return nil, err
	}
	keys, err := r.CreateAccount()
	if err != nil {
		return nil, err
	}
	log.Infof("ProvisionerGrantBucketAccess: Successfully created backend account %q", r.AccountName)

	return &cosi.ProvisionerGrantBucketAccessResponse{
		AccountId:   r.AccountName,
		Credentials: fetchUserCredentials(*keys),
	}, nil
}

// ProvisionerRevokeBucketAccess is an idempotent method for revoking bucket access
// It is expected to revoke access of a given accountId to the given bucketId
// Return values
//    nil -                   Access successfully revoked
//    non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (p *Provisioner) ProvisionerRevokeBucketAccess(ctx context.Context,
	req *cosi.ProvisionerRevokeBucketAccessRequest) (*cosi.ProvisionerRevokeBucketAccessResponse, error) {
	log := p.Logger
	log.Infof("ProvisionerRevokeBucketAccess: got request to revoke user %q access to bucket %q", req.AccountId, req.BucketId)
	r, err := NewAccountRequest(p, nil, req)
	if err != nil {
		return nil, err
	}
	err = r.DeleteAccount()
	if err != nil {
		return nil, err
	}
	log.Infof("ProvisionerGrantBucketAccess: Successfully deleted Account %q", r.AccountName)
	return &cosi.ProvisionerRevokeBucketAccessResponse{}, nil
}


// APIRequest is the context of handling a single api request (bucket or account)
type APIRequest struct {
	Provisioner *Provisioner
	BucketName  string
	AccountName  string
	SysClient   *system.Client
	Policy 		*nbv1.BucketClassSpec
}

// NewBucketRequest initializes a cosi bucket request
func NewBucketRequest(
	p *Provisioner,
	bucketCreateReq *cosi.ProvisionerCreateBucketRequest,
	bucketDelReq    *cosi.ProvisionerDeleteBucketRequest,
) (*APIRequest, error) {
	log := p.Logger
	sysClient, err := system.Connect(false)
	if err != nil {
		return nil, err
	}

	r := &APIRequest{
		Provisioner: p,
		SysClient:   sysClient,
	}

	if bucketCreateReq != nil {
		r.BucketName = bucketCreateReq.Name
		if bucketCreateReq.Parameters["policy"] != "" {
			err = json.Unmarshal([]byte(bucketCreateReq.Parameters["policy"]), &r.Policy)
			if err != nil {
				msg := fmt.Sprintf("failed to parse policy in COSI params %s",
					bucketCreateReq.Parameters["policy"],
				)	
				log.Error(msg)
				return nil, status.Error(codes.Internal, msg)
			}
		}
	} else if bucketDelReq != nil {
		r.BucketName = bucketDelReq.BucketId
	}
	return r, nil
}

// NewAccountRequest initializes a cosi bucket access request
func NewAccountRequest(
	p *Provisioner,
	accountCreateReq *cosi.ProvisionerGrantBucketAccessRequest,
	accountDelReq    *cosi.ProvisionerRevokeBucketAccessRequest,
) (*APIRequest, error) {

	sysClient, err := system.Connect(false)
	if err != nil {
		return nil, err
	}

	r := &APIRequest{
		Provisioner: p,
		SysClient:   sysClient,
	}

	if accountCreateReq != nil {
		r.BucketName = accountCreateReq.BucketId
		r.AccountName = accountCreateReq.AccountName
	} else if accountDelReq != nil {
		r.BucketName = accountDelReq.BucketId
		r.AccountName = accountDelReq.AccountId
	}

	return r, nil
}

// CreateBucket creates a new cosi bucket
func (r *APIRequest) CreateBucket(
	p *Provisioner,
	bucketReq *cosi.ProvisionerCreateBucketRequest,
) error {

	log := r.Provisioner.Logger

	_, err := r.SysClient.NBClient.ReadBucketAPI(nb.ReadBucketParams{Name: r.BucketName})
	if err == nil {
		msg := fmt.Sprintf("Bucket %q already exists", r.BucketName)
		log.Error(msg)
		return status.Error(codes.InvalidArgument, msg)
	}
	if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode != "NO_SUCH_BUCKET" {
		return status.Error(codes.Internal, nbErr.RPCCode)
	}

	if r.Policy == nil {
		msg := fmt.Sprintf("BucketClass/Bucket policy not loaded %#v", r)
		log.Error(msg)
		return status.Error(codes.Internal, msg)
	}
	createBucketParams := &nb.CreateBucketParams{
		Name: r.BucketName,
	}
	if r.Policy.PlacementPolicy != nil {
		tierName, err := bucketclass.CreateTieringStructure(*r.Policy.PlacementPolicy, r.BucketName, r.SysClient.NBClient)
		createBucketParams.Tiering = tierName
		if err != nil {
			return fmt.Errorf("CreateTieringStructure for PlacementPolicy failed to create policy %q with error: %v", tierName, err)
		}
		createBucketParams.Tiering = tierName
	}

	// create NS bucket
	if r.Policy.NamespacePolicy != nil {
		namespacePolicyType := r.Policy.NamespacePolicy.Type
		var readResources []nb.NamespaceResourceFullConfig
		createBucketParams.Namespace = &nb.NamespaceBucketInfo{}

		if namespacePolicyType == nbv1.NSBucketClassTypeSingle {
			createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{ 
				Resource: r.Policy.NamespacePolicy.Single.Resource,
				Path:  bucketReq.Parameters["path"],
			}
			createBucketParams.Namespace.ReadResources = append(readResources, nb.NamespaceResourceFullConfig{ 
				Resource: r.Policy.NamespacePolicy.Single.Resource })
		} else if namespacePolicyType == nbv1.NSBucketClassTypeMulti {
			createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{ 
				Resource: r.Policy.NamespacePolicy.Multi.WriteResource,
				Path:  bucketReq.Parameters["path"],
			}
			for i := range r.Policy.NamespacePolicy.Multi.ReadResources {
				rr := r.Policy.NamespacePolicy.Multi.ReadResources[i]
				readResources = append(readResources, nb.NamespaceResourceFullConfig{Resource: rr})
			}
			createBucketParams.Namespace.ReadResources = readResources
		} else if namespacePolicyType == nbv1.NSBucketClassTypeCache {
			createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{
				Resource: r.Policy.NamespacePolicy.Cache.HubResource}
			createBucketParams.Namespace.ReadResources = append(readResources, nb.NamespaceResourceFullConfig{
				Resource: r.Policy.NamespacePolicy.Cache.HubResource})
			createBucketParams.Namespace.Caching = &nb.CacheSpec{TTLMs: r.Policy.NamespacePolicy.Cache.Caching.TTL}
			//cachePrefix := r.BucketClass.Spec.NamespacePolicy.Cache.Prefix
		}
	}
	err = r.SysClient.NBClient.CreateBucketAPI(*createBucketParams)

	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok {
			if nbErr.RPCCode == "BUCKET_ALREADY_EXISTS" {
				msg := fmt.Sprintf("Bucket %q already exists", r.BucketName)
				log.Error(msg)
				return status.Error(codes.InvalidArgument, msg)
			}
		}
		return fmt.Errorf("Failed to create bucket %q with error: %v", r.BucketName, err)
	}

	log.Infof("✅ Successfully created bucket %q", r.BucketName)
	return nil
}

// CreateAccount creates a new bar account
func (r *APIRequest) CreateAccount() (*nb.S3AccessKeys, error) {
	log := r.Provisioner.Logger
	_, err := r.SysClient.NBClient.ReadBucketAPI(nb.ReadBucketParams{Name: r.BucketName})
	if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode == "NO_SUCH_BUCKET" {
		return nil, status.Error(codes.Internal, nbErr.RPCCode)
	}
	accountInfo, err := r.SysClient.NBClient.CreateAccountAPI(nb.CreateAccountParams{
		Name:              r.AccountName,
		Email:             r.AccountName,
		DefaultResource:   "",
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
		return nil, err
	}

	var accessKeys nb.S3AccessKeys
	// if we didn't get the access keys in the create_account reply we might be talking to an older noobaa version (prior to 5.1)
	// in that case try to get it using read account
	if len(accountInfo.AccessKeys) == 0 {
		log.Info("CreateAccountAPI did not return access keys. calling ReadAccountAPI to get keys..")
		readAccountReply, err := r.SysClient.NBClient.ReadAccountAPI(nb.ReadAccountParams{Email: r.AccountName})
		if err != nil {
			return nil, err
		}
		accessKeys = readAccountReply.AccessKeys[0]
	} else {
		accessKeys = accountInfo.AccessKeys[0]
	}

	log.Infof("✅ Successfully created account %q with access to bucket %q", r.AccountName, r.BucketName)
	return &accessKeys, nil
}

// DeleteAccount deletes the bar account
func (r *APIRequest) DeleteAccount() error {

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

// DeleteBucket deletes the cosi bucket **including data**
func (r *APIRequest) DeleteBucket() error {
	// TODO we should support the different delete options - not completely closed by the KEP
	var err error
	log := r.Provisioner.Logger
	info, err := r.SysClient.NBClient.ReadBucketAPI(nb.ReadBucketParams{Name: r.BucketName})
	if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode != "NO_SUCH_BUCKET" {
		log.Warnf("Bucket to delete was not found %q", r.BucketName)
		return nil
	}
	log.Infof("deleting bucket %q", r.BucketName)
	if info.Namespace != nil {
		err = r.SysClient.NBClient.DeleteBucketAPI(nb.DeleteBucketParams{Name: r.BucketName})
	} else {
		err = r.SysClient.NBClient.DeleteBucketAndObjectsAPI(nb.DeleteBucketParams{Name: r.BucketName})
	}
	if err != nil {
		return fmt.Errorf("failed to delete bucket %q. got error: %v", r.BucketName, err)
	}

	log.Infof("✅ Successfully deleted bucket %q", r.BucketName)
	return nil
}

func fetchUserCredentials(accessKeys nb.S3AccessKeys) string {
	return fmt.Sprintf("[default]\naws_access_key %s\naws_secret_key %s", accessKeys.AccessKey, accessKeys.SecretKey)
}
