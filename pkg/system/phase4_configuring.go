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
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	if err := r.ReconcileSystemSecrets(); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.DeploymentEndpoint, r.SetDesiredDeploymentEndpoint); err != nil {
		return err
	}
	if err := r.ReconcileHPAEndpoint(); err != nil {
		return err
	}
	if err := r.RegisterToCluster(); err != nil {
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
	if err := r.ReconcilePrometheusRule(); err != nil {
		return err
	}
	if err := r.ReconcileServiceMonitor(); err != nil {
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

// ReconcileSystemSecrets reconciles secrets used by the system and
// create the system if does not exists
func (r *Reconciler) ReconcileSystemSecrets() error {
	if r.JoinSecret == nil {
		if err := r.ReconcileObject(r.SecretAdmin, r.SetDesiredSecretAdmin); err != nil {
			return err
		}

		// Point the admin account secret reference to the admin secret.
		r.NooBaa.Status.Accounts.Admin.SecretRef.Name = r.SecretAdmin.Name
		r.NooBaa.Status.Accounts.Admin.SecretRef.Namespace = r.SecretAdmin.Namespace
	}

	if err := r.ReconcileObject(r.SecretOp, r.SetDesiredSecretOp); err != nil {
		return err
	}
	r.NBClient.SetAuthToken(r.SecretOp.StringData["auth_token"])

	if r.JoinSecret == nil {
		if err := r.ReconcileObject(r.SecretAdmin, r.SetDesiredSecretAdminAccountInfo); err != nil {
			return err
		}
	}

	if err := r.ReconcileObject(r.SecretEndpoints, r.SetDesiredSecretEndpoints); err != nil {
		return err
	}
	return nil
}

// SetDesiredSecretAdmin set auth related info in admin secret
func (r *Reconciler) SetDesiredSecretAdmin() error {
	// Load string data from data
	util.SecretResetStringDataFromData(r.SecretAdmin)

	r.SecretAdmin.StringData["system"] = r.NooBaa.Name
	r.SecretAdmin.StringData["email"] = options.AdminAccountEmail
	if r.SecretAdmin.StringData["password"] == "" {
		r.SecretAdmin.StringData["password"] = util.RandomBase64(16)
	}
	return nil
}

// SetDesiredSecretOp set auth token in operator secret
func (r *Reconciler) SetDesiredSecretOp() error {
	// Load string data from data
	util.SecretResetStringDataFromData(r.SecretOp)

	// SecretOp exists means the system already created and we can skip
	if r.SecretOp.StringData["auth_token"] != "" {
		return nil
	}

	if r.JoinSecret == nil {
		// Trying to create token for admin so we could use it to create
		// a token for the operator account
		res1, err := r.NBClient.CreateAuthAPI(nb.CreateAuthParams{
			System:   r.Request.Name,
			Role:     "admin",
			Email:    r.SecretAdmin.StringData["email"],
			Password: r.SecretAdmin.StringData["password"],
		})
		if err == nil {
			r.NBClient.SetAuthToken(res1.Token)
			res2, err := r.NBClient.CreateAuthAPI(nb.CreateAuthParams{
				System: r.Request.Name,
				Role:   "operator",
				Email:  options.OperatorAccountEmail,
			})
			if err != nil {
				return fmt.Errorf("cannot create an auth token for operator, error: %v", err)
			}
			r.SecretOp.StringData["auth_token"] = res2.Token

		} else {
			// A failure to create a token for admin usually means that a system need to be created
			res3, err := r.NBClient.CreateSystemAPI(nb.CreateSystemParams{
				Name:     r.Request.Name,
				Email:    r.SecretAdmin.StringData["email"],
				Password: r.SecretAdmin.StringData["password"],
			})
			if err != nil {
				return fmt.Errorf("system creation failed, error: %v", err)
			}
			r.SecretOp.StringData["auth_token"] = res3.OperatorToken
		}
	} else {
		// Set the operator secret from the join secret
		r.SecretOp.StringData["auth_token"] = r.JoinSecret.StringData["auth_token"]
	}

	return nil
}

// SetDesiredSecretAdminAccountInfo set account related info in admin secret
func (r *Reconciler) SetDesiredSecretAdminAccountInfo() error {
	util.SecretResetStringDataFromData(r.SecretAdmin)

	account, err := r.NBClient.ReadAccountAPI(nb.ReadAccountParams{
		Email: r.SecretAdmin.StringData["email"],
	})
	if err != nil {
		return fmt.Errorf("cannot read admin account info, error: %v", err)
	}
	if account.AccessKeys == nil || len(account.AccessKeys) <= 0 {
		return fmt.Errorf("admin account has no access keys yet")
	}

	r.SecretAdmin.StringData["AWS_ACCESS_KEY_ID"] = account.AccessKeys[0].AccessKey
	r.SecretAdmin.StringData["AWS_SECRET_ACCESS_KEY"] = account.AccessKeys[0].SecretKey
	return nil
}

// SetDesiredSecretEndpoints set auth related info in endpoints secret
func (r *Reconciler) SetDesiredSecretEndpoints() error {
	if r.SecretEndpoints.UID != "" {
		return nil
	}

	// Load string data from data
	util.SecretResetStringDataFromData(r.SecretEndpoints)

	res, err := r.NBClient.CreateAuthAPI(nb.CreateAuthParams{
		System: r.Request.Name,
		Role:   "admin",
		Email:  options.AdminAccountEmail,
	})
	if err != nil {
		return fmt.Errorf("cannot create auth token for use by endpoints, error: %v", err)
	}

	r.SecretEndpoints.StringData["auth_token"] = res.Token
	return nil
}

// SetDesiredDeploymentEndpoint updates the endpoint deployment as desired for reconciling
func (r *Reconciler) SetDesiredDeploymentEndpoint() error {
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

			mgmtBaseAddr := ""
			s3BaseAddr := ""
			if r.JoinSecret == nil {
				mgmtBaseAddr = fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceMgmt.Name, r.Request.Namespace)
				s3BaseAddr = fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceS3.Name, r.Request.Namespace)
			}

			for j := range c.Env {
				switch c.Env[j].Name {
				case "MGMT_ADDR":
					if r.JoinSecret == nil {
						port := nb.FindPortByName(r.ServiceMgmt, "mgmt-https")
						c.Env[j].Value = fmt.Sprintf(`%s:%d`, mgmtBaseAddr, port.Port)
					} else {
						c.Env[j].Value = r.JoinSecret.StringData["mgmt_addr"]
					}
				case "BG_ADDR":
					if r.JoinSecret == nil {
						port := nb.FindPortByName(r.ServiceMgmt, "bg-https")
						c.Env[j].Value = fmt.Sprintf(`%s:%d`, mgmtBaseAddr, port.Port)
					} else {
						c.Env[j].Value = r.JoinSecret.StringData["bg_addr"]
					}
				case "MD_ADDR":
					if r.JoinSecret == nil {
						port := nb.FindPortByName(r.ServiceS3, "md-https")
						c.Env[j].Value = fmt.Sprintf(`%s:%d`, s3BaseAddr, port.Port)
					} else {
						c.Env[j].Value = r.JoinSecret.StringData["md_addr"]
					}
				case "HOSTED_AGENTS_ADDR":
					if r.JoinSecret == nil {
						port := nb.FindPortByName(r.ServiceMgmt, "hosted-agents-https")
						c.Env[j].Value = fmt.Sprintf(`%s:%d`, mgmtBaseAddr, port.Port)
					} else {
						c.Env[j].Value = r.JoinSecret.StringData["hosted_agents_addr"]
					}
				case "MONGODB_URL":
					if r.JoinSecret == nil {
						c.Env[j].Value = fmt.Sprintf(`mongodb://%s-0.%s/nbcore`,
							r.NooBaaDB.Name, r.NooBaaDB.Spec.ServiceName)
					}
				case "LOCAL_MD_SERVER":
					if r.JoinSecret == nil {
						c.Env[j].Value = "true"
					}
				case "LOCAL_N2N_AGENT":
					if r.JoinSecret == nil {
						c.Env[j].Value = "true"
					}
				case "JWT_SECRET":
					if r.JoinSecret == nil {
						c.Env[j].ValueFrom = &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "noobaa-server",
								},
								Key: "jwt",
							},
						}
					}
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
	return nil
}

