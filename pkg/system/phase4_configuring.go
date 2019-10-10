package system

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/nb"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ReconcilePhaseConfiguring runs the reconcile phase
func (r *Reconciler) ReconcilePhaseConfiguring() error {

	r.SetPhase(
		nbv1.SystemPhaseConfiguring,
		"SystemPhaseConfiguring",
		"noobaa operator started phase 4/4 - \"Configuring\"",
	)

	if err := r.ReconcileSecretOp(); err != nil {
		return err
	}
	if err := r.ReconcileSecretAdmin(); err != nil {
		return err
	}
	if err := r.ReconcileDefaultBackingStore(); err != nil {
		return err
	}
	if err := r.ReconcileDefaultBucketClass(); err != nil {
		return err
	}
	if err := r.ReconcileOBCStorageClass(); err != nil {
		return err
	}

	if err := r.ReconcileObject(r.PrometheusRule, nil); err != nil {
		if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
			r.Logger.Printf("No PrometheusRule CRD existing, skip creating PrometheusRules\n")
		} else {
			return err
		}
	}

	if err := r.ReconcileObject(r.ServiceMonitor, nil); err != nil {
		if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
			r.Logger.Printf("No ServiceMonitor CRD existing, skip creating ServiceMonitor\n")
		} else {
			return err
		}
	}

	if err := r.ReconileReadSystem(); err != nil {
		return err
	}

	return nil
}

// ReconcileSecretOp creates a new system in the noobaa server if not created yet.
func (r *Reconciler) ReconcileSecretOp() error {

	// log := r.Logger.WithName("ReconcileSecretOp")
	util.KubeCheck(r.SecretOp)

	if r.SecretOp.StringData["auth_token"] != "" {
		r.NBClient.SetAuthToken(r.SecretOp.StringData["auth_token"])
		return nil
	}

	if r.SecretOp.StringData["email"] == "" {
		r.SecretOp.StringData["email"] = options.AdminAccountEmail
	}

	if r.SecretOp.StringData["password"] == "" {
		r.SecretOp.StringData["password"] = util.RandomBase64(16)
		r.Own(r.SecretOp)
		err := r.Client.Create(r.Ctx, r.SecretOp)
		if err != nil {
			return err
		}
	}

	res, err := r.NBClient.CreateAuthAPI(nb.CreateAuthParams{
		System:   r.Request.Name,
		Role:     "admin",
		Email:    r.SecretOp.StringData["email"],
		Password: r.SecretOp.StringData["password"],
	})
	if err == nil {
		// TODO this recovery flow does not allow us to get OperatorToken like CreateSystem
		r.SecretOp.StringData["auth_token"] = res.Token
	} else {
		res, err := r.NBClient.CreateSystemAPI(nb.CreateSystemParams{
			Name:     r.Request.Name,
			Email:    r.SecretOp.StringData["email"],
			Password: r.SecretOp.StringData["password"],
		})
		if err != nil {
			return err
		}
		// TODO use res.OperatorToken after https://github.com/noobaa/noobaa-core/issues/5635
		r.SecretOp.StringData["auth_token"] = res.Token
	}
	r.NBClient.SetAuthToken(r.SecretOp.StringData["auth_token"])
	return r.Client.Update(r.Ctx, r.SecretOp)
}

