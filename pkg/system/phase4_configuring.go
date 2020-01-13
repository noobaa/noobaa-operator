package system

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/nb"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	/*if err := r.NBClient.RegisterToCluster(); err != nil {
		return err
	}*/
	if err := r.ReconcileObject(r.DeploymentEndpoint, r.SetDesiredDeploymentEndpoint); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.HPAEndpoint, r.SetDesiredHPAEndpoint); err != nil {
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
	if err := r.ReconcileObjectOptional(r.PrometheusRule, nil); err != nil {
		return err
	}
	if err := r.ReconcileObjectOptional(r.ServiceMonitor, nil); err != nil {
		return err
	}
	if err := r.ReconcileReadSystem(); err != nil {
		return err
	}
	if err := r.ReconcileDeploymentEndpointStatus(); err != nil {
		return err
	}

	return nil
}

// ReconcileDeploymentEndpointStatus creates/updates the endpoints deployment
func (r *Reconciler) ReconcileDeploymentEndpointStatus() error {
	if !util.KubeCheck(r.DeploymentEndpoint) {
		return fmt.Errorf("Could not load endpoint deployment")
	}
	if r.DeploymentEndpoint.Status.ReadyReplicas == 0 {
		return fmt.Errorf("First endpoint is not ready yet")
	}

	podSpec := &r.DeploymentEndpoint.Spec.Template.Spec
	virtualHosts := []string{}
	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		if c.Name == "endpoint" {
			for j := range c.Env {
				e := c.Env[j]
				if e.Name == "VIRTUAL_HOSTS" {
					virtualHosts = append(virtualHosts, strings.Fields(e.Value)...)
				}
			}
		}
	}

	r.NooBaa.Status.Endpoints = &nbv1.EndpointsStatus{
		ReadyCount:   r.DeploymentEndpoint.Status.ReadyReplicas,
		VirtualHosts: virtualHosts,
	}

	return nil
}

// SetDesiredDeploymentEndpoint updates the endpoint deployment as desired for reconciling
func (r *Reconciler) SetDesiredDeploymentEndpoint() {
	r.DeploymentEndpoint.Spec.Selector.MatchLabels["noobaa-s3"] = r.Request.Name
	r.DeploymentEndpoint.Spec.Template.Labels["noobaa-s3"] = r.Request.Name

	endpointsSpec := r.NooBaa.Spec.Endpoints
	podSpec := &r.DeploymentEndpoint.Spec.Template.Spec
	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		switch c.Name {
		case "endpoint":
			c.Image = r.NooBaa.Status.ActualImage
			if endpointsSpec != nil && endpointsSpec.Resources != nil {
				c.Resources = *endpointsSpec.Resources
			}

			for j := range c.Env {
				switch c.Env[j].Name {
				case "MGMT_URL":
					c.Env[j].Value = fmt.Sprintf(`https://%s.%s.svc`,
						r.ServiceMgmt.Name, r.Request.Namespace)

				case "MONGODB_URL":
					c.Env[j].Value = fmt.Sprintf(`mongodb://%s-0.%s/nbcore`,
						r.NooBaaDB.Name, r.NooBaaDB.Spec.ServiceName)

				case "VIRTUAL_HOSTS":
					hosts := []string{}
					for _, addr := range r.NooBaa.Status.Services.ServiceS3.InternalDNS {
						// Ignore mailformed addresses
						if u, err := url.Parse(addr); err == nil {
							if host, _, err := net.SplitHostPort(u.Host); err == nil {
								hosts = append(hosts, host)
							}
						}
					}
					for _, addr := range r.NooBaa.Status.Services.ServiceS3.ExternalDNS {
						// Ignore mailformed addresses
						if u, err := url.Parse(addr); err == nil {
							if host, _, err := net.SplitHostPort(u.Host); err == nil {
								hosts = append(hosts, host)
							}
						}
					}
					if endpointsSpec != nil {
						hosts = append(hosts, endpointsSpec.AdditionalVirtualHosts...)
					}
					c.Env[j].Value = fmt.Sprintf(strings.Join(hosts[:], " "))

				case "ENDPOINT_GROUP_ID":
					c.Env[j].Value = fmt.Sprint(r.NooBaa.UID)

					// Commented as of Guy's requests, feature needs further deliberation
					// case "REGION":
					// 	if r.NooBaa.Spec.Endpoints.Region != nil {
					// 		c.Env[j].Value = *r.NooBaa.Spec.Endpoints.Region
					// 	}
				}
			}
		}
	}
}

// SetDesiredHPAEndpoint updates the endpoint horizontal pod autoscaler as desired for reconciling
func (r *Reconciler) SetDesiredHPAEndpoint() {
	endpointsSpec := r.NooBaa.Spec.Endpoints
	if endpointsSpec != nil {
		r.HPAEndpoint.Spec.MinReplicas = &endpointsSpec.MinCount
		r.HPAEndpoint.Spec.MaxReplicas = endpointsSpec.MaxCount
	}
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
		r.NooBaa.Status.Accounts.Admin.SecretRef.Name = r.SecretAdmin.Name
		r.NooBaa.Status.Accounts.Admin.SecretRef.Namespace = r.SecretAdmin.Namespace
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

	if r.CephObjectstoreUser.UID != "" {
		log.Infof("CephObjectstoreUser %q created.  creating default backing store on ceph objectstore", r.CephObjectstoreUser.Name)
		if err := r.prepareCephBackingStore(); err != nil {
			return err
		}
	} else if r.CloudCreds.UID != "" {
		log.Infof("CredentialsRequest %q created.  creating default backing store on ceph objectstore", r.CloudCreds.Name)
		if err := r.prepareAWSBackingStore(); err != nil {
			return err
		}
	} else {

		return nil
	}

	r.Own(r.DefaultBackingStore)
	if err := r.Client.Create(r.Ctx, r.DefaultBackingStore); err != nil {
		log.Errorf("got error on DefaultBackingStore creation. error: %v", err)
		return err
	}
	return nil
}