// ReconcileHPAEndpoint reconcile the endpoint's HPS and report the configuration
// back to the noobaa core
func (r *Reconciler) ReconcileHPAEndpoint() error {
	if err := r.ReconcileObject(r.HPAEndpoint, r.SetDesiredHPAEndpoint); err != nil {
		return err
	}

	max := r.HPAEndpoint.Spec.MaxReplicas
	min := r.HPAEndpoint.Spec.MaxReplicas
	if r.HPAEndpoint.Spec.MinReplicas != nil {
		min = *r.HPAEndpoint.Spec.MinReplicas
	}

	return r.NBClient.UpdateEndpointGroupAPI(nb.UpdateEndpointGroupParams{
		GroupName: fmt.Sprint(r.NooBaa.UID),
		IsRemote:  r.JoinSecret != nil,
		EndpointRange: nb.IntRange{
			Min: min,
			Max: max,
		},
	})
}

// SetDesiredHPAEndpoint updates the endpoint horizontal pod autoscaler as desired for reconciling
func (r *Reconciler) SetDesiredHPAEndpoint() error {
	endpointsSpec := r.NooBaa.Spec.Endpoints
	if endpointsSpec != nil {
		r.HPAEndpoint.Spec.MinReplicas = &endpointsSpec.MinCount
		r.HPAEndpoint.Spec.MaxReplicas = endpointsSpec.MaxCount
	}
	return nil
}