// ReconcileSecretAdmin creates the admin secret
func (r *Reconciler) ReconcileSecretAdmin() error {

	log := r.Logger.WithField("func", "ReconcileSecretAdmin")

	util.KubeCheck(r.SecretAdmin)

	// already exists - we can skip
	if r.SecretAdmin.UID != "" {
		return nil
	}

	log.Infof("listing accounts")
	res, err := r.NBClient.ListAccountsAPI()
	if err != nil {
		return err
	}
	var account *nb.AccountInfo
	for _, a := range res.Accounts {
		if a.Email == options.AdminAccountEmail {
			account = a
		}
	}
	if account == nil || account.AccessKeys == nil || len(account.AccessKeys) <= 0 {
		return fmt.Errorf("admin account has no access keys yet")
	}

	r.SecretAdmin.StringData["system"] = r.NooBaa.Name
	r.SecretAdmin.StringData["email"] = options.AdminAccountEmail
	r.SecretAdmin.StringData["password"] = r.SecretOp.StringData["password"]
	r.SecretAdmin.StringData["AWS_ACCESS_KEY_ID"] = account.AccessKeys[0].AccessKey
	r.SecretAdmin.StringData["AWS_SECRET_ACCESS_KEY"] = account.AccessKeys[0].SecretKey
	r.Own(r.SecretAdmin)
	err = r.Client.Create(r.Ctx, r.SecretAdmin)
	if err != nil {
		return err
	}

	r.NooBaa.Status.Accounts.Admin.SecretRef.Name = r.SecretAdmin.Name
	r.NooBaa.Status.Accounts.Admin.SecretRef.Namespace = r.SecretAdmin.Namespace
	return nil
}

// ReconcileDefaultBackingStore attempts to get credentials to cloud storage using the cloud-credentials operator
// and use it for the default backing store
func (r *Reconciler) ReconcileDefaultBackingStore() error {

	log := r.Logger.WithField("func", "ReconcileDefaultBackingStore")

	util.KubeCheck(r.DefaultBackingStore)
	// backing store already exists - we can skip
	// TODO: check if there are any changes to reconcile
	if r.DefaultBackingStore.UID != "" {
		log.Infof("Backing store %s already exists. skipping ReconcileCloudCredentials", r.DefaultBackingStore.Name)
		return nil
	}

	// after we have cloud credential request, wait for credentials secret
	cloudCredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.CloudCreds.Spec.SecretRef.Name,
			Namespace: r.CloudCreds.Spec.SecretRef.Namespace,
		},
	}

	if r.CloudCreds.UID == "" {
		log.Info("CredentialsRequest was not created. probably not supported. skipping ReconcileCloudCredentials")
		// err := reconcileCephObjecetStoreUser()
		// if err != nil {
		// 	return err
		// }
		// // change secret name to ceph objectstore user name
		// cloudCredsSecret.ObjectMeta.Name = r.CephObjectstoreUser.ObjectMeta.Name
		return nil
	}

	util.KubeCheck(cloudCredsSecret)
	if cloudCredsSecret.UID == "" {
		// TODO: we need to figure out why secret is not created, and react accordingly
		// e.g. maybe we are running on azure but our CredentialsRequest is for AWS
		log.Infof("Secret %s was not created yet by cloud-credentials operator. retry on next reconcile..", r.CloudCreds.Spec.SecretRef.Name)
		return fmt.Errorf("cloud credentials secret is not ready yet")
	}

	log.Infof("Secret %s was created succesfully by cloud-credentials operator", r.CloudCreds.Spec.SecretRef.Name)

	// create the acutual S3 bucket
	region := util.GetAWSRegion()
	r.Logger.Infof("identified aws region %s", region)
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			cloudCredsSecret.StringData["aws_access_key_id"],
			cloudCredsSecret.StringData["aws_secret_access_key"],
			"",
		),
		Region: &region,
	}
	s3Session, err := session.NewSession(s3Config)
	if err != nil {
		return err
	}
	S3Client := s3.New(s3Session)

	bucketName := r.DefaultBackingStore.Spec.AWSS3.TargetBucket
	log.Infof("creating bucket %s", bucketName)
	createBucketOutout, err := S3Client.CreateBucket(&s3.CreateBucketInput{Bucket: &bucketName})
	if err != nil {
		awsErr, isAwsErr := err.(awserr.Error)
		if isAwsErr && awsErr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
			log.Infof("bucket was already created. continuing")
		} else {
			log.Errorf("got error when trying to create bucket %s. error: %v", bucketName, err)
			return err
		}
	} else {
		log.Infof("Successfully created bucket %s. result = %v", bucketName, createBucketOutout)
	}

	// create backing store
	r.DefaultBackingStore.Spec.AWSS3.Secret.Name = cloudCredsSecret.Name
	r.DefaultBackingStore.Spec.AWSS3.Secret.Namespace = cloudCredsSecret.Namespace
	r.DefaultBackingStore.Spec.AWSS3.Region = region
	r.Own(r.DefaultBackingStore)
	err = r.Client.Create(r.Ctx, r.DefaultBackingStore)
	if err != nil {
		log.Errorf("got error on DefaultBackingStore creation. error: %v", err)
		return err
	}
	return nil
}

