package cosi

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/obc"
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

// TODO: check if this should be used
// const (
// 	allNamespaces = ""
// )

// Provisioner implements COSI driver callbacks
type Provisioner struct {
	client    client.Client
	scheme    *runtime.Scheme
	recorder  record.EventRecorder
	Logger    *logrus.Entry
	Namespace string
}

// RunProvisioner will run the COSI provisioner
func RunProvisioner(client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder) error {

	driverName := options.COSIDriverName()
	log := logrus.WithField("provisioner", driverName)
	log.Info("COSI Provisioner/Driver - start..")

	p := &Provisioner{
		client:    client,
		scheme:    scheme,
		recorder:  recorder,
		Logger:    log,
		Namespace: options.Namespace,
	}

	identityServer := &IdentityServer{
		driver: driverName,
	}
	// Create and run the s3 provisioner controller.
	// It implements the Provisioner interface expected by the cosi lib.
	cosiProv, err := provisioner.NewDefaultCOSIProvisionerServer(fmt.Sprintf("unix://%s", options.CosiDriverPath), identityServer, p)
	if err != nil {
		log.Error(err, "failed to create noobaa cosi provisioner/driver")
		return err
	}

	log.Info("running noobaa cosi provisioner/driver", driverName)
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		// remove socket in case it is already bound
		os.Remove(options.CosiDriverPath)
		util.Panic(cosiProv.Run(ctx))
	}()

	return nil
}

// DriverCreateBucket is an idempotent method for creating buckets
// It is expected to create the same bucket given a bucketName and protocol
// If the bucket already exists, then it MUST return codes.AlreadyExists
// Return values:
//
//	nil -                   Bucket successfully created
//	codes.AlreadyExists -   Bucket already exists. No more retries
//	non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (p *Provisioner) DriverCreateBucket(ctx context.Context,
	req *cosi.DriverCreateBucketRequest) (*cosi.DriverCreateBucketResponse, error) {
	log := p.Logger
	// TODO - this was removed, need to check if needs special handling
	// protocol := req.
	// if protocol == nil {
	// 	msg := "Protocol is nil"
	// 	log.Errorf(msg)
	// 	return nil, status.Error(codes.InvalidArgument, msg)
	// }
	// s3 := protocol.GetS3()
	// if s3 == nil {
	// 	msg := "S3 protocol is missing, only S3 is supported"
	// 	log.Errorf(msg)
	// 	return nil, status.Error(codes.InvalidArgument, msg)
	// }
	bucketName := req.GetName()
	log.Infof("DriverCreateBucket: got request to provision bucket %q", req.Name)

	r, err := NewBucketRequest(p, req, nil)
	if err != nil {
		return nil, err
	}

	err = ValidateCOSIBucketClaim(req.GetName(), r.Provisioner.Namespace, *r.BucketClass, false)
	log.Infof("DriverCreateBucket: ValidateCOSIBucketClaim err %q", err)

	if err != nil {
		return nil, err
	}

	if r.SysClient.NooBaa.DeletionTimestamp != nil {
		finalizersArray := r.SysClient.NooBaa.GetFinalizers()
		if util.Contains(finalizersArray, nbv1.GracefulFinalizer) {
			msg := "NooBaa is in deleting state, new requests will be ignored"
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
	}
	// TODO: we need to better handle the case that a bucket was created, but Provision failed
	// right now we will fail on create bucket when Provision is called the second time
	// Notice: update is not supported yet
	log.Infof("DriverCreateBucket: creating bucket ")

	err = r.CreateBucket(p, req)
	log.Infof("DriverCreateBucket: created bucket %q", err)

	if err != nil {
		return nil, err
	}
	log.Infof("DriverCreateBucket: Successfully created backend Bucket %q", bucketName)

	// TODO: in OBC we convert OBC user labels to tags, labels can be extracted out of the bucket claim after calling kubecheck
	return &cosi.DriverCreateBucketResponse{
		BucketId: bucketName,
	}, nil
}

// DriverDeleteBucket is an idempotent method for deleting buckets
// It is expected to delete the same bucket given a bucketName
// Return values:
//
//	nil -                   Bucket successfully deleted
//	non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (p *Provisioner) DriverDeleteBucket(ctx context.Context,
	req *cosi.DriverDeleteBucketRequest) (*cosi.DriverDeleteBucketResponse, error) {
	log := p.Logger
	log.Infof("DriverDeleteBucket: got request to delete bucket %q", req.GetBucketId())

	r, err := NewBucketRequest(p, nil, req)
	if err != nil {
		return nil, err
	}
	err = r.DeleteBucket()
	if err != nil {
		return nil, err
	}
	log.Infof("DriverDeleteBucket: Successfully deleted Bucket %q", req.GetBucketId())

	return &cosi.DriverDeleteBucketResponse{}, nil
}