// RegisterToCluster registers the noobaa client with the noobaa cluster
func (r *Reconciler) RegisterToCluster() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

	return r.NBClient.RegisterToCluster()
}

// ReconcileDefaultBackingStore attempts to get credentials to cloud storage using the cloud-credentials operator
// and use it for the default backing store
func (r *Reconciler) ReconcileDefaultBackingStore() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

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
		suffix := ""
		hostItems := strings.Split(r.RouteMgmt.Spec.Host, ".")
		for i := len(hostItems) - 1; i >= 0; i-- {
			hostItem := strings.Trim(hostItems[i], "-.")
			if len(name)+1+len(hostItem)+1+len(suffix) > 63 {
				break
			}
			suffix = hostItem + "." + suffix
		}
		name += "." + suffix
	}
	// make sure the name is ended with a valid charecter
	name = strings.Trim(name, "-.")

	// if for some reason the bucket name is not valid then fallback to nb.timestamp
	if !util.IsValidS3BucketName(name) {
		oldName := name
		name = "nb." + tsMilli
		logrus.Warnf("generated bucket name (%s) is invalid. falling back to (%s)", oldName, name)
	}
	return name
}

// ReconcileDefaultBucketClass creates the default bucket class
func (r *Reconciler) ReconcileDefaultBucketClass() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

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
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

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

// ReconcilePrometheusRule reconciles prometheus rule
func (r *Reconciler) ReconcilePrometheusRule() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

	return r.ReconcileObjectOptional(r.PrometheusRule, nil)
}

// ReconcileServiceMonitor reconciles service monitor
func (r *Reconciler) ReconcileServiceMonitor() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

	return r.ReconcileObjectOptional(r.ServiceMonitor, nil)
}

// ReconcileReadSystem calls read_system on noobaa server and stores the result
func (r *Reconciler) ReconcileReadSystem() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

	// update noobaa-core version in reconciler struct
	systemInfo, err := r.NBClient.ReadSystemAPI()
	if err != nil {
		r.Logger.Errorf("failed to read system info: %v", err)
		return err
	}
	r.SystemInfo = &systemInfo
	r.Logger.Infof("updating noobaa-core version to %s", systemInfo.Version)
	r.CoreVersion = systemInfo.Version

	// update backingstores and bucketclass mode
	r.UpdateBackingStoresPhase(systemInfo.Pools)
	r.UpdateBucketClassesPhase(systemInfo.Buckets)

	return nil
}

// UpdateBackingStoresPhase updates newPhase of backingstore after readSystem
func (r *Reconciler) UpdateBackingStoresPhase(pools []nb.PoolInfo) {

	bsList := &nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	}
	if !util.KubeList(bsList, &client.ListOptions{Namespace: options.Namespace}) {
		logrus.Errorf("not found: Backing Store list")
	}
	for i := range bsList.Items {
		bs := &bsList.Items[i]
		for _, pool := range pools {
			if pool.Name == bs.Name && bs.Status.Mode.ModeCode != pool.Mode {
				bs.Status.Mode.ModeCode = pool.Mode
				bs.Status.Mode.TimeStamp = fmt.Sprint(time.Now())
				r.NooBaa.Status.ObservedGeneration = r.NooBaa.Generation
				r.Client.Status().Update(r.Ctx, bs)
			}
		}
	}
}

// UpdateBucketClassesPhase updates newPhase of bucketclass after readSystem
func (r *Reconciler) UpdateBucketClassesPhase(Buckets []nb.BucketInfo) {

	bucketclassList := &nbv1.BucketClassList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClassList"},
	}
	if !util.KubeList(bucketclassList, &client.ListOptions{Namespace: options.Namespace}) {
		logrus.Errorf("not found: Backing Store list")
	}
	for i := range bucketclassList.Items {
		bc := &bucketclassList.Items[i]
		for _, bucket := range Buckets {

			bucketTieringPolicyName := ""
			if bucket.BucketClaim != nil {
				bucketTieringPolicyName = bucket.BucketClaim.BucketClass
			}
			if bc.Name == bucketTieringPolicyName && bucket.Tiering.Mode != bc.Status.Mode {
				bc.Status.Mode = bucket.Tiering.Mode
				r.NooBaa.Status.ObservedGeneration = r.NooBaa.Generation
				r.Client.Status().Update(r.Ctx, bc)

			}
		}
	}
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