func (r *Reconciler) reconcileCephObjecetStoreUser() error {
	util.KubeCheck(r.CephObjectstoreUser)
	if r.CephObjectstoreUser.UID != "" {
		// already exists
		return nil
	}
	// list ceph objectstores and pick the first one
	cephObjectStoresList := &cephv1.CephObjectStoreList{}
	if !util.KubeList(cephObjectStoresList, &client.ListOptions{Namespace: options.Namespace}) {
		return fmt.Errorf("failed to list ceph objectstores")
	}
	if len(cephObjectStoresList.Items) == 0 {
		// no object stores
		return nil
	}
	// for now take the first one. need to decide what to do if multiple objectstores in one namespace
	storeName := cephObjectStoresList.Items[0].ObjectMeta.Name
	r.CephObjectstoreUser.Spec.Store = storeName

	// create ceph objectstore user
	err := r.Client.Create(r.Ctx, r.CephObjectstoreUser)
	if err != nil {
		r.Logger.Errorf("got error on CephObjectstoreUser creation. error: %v", err)
		return err
	}

	return nil
}

// ReconcileDefaultBucketClass creates the default bucket class
func (r *Reconciler) ReconcileDefaultBucketClass() error {

	util.KubeCheck(r.DefaultBucketClass)
	if r.DefaultBucketClass.UID != "" {
		return nil
	}

	r.DefaultBucketClass.Spec.PlacementPolicy = nbv1.PlacementPolicy{
		Tiers: []nbv1.Tier{{
			BackingStores: []nbv1.BackingStoreName{
				r.DefaultBackingStore.Name,
			},
		}},
	}

	r.Own(r.DefaultBucketClass)

	err := r.Client.Create(r.Ctx, r.DefaultBucketClass)
	if err != nil {
		return err
	}

	return nil
}

// ReconcileOBCStorageClass reconciles default OBC storage class for the system
func (r *Reconciler) ReconcileOBCStorageClass() error {

	util.KubeCheck(r.OBCStorageClass)
	if r.OBCStorageClass.UID != "" {
		return nil
	}

	r.OBCStorageClass.Parameters = map[string]string{
		"bucketclass": r.DefaultBucketClass.Name,
	}

	// unsetting BlockOwnerDeletion to acoid error when trying to own storage class:
	// "cannot set blockOwnerDeletion if an ownerReference refers to a resource you can't set finalizers on"
	r.Own(r.OBCStorageClass)
	r.OBCStorageClass.OwnerReferences[0].BlockOwnerDeletion = nil

	err := r.Client.Create(r.Ctx, r.OBCStorageClass)
	if err != nil {
		return err
	}

	return nil
}

// ReconileReadSystem calls read_system on noobaa server and stores the result
func (r *Reconciler) ReconileReadSystem() error {
	// update noobaa-core version in reconciler struct
	systemInfo, err := r.NBClient.ReadSystemAPI()
	if err != nil {
		r.Logger.Errorf("failed to read system info: %v", err)
		return err
	}
	r.SystemInfo = &systemInfo
	r.Logger.Infof("updating noobaa-core version to %s", systemInfo.Version)
	r.CoreVersion = systemInfo.Version
	return nil
}