func (r *Reconciler) prepareAWSBackingStore() error {
	// after we have cloud credential request, wait for credentials secret
	cloudCredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.CloudCreds.Spec.SecretRef.Name,
			Namespace: r.CloudCreds.Spec.SecretRef.Namespace,
		},
	}

	util.KubeCheck(cloudCredsSecret)
	if cloudCredsSecret.UID == "" {
		// TODO: we need to figure out why secret is not created, and react accordingly
		// e.g. maybe we are running on azure but our CredentialsRequest is for AWS
		r.Logger.Infof("Secret %q was not created yet by cloud-credentials operator. retry on next reconcile..", r.CloudCreds.Spec.SecretRef.Name)
		return fmt.Errorf("cloud credentials secret %q is not ready yet", r.CloudCreds.Spec.SecretRef.Name)
	}
	r.Logger.Infof("Secret %s was created succesfully by cloud-credentials operator", r.CloudCreds.Spec.SecretRef.Name)

	// create the acutual S3 bucket
	region, err := util.GetAWSRegion()
	if err != nil {
		r.Recorder.Eventf(r.NooBaa, corev1.EventTypeWarning, "DefaultBackingStoreFailure",
			"Failed to get AWSRegion. using	 us-east-1 as the default region. %q", err)
		region = "us-east-1"
	}
	r.Logger.Infof("identified aws region %s", region)
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			cloudCredsSecret.StringData["aws_access_key_id"],
			cloudCredsSecret.StringData["aws_secret_access_key"],
			"",
		),
		Region: &region,
	}

	bucketName := r.DefaultBackingStore.Spec.AWSS3.TargetBucket
	if err := r.createS3BucketForBackingStore(s3Config, bucketName); err != nil {
		return err
	}

	// create backing store
	r.DefaultBackingStore.Spec.Type = nbv1.StoreTypeAWSS3
	r.DefaultBackingStore.Spec.AWSS3.Secret.Name = cloudCredsSecret.Name
	r.DefaultBackingStore.Spec.AWSS3.Secret.Namespace = cloudCredsSecret.Namespace
	r.DefaultBackingStore.Spec.AWSS3.Region = region
	return nil
}

func (r *Reconciler) prepareCephBackingStore() error {

	secretName := "rook-ceph-object-user-" + r.CephObjectstoreUser.Spec.Store + "-" + r.CephObjectstoreUser.Name

	// get access\secret keys from user secret
	cephObjectUserSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: options.Namespace,
		},
	}

	util.KubeCheck(cephObjectUserSecret)
	if cephObjectUserSecret.UID == "" {
		// TODO: we need to figure out why secret is not created, and react accordingly
		// e.g. maybe we are running on azure but our CredentialsRequest is for AWS
		r.Logger.Infof("Ceph object user secret %q was not created yet. retry on next reconcile..", secretName)
		return fmt.Errorf("Ceph object user secret %q is not ready yet", secretName)
	}

	endpoint := "http://rook-ceph-rgw-" + r.CephObjectstoreUser.Spec.Store + "." + options.Namespace + ":80"
	region := "us-east-1"
	forcePathStyle := true
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			cephObjectUserSecret.StringData["AccessKey"],
			cephObjectUserSecret.StringData["SecretKey"],
			"",
		),
		Endpoint:         &endpoint,
		Region:           &region,
		S3ForcePathStyle: &forcePathStyle,
	}
	bucketName := r.generateBackingStoreTargetName()
	if err := r.createS3BucketForBackingStore(s3Config, bucketName); err != nil {
		return err
	}

	// create backing store
	r.DefaultBackingStore.Spec.Type = nbv1.StoreTypeS3Compatible
	r.DefaultBackingStore.Spec.S3Compatible = &nbv1.S3CompatibleSpec{
		Secret:           corev1.SecretReference{Name: secretName, Namespace: options.Namespace},
		TargetBucket:     bucketName,
		Endpoint:         endpoint,
		SignatureVersion: nbv1.S3SignatureVersionV4,
	}
	return nil
}

func (r *Reconciler) generateBackingStoreTargetName() string {
	const MaxNameLength = 63
	tsMilli := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	name := "nb." + tsMilli
	if r.RouteMgmt.Spec.Host != "" {
		name += "." + r.RouteMgmt.Spec.Host
	}
	if len(name) > MaxNameLength {
		name = name[:MaxNameLength]
	}
	return name
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

	err := r.Client.Create(r.Ctx, r.OBCStorageClass)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) createS3BucketForBackingStore(s3Config *aws.Config, bucketName string) error {
	s3Session, err := session.NewSession(s3Config)
	if err != nil {
		return err
	}
	s3Client := s3.New(s3Session)

	r.Logger.Infof("creating bucket %s", bucketName)
	createBucketOutout, err := s3Client.CreateBucket(&s3.CreateBucketInput{Bucket: &bucketName})
	if err != nil {
		awsErr, isAwsErr := err.(awserr.Error)
		if isAwsErr && awsErr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
			r.Logger.Infof("bucket was already created. continuing")
		} else {
			r.Logger.Errorf("got error when trying to create bucket %s. error: %v", bucketName, err)
			return err
		}
	} else {
		r.Logger.Infof("Successfully created bucket %s. result = %v", bucketName, createBucketOutout)
	}
	return nil
}

// ReconcileReadSystem calls read_system on noobaa server and stores the result
func (r *Reconciler) ReconcileReadSystem() error {
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