// DriverGrantBucketAccess is an idempotent method for granting bucket access
// It is expected to grant access to the given bucketName
// Return values:
//
//	nil -                   Access was successfully granted
//	non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (p *Provisioner) DriverGrantBucketAccess(ctx context.Context,
	req *cosi.DriverGrantBucketAccessRequest) (*cosi.DriverGrantBucketAccessResponse, error) {
	log := p.Logger
	log.Infof("DriverGrantBucketAccess: got request to grant user %q access to bucket %q", req.Name, req.BucketId)
	r, err := NewAccountRequest(p, req, nil)
	if err != nil {
		return nil, err
	}

	if r.SysClient.NooBaa.DeletionTimestamp != nil {
		finalizersArray := r.SysClient.NooBaa.GetFinalizers()
		if util.Contains(finalizersArray, nbv1.GracefulFinalizer) {
			msg := "NooBaa is in deleting state, new requests will be ignored"
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
	}

	keys, err := r.CreateAccount()
	if err != nil {
		return nil, err
	}

	log.Infof("DriverGrantBucketAccess: Successfully created backend account %q", r.AccountName)
	return &cosi.DriverGrantBucketAccessResponse{
		AccountId:   r.AccountName,
		Credentials: fetchUserCredentials(*keys),
	}, nil
}

// DriverRevokeBucketAccess is an idempotent method for revoking bucket access
// It is expected to revoke access of a given accountId to the given bucketId
// Return values:
//
//	nil -                   Access successfully revoked
//	non-nil err -           Internal error                                [requeue'd with exponential backoff]
func (p *Provisioner) DriverRevokeBucketAccess(ctx context.Context,
	req *cosi.DriverRevokeBucketAccessRequest) (*cosi.DriverRevokeBucketAccessResponse, error) {
	log := p.Logger
	log.Infof("DriverRevokeBucketAccess: got request to revoke user %q access to bucket %q", req.AccountId, req.BucketId)
	r, err := NewAccountRequest(p, nil, req)
	if err != nil {
		return nil, err
	}
	err = r.DeleteAccount()
	if err != nil {
		return nil, err
	}
	log.Infof("DriverRevokeBucketAccess: Successfully deleted Account %q", r.AccountName)
	return &cosi.DriverRevokeBucketAccessResponse{}, nil
}

// APIRequest is the context of handling a single api request (bucket or account)
type APIRequest struct {
	Provisioner *Provisioner
	BucketName  string
	AccountName string
	SysClient   *system.Client
	BucketClass *nbv1.BucketClassSpec
}

// NewBucketRequest initializes a cosi bucket request
func NewBucketRequest(
	p *Provisioner,
	bucketCreateReq *cosi.DriverCreateBucketRequest,
	bucketDelReq *cosi.DriverDeleteBucketRequest,
) (*APIRequest, error) {
	log := p.Logger
	IsExternalRPCConnection := false
	if util.IsTestEnv() {
		IsExternalRPCConnection = true
	}
	sysClient, err := system.Connect(IsExternalRPCConnection)
	if err != nil {
		return nil, err
	}

	r := &APIRequest{
		Provisioner: p,
		SysClient:   sysClient,
	}

	if bucketCreateReq != nil {
		r.BucketName = bucketCreateReq.GetName()

		spec, errMsg := CreateBucketClassSpecFromParameters(bucketCreateReq.Parameters)
		if errMsg != "" {
			log.Error(errMsg)
			return nil, status.Error(codes.Internal, errMsg)
		}
		r.BucketClass = spec

		if r.BucketClass.PlacementPolicy == nil && r.BucketClass.NamespacePolicy == nil {
			msg := fmt.Sprintf("must provide at least one of placement policy or namespace policy %+v",
				r.BucketClass,
			)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
	} else if bucketDelReq != nil {
		r.BucketName = bucketDelReq.BucketId
	}
	return r, nil
}

// NewAccountRequest initializes a cosi bucket access request
func NewAccountRequest(
	p *Provisioner,
	accountCreateReq *cosi.DriverGrantBucketAccessRequest,
	accountDelReq *cosi.DriverRevokeBucketAccessRequest,
) (*APIRequest, error) {
	IsExternalRPCConnection := false
	if util.IsTestEnv() {
		IsExternalRPCConnection = true
	}
	sysClient, err := system.Connect(IsExternalRPCConnection)
	if err != nil {
		return nil, err
	}

	r := &APIRequest{
		Provisioner: p,
		SysClient:   sysClient,
	}

	if accountCreateReq != nil {
		r.BucketName = accountCreateReq.BucketId
		r.AccountName = accountCreateReq.Name
	} else if accountDelReq != nil {
		r.BucketName = accountDelReq.BucketId
		r.AccountName = accountDelReq.AccountId
	}

	return r, nil
}

// CreateBucket creates a new cosi bucket
func (r *APIRequest) CreateBucket(
	p *Provisioner,
	bucketReq *cosi.DriverCreateBucketRequest,
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

	if r.BucketClass == nil {
		msg := fmt.Sprintf("BucketClass/Bucket not loaded %#v", r)
		log.Error(msg)
		return status.Error(codes.Internal, msg)
	}

	log.Infof("COSI Provisioner: BucketClass %+v", r.BucketClass)
	var replicationParams *nb.BucketReplicationParams
	if r.BucketClass.ReplicationPolicy != "" {
		if replicationParams, _, err = obc.PrepareReplicationParams(r.BucketName, r.BucketClass.ReplicationPolicy, false); err != nil {
			return err
		}
	}

	createBucketParams := &nb.CreateBucketParams{
		Name: r.BucketName,
	}
	if r.BucketClass.PlacementPolicy != nil {
		tierName, err := bucketclass.CreateTieringStructure(*r.BucketClass.PlacementPolicy, r.BucketName, r.SysClient.NBClient)
		if err != nil {
			return fmt.Errorf("CreateTieringStructure for PlacementPolicy failed to create policy %q with error: %v", tierName, err)
		}
		createBucketParams.Tiering = tierName
	}

	// create NS bucket
	if r.BucketClass.NamespacePolicy != nil {
		// TODO: NO additional paramters recieved from driverCreateBucketRequest so no able to recieve bucket claim specific NSFS path
		// we need to recieve it from the labels of the bucket claim after calling kubecheck
		createBucketParams.Namespace = bucketclass.CreateNamespaceBucketInfoStructure(*r.BucketClass.NamespacePolicy, "")
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
		return fmt.Errorf("failed to create bucket %q with error: %v", r.BucketName, err)
	}
	log.Infof("✅ Successfully created bucket %q", r.BucketName)

	log.Infof("Update replication on created bucket - replication params: %v", replicationParams)

	// update replication policy
	if replicationParams != nil {
		err = r.SysClient.NBClient.PutBucketReplicationAPI(*replicationParams)
		if err != nil {
			return fmt.Errorf("Provisioner Failed to update replication on bucket %q with error: %v", r.BucketName, err)
		}
	}

	// TODO: this behavior is copied from OBC, we should avoid updating if config is empty
	quotaConfig, err := obc.GetQuotaConfig(r.BucketName, r.BucketClass, nil, log)
	if err != nil {
		return err
	}
	createBucketParams = &nb.CreateBucketParams{
		Name:  r.BucketName,
		Quota: quotaConfig,
	}

	err = r.SysClient.NBClient.UpdateBucketAPI(*createBucketParams)
	if err != nil {
		return fmt.Errorf("failed to update bucket %q with error: %v", r.BucketName, err)
	}
	return nil
}

// CreateAccount creates a new bar account
func (r *APIRequest) CreateAccount() (*nb.S3AccessKeys, error) {
	log := r.Provisioner.Logger
	_, err := r.SysClient.NBClient.ReadBucketAPI(nb.ReadBucketParams{Name: r.BucketName})
	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode == "NO_SUCH_BUCKET" {
			return nil, status.Error(codes.Internal, nbErr.RPCCode)
		}
		return nil, fmt.Errorf("CreateAccount: failed to create bucket %q. got error: %v", r.BucketName, err)
	}

	accountInfo, err := r.SysClient.NBClient.CreateAccountAPI(nb.CreateAccountParams{
		Name:              r.AccountName,
		Email:             r.AccountName,
		DefaultResource:   "",
		HasLogin:          false,
		S3Access:          true,
		AllowBucketCreate: false,
		BucketClaimOwner:  r.BucketName,
	})
	if err != nil {
		return nil, err
	}

	accessKeys := accountInfo.AccessKeys[0]

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
	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode == "NO_SUCH_BUCKET" {
			log.Warnf("Bucket to delete was not found %q", r.BucketName)
			return nil
		}
		return fmt.Errorf("failed to delete bucket %q. got error: %v", r.BucketName, err)
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

func fetchUserCredentials(accessKeys nb.S3AccessKeys) map[string]*cosi.CredentialDetails {
	s3Keys := make(map[string]string)
	s3Keys["accessKeyID"] = accessKeys.AccessKey
	s3Keys["accessSecretKey"] = accessKeys.SecretKey
	creds := &cosi.CredentialDetails{
		Secrets: s3Keys,
	}
	credDetails := make(map[string]*cosi.CredentialDetails)
	credDetails["s3"] = creds
	return credDetails
}
