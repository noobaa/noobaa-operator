package system

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/marstr/randname"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	ibmEndpoint = "https://s3.direct.%s.cloud-object-storage.appdomain.cloud"
	ibmLocation = "%s-standard"
	ibmCOSCred  = "ibm-cloud-cos-creds"
)

type gcpAuthJSON struct {
	ProjectID string `json:"project_id"`
}

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
	if err := r.ReconcileServiceMonitors(); err != nil {
		return err
	}
	if err := r.ReconcileReadSystem(); err != nil {
		return err
	}
	if err := r.ReconcileDeploymentEndpointStatus(); err != nil {
		return err
	}
	if err := r.ReconcileSystemConfigMap(); err != nil {
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

	// Local noobaa case
	if r.JoinSecret == nil {
		res1, err := r.NBClient.ReadSystemStatusAPI()
		if err != nil {
			return fmt.Errorf("Could not read the system status, error: %v", err)
		}

		if res1.State == "DOES_NOT_EXIST" {
			res2, err := r.NBClient.CreateSystemAPI(nb.CreateSystemParams{
				Name:     r.Request.Name,
				Email:    r.SecretAdmin.StringData["email"],
				Password: r.SecretAdmin.StringData["password"],
			})
			if err != nil {
				return fmt.Errorf("system creation failed, error: %v", err)
			}

			r.SecretOp.StringData["auth_token"] = res2.OperatorToken

		} else if res1.State == "COULD_NOT_INITIALIZE" {
			// TODO: Try to recover from this situation, maybe delete the system.
			return util.NewPersistentError("SystemCouldNotInitialize",
				"Something went wrong during system initialization")

		} else if res1.State == "READY" {
			// Trying to create token for admin so we could use it to create
			// a token for the operator account
			res3, err := r.NBClient.CreateAuthAPI(nb.CreateAuthParams{
				System:   r.Request.Name,
				Role:     "admin",
				Email:    r.SecretAdmin.StringData["email"],
				Password: r.SecretAdmin.StringData["password"],
			})
			if err != nil {
				return fmt.Errorf("cannot create an auth token for admin, error: %v", err)
			}

			r.NBClient.SetAuthToken(res3.Token)
			res4, err := r.NBClient.CreateAuthAPI(nb.CreateAuthParams{
				System: r.Request.Name,
				Role:   "operator",
				Email:  options.OperatorAccountEmail,
			})
			if err != nil {
				return fmt.Errorf("cannot create an auth token for operator, error: %v", err)
			}
			r.SecretOp.StringData["auth_token"] = res4.Token
		}

		// Remote noobaa case
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
	r.DeploymentEndpoint.Spec.Template.Labels["app"] = r.Request.Name

	endpointsSpec := r.NooBaa.Spec.Endpoints
	podSpec := &r.DeploymentEndpoint.Spec.Template.Spec
	if r.NooBaa.Spec.Tolerations != nil {
		podSpec.Tolerations = r.NooBaa.Spec.Tolerations
	}
	if r.NooBaa.Spec.Affinity != nil {
		podSpec.Affinity = r.NooBaa.Spec.Affinity
	}
	if r.NooBaa.Spec.ImagePullSecret == nil {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{}
	} else {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{*r.NooBaa.Spec.ImagePullSecret}
	}
	rootUIDGid := int64(0)
	podSpec.SecurityContext.RunAsUser = &rootUIDGid
	podSpec.SecurityContext.RunAsGroup = &rootUIDGid
	podSpec.ServiceAccountName = "noobaa-endpoint"

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
			util.MergeEnvArrays(&c.Env, &r.DefaultDeploymentEndpoint.Containers[0].Env)
			if r.JoinSecret == nil {
				mgmtBaseAddr = fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceMgmt.Name, r.Request.Namespace)
				s3BaseAddr = fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceS3.Name, r.Request.Namespace)
				r.setDesiredCoreEnv(c)
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
						c.Env[j].Value = r.MongoConnectionString
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
				case "NOOBAA_ROOT_SECRET":
					c.Env[j].Value = r.SecretRootMasterKey
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
					c.Env[j].Value = fmt.Sprint(strings.Join(hosts[:], " "))
				case "ENDPOINT_GROUP_ID":
					c.Env[j].Value = fmt.Sprint(r.NooBaa.UID)

				case "REGION":
					if r.NooBaa.Spec.Region != nil {
						c.Env[j].Value = *r.NooBaa.Spec.Region
					} else {
						c.Env[j].Value = ""
					}

				case "NODE_EXTRA_CA_CERTS":
					c.Env[j].Value = r.ApplyCAsToPods
				}
			}

			c.SecurityContext = &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"SETUID", "SETGID"},
				},
			}

			util.ReflectEnvVariable(&c.Env, "HTTP_PROXY")
			util.ReflectEnvVariable(&c.Env, "HTTPS_PROXY")
			util.ReflectEnvVariable(&c.Env, "NO_PROXY")

			if r.DeploymentEndpoint.ObjectMeta.Annotations == nil {
				r.DeploymentEndpoint.ObjectMeta.Annotations = make(map[string]string)
			}

			r.SetConfigMapAnnotation(r.DeploymentEndpoint.ObjectMeta.Annotations)

			return r.setDesiredEndpointMounts(podSpec, c)
		}
	}
	return nil
}

func (r *Reconciler) setDesiredEndpointMounts(podSpec *corev1.PodSpec, container *corev1.Container) error {

	namespaceStoreList := &nbv1.NamespaceStoreList{}
	if !util.KubeList(namespaceStoreList, client.InNamespace(options.Namespace)) {
		return fmt.Errorf("Error: Cant list namespacestores")
	}
	podSpec.Volumes = r.DefaultDeploymentEndpoint.Volumes
	container.VolumeMounts = r.DefaultDeploymentEndpoint.Containers[0].VolumeMounts

	for _, nsStore := range namespaceStoreList.Items {
		// Since namespacestore is able to get a rejected state on runtime errors,
		// we want to skip namespacestores with invalid configuration only.
		// Remove this validation when the kubernetes validations hooks will be available.
		if !r.validateNsStoreNSFS(&nsStore) {
			continue
		}
		if nsStore.Spec.NSFS != nil {
			pvcName := nsStore.Spec.NSFS.PvcName
			isPvcExist := false
			volumeName := "nsfs-" + nsStore.Name
			for _, volume := range podSpec.Volumes {
				if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName {
					isPvcExist = true // PVC already attached to the pods - no need to add
					volumeName = volume.Name
					break
				}
			}
			if !isPvcExist {
				podSpec.Volumes = append(podSpec.Volumes, corev1.Volume{
					Name: volumeName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				})
			}
			subPath := nsStore.Spec.NSFS.SubPath
			mountPath := "/nsfs/" + nsStore.Name
			isMountExist := false
			for _, volumeMount := range container.VolumeMounts {
				if volumeMount.Name == volumeName && volumeMount.SubPath == subPath {
					isMountExist = true // volumeMount already created - no need to add
					break
				}
			}
			if !isMountExist {
				container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
					Name:      volumeName,
					MountPath: mountPath,
					SubPath:   subPath,
				})
			}
		}
	}
	return nil
}

// Duplicate code from validation.go namespacetore pkg.
// Cannot import the namespacestore pkg, because the pkg imports the system pkg
// TODO remove the code
func (r *Reconciler) validateNsStoreNSFS(nsStore *nbv1.NamespaceStore) bool {
	nsfs := nsStore.Spec.NSFS

	if nsfs == nil {
		return true
	}

	//pvcName validation
	if nsfs.PvcName == "" {
		return false
	}

	//Check the mountPath
	mountPath := "/nsfs/" + nsStore.Name
	if len(mountPath) > 63 {
		return false
	}

	//SubPath validation
	if nsfs.SubPath != "" {
		path := nsfs.SubPath
		if len(path) > 0 && path[0] == '/' {
			return false
		}
		parts := strings.Split(path, "/")
		for _, item := range parts {
			if item == ".." {
				return false
			}
		}
	}
	return true
}

// awaitEndpointDeploymentPods wait for the the endpoint deployment to become ready
// before creating the controlling HPA
// See https://bugzilla.redhat.com/show_bug.cgi?id=1885524
func (r *Reconciler) awaitEndpointDeploymentPods() error {

	// Check that all deployment pods are available
	availablePods := r.DeploymentEndpoint.Status.AvailableReplicas
	desiredPods := r.DeploymentEndpoint.Status.Replicas
	if availablePods == 0 || availablePods != desiredPods {
		return errors.New("not enough available replicas in endpoint deployment")
	}

	// Check that deployment is ready
	for _, condition := range r.DeploymentEndpoint.Status.Conditions {
		if condition.Status != "True" {
			return errors.New("endpoint deployment is not ready")
		}
	}

	return nil
}

// ReconcileHPAEndpoint reconcile the endpoint's HPA and report the configuration
// back to the noobaa core
func (r *Reconciler) ReconcileHPAEndpoint() error {
	// Wait for the the endpoint deployment to become ready
	// only if HPA was not created yet
	if r.HPAEndpoint.UID == "" {
		if err := r.awaitEndpointDeploymentPods(); err != nil {
			return err
		}
	}

	if err := r.ReconcileObject(r.HPAEndpoint, r.SetDesiredHPAEndpoint); err != nil {
		return err
	}

	max := r.HPAEndpoint.Spec.MaxReplicas
	min := r.HPAEndpoint.Spec.MaxReplicas
	if r.HPAEndpoint.Spec.MinReplicas != nil {
		min = *r.HPAEndpoint.Spec.MinReplicas
	}

	region := ""
	if r.NooBaa.Spec.Region != nil {
		region = *r.NooBaa.Spec.Region
	}

	return r.NBClient.UpdateEndpointGroupAPI(nb.UpdateEndpointGroupParams{
		GroupName: fmt.Sprint(r.NooBaa.UID),
		IsRemote:  r.JoinSecret != nil,
		Region:    region,
		EndpointRange: nb.IntRange{
			Min: min,
			Max: max,
		},
	})
}

// SetDesiredHPAEndpoint updates the endpoint horizontal pod autoscaler as desired for reconciling
func (r *Reconciler) SetDesiredHPAEndpoint() error {
	var minReplicas int32 = 1
	var maxReplicas int32 = 2

	endpointsSpec := r.NooBaa.Spec.Endpoints
	if endpointsSpec != nil {
		minReplicas = endpointsSpec.MinCount
		maxReplicas = endpointsSpec.MaxCount
	}

	r.HPAEndpoint.Spec.MinReplicas = &minReplicas
	r.HPAEndpoint.Spec.MaxReplicas = maxReplicas

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

	if r.CephObjectStoreUser.UID != "" {
		log.Infof("CephObjectStoreUser %q created. Creating default backing store on ceph objectstore", r.CephObjectStoreUser.Name)
		if err := r.prepareCephBackingStore(); err != nil {
			return err
		}
	} else if r.AWSCloudCreds.UID != "" {
		log.Infof("CredentialsRequest %q created. Creating default backing store on AWS objectstore", r.AWSCloudCreds.Name)
		if err := r.prepareAWSBackingStore(); err != nil {
			return err
		}
	} else if r.AzureCloudCreds.UID != "" {
		log.Infof("CredentialsRequest %q created. Creating default backing store on Azure objectstore", r.AzureCloudCreds.Name)
		if err := r.prepareAzureBackingStore(); err != nil {
			return err
		}
	} else if r.GCPCloudCreds.UID != "" {
		log.Infof("CredentialsRequest %q created.  creating default backing store on GCP objectstore", r.GCPCloudCreds.Name)
		if err := r.prepareGCPBackingStore(); err != nil {
			return err
		}
	} else if r.IBMCloudCOSCreds.UID != "" {
		log.Infof("IBM objectstore credentials %q created. Creating default backing store on IBM objectstore", r.IBMCloudCOSCreds.Name)
		if err := r.prepareIBMBackingStore(); err != nil {
			return err
		}
	} else {
		minutesSinceCreation := time.Since(r.NooBaa.CreationTimestamp.Time).Minutes()
		if minutesSinceCreation < 2 {
			return nil
		}
		if err := r.preparePVPoolBackingStore(); err != nil {
			return err
		}
	}

	r.Own(r.DefaultBackingStore)
	if err := r.Client.Create(r.Ctx, r.DefaultBackingStore); err != nil {
		log.Errorf("got error on DefaultBackingStore creation. error: %v", err)
		return err
	}
	return nil
}

func (r *Reconciler) preparePVPoolBackingStore() error {

	// create backing store
	defaultPVSize := int64(50) * 1024 * 1024 * 1024 // 50GB
	r.DefaultBackingStore.Spec.Type = nbv1.StoreTypePVPool
	r.DefaultBackingStore.Spec.PVPool = &nbv1.PVPoolSpec{}
	r.DefaultBackingStore.Spec.PVPool.NumVolumes = 1
	r.DefaultBackingStore.Spec.PVPool.VolumeResources = &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceStorage: *resource.NewQuantity(defaultPVSize, resource.BinarySI),
		},
	}
	if r.NooBaa.Spec.PVPoolDefaultStorageClass != nil {
		r.DefaultBackingStore.Spec.PVPool.StorageClass = *r.NooBaa.Spec.PVPoolDefaultStorageClass
	} else {
		storageClassName, err := r.findLocalStorageClass()
		if err != nil {
			r.Logger.Errorf("got error finding a default/local storage class. error: %v", err)
			return err
		}
		r.DefaultBackingStore.Spec.PVPool.StorageClass = storageClassName
	}
	return nil
}

func (r *Reconciler) prepareAWSBackingStore() error {
	// after we have cloud credential request, wait for credentials secret
	cloudCredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.AWSCloudCreds.Spec.SecretRef.Name,
			Namespace: r.AWSCloudCreds.Spec.SecretRef.Namespace,
		},
	}

	util.KubeCheck(cloudCredsSecret)
	if cloudCredsSecret.UID == "" {
		// TODO: we need to figure out why secret is not created, and react accordingly
		// e.g. maybe we are running on azure but our CredentialsRequest is for AWS
		r.Logger.Infof("Secret %q was not created yet by cloud-credentials operator. retry on next reconcile..", r.AWSCloudCreds.Spec.SecretRef.Name)
		return fmt.Errorf("cloud credentials secret %q is not ready yet", r.AWSCloudCreds.Spec.SecretRef.Name)
	}
	r.Logger.Infof("Secret %s was created successfully by cloud-credentials operator", r.AWSCloudCreds.Spec.SecretRef.Name)

	// create the actual S3 bucket
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

func (r *Reconciler) prepareAzureBackingStore() error {
	// after we have cloud credential request, wait for credentials secret
	cloudCredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.AzureCloudCreds.Spec.SecretRef.Name,
			Namespace: r.AzureCloudCreds.Spec.SecretRef.Namespace,
		},
	}

	util.KubeCheck(cloudCredsSecret)
	if cloudCredsSecret.UID == "" {
		// TODO: we need to figure out why secret is not created, and react accordingly
		// e.g. maybe we are running on AWS but our CredentialsRequest is for Azure
		r.Logger.Infof("Secret %q was not created yet by cloud-credentials operator. retry on next reconcile..", r.AzureCloudCreds.Spec.SecretRef.Name)
		return fmt.Errorf("cloud credentials secret %q is not ready yet", r.AzureCloudCreds.Spec.SecretRef.Name)
	}
	r.Logger.Infof("Secret %s was created successfully by cloud-credentials operator", r.AzureCloudCreds.Spec.SecretRef.Name)

	util.KubeCheck(r.AzureContainerCreds)
	if r.AzureContainerCreds.UID == "" {
		// AzureContainerCreds does not exist. create one
		r.Logger.Info("Creating AzureContainerCreds secret")
		r.AzureContainerCreds.StringData = map[string]string{}
		r.AzureContainerCreds.StringData = cloudCredsSecret.StringData
		r.Own(r.AzureContainerCreds)
		if err := r.Client.Create(r.Ctx, r.AzureContainerCreds); err != nil {
			return fmt.Errorf("got error on AzureContainerCreds creation. error: %v", err)
		}
	}
	r.Logger.Infof("Secret %s was created successfully", r.AzureContainerCreds.Name)

	var azureGroupName = r.AzureContainerCreds.StringData["azure_resourcegroup"]

	if r.AzureContainerCreds.StringData["AccountName"] == "" {
		var azureAccountName = strings.ToLower(randname.GenerateWithPrefix("noobaaaccount", 5))
		_, err := r.CreateStorageAccount(azureAccountName, azureGroupName)
		if err != nil {
			return err
		}
		r.AzureContainerCreds.StringData["AccountName"] = azureAccountName
	}

	if r.AzureContainerCreds.StringData["AccountKey"] == "" {
		var azureAccountName = r.AzureContainerCreds.StringData["AccountName"]
		key := r.getAccountPrimaryKey(azureAccountName, azureGroupName)
		r.AzureContainerCreds.StringData["AccountKey"] = key
	}

	azureContainerName := ""
	if r.AzureContainerCreds.StringData["targetBlobContainer"] == "" {
		azureContainerName = strings.ToLower(randname.GenerateWithPrefix("noobaacontainer", 5))
		_, err := r.CreateContainer(r.AzureContainerCreds.StringData["AccountName"], azureGroupName, azureContainerName)
		if err != nil {
			return err
		}
		r.AzureContainerCreds.StringData["targetBlobContainer"] = azureContainerName
	}

	if errUpdate := r.Client.Update(r.Ctx, r.AzureContainerCreds); errUpdate != nil {
		return fmt.Errorf("got error on AzureContainerCreds update. error: %v", errUpdate)
	}

	// create backing store
	r.DefaultBackingStore.Spec.Type = nbv1.StoreTypeAzureBlob
	r.DefaultBackingStore.Spec.AzureBlob = &nbv1.AzureBlobSpec{
		TargetBlobContainer: azureContainerName,
		Secret: corev1.SecretReference{
			Name:      r.AzureContainerCreds.Name,
			Namespace: r.AzureContainerCreds.Namespace,
		},
	}

	return nil
}

func (r *Reconciler) prepareGCPBackingStore() error {

	cloudCredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.GCPCloudCreds.Spec.SecretRef.Name,
			Namespace: r.GCPCloudCreds.Spec.SecretRef.Namespace,
		},
	}

	util.KubeCheck(cloudCredsSecret)
	if cloudCredsSecret.UID == "" {
		// TODO: we need to figure out why secret is not created, and react accordingly
		// e.g. maybe we are running on AWS but our CredentialsRequest is for GCP
		r.Logger.Infof("Secret %q was not created yet by cloud-credentials operator. retry on next reconcile..", r.GCPCloudCreds.Spec.SecretRef.Name)
		return fmt.Errorf("cloud credentials secret %q is not ready yet", r.GCPCloudCreds.Spec.SecretRef.Name)
	}
	r.Logger.Infof("Secret %s was created successfully by cloud-credentials operator", r.GCPCloudCreds.Spec.SecretRef.Name)

	util.KubeCheck(r.GCPBucketCreds)
	if r.GCPBucketCreds.UID == "" {
		r.GCPBucketCreds.StringData = cloudCredsSecret.StringData
		r.Own(r.GCPBucketCreds)
		if err := r.Client.Create(r.Ctx, r.GCPBucketCreds); err != nil {
			return fmt.Errorf("got error on GCPBucketCreds creation. error: %v", err)
		}
	}
	authJSON := &gcpAuthJSON{}
	err := json.Unmarshal([]byte(cloudCredsSecret.StringData["service_account.json"]), authJSON)
	if err != nil {
		fmt.Println("Failed to parse secret", err)
		return err
	}
	projectID := authJSON.ProjectID
	r.GCPBucketCreds.StringData["GoogleServiceAccountPrivateKeyJson"] = cloudCredsSecret.StringData["service_account.json"]
	ctx := context.Background()
	gcpclient, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(cloudCredsSecret.StringData["service_account.json"])))
	if err != nil {
		r.Logger.Info(err)
		return err
	}

	var bucketName = strings.ToLower(randname.GenerateWithPrefix("noobaabucket", 5))
	if err := r.createGCPBucketForBackingStore(gcpclient, projectID, bucketName); err != nil {
		r.Logger.Info(err)
		return err
	}

	if errUpdate := r.Client.Update(r.Ctx, r.GCPBucketCreds); errUpdate != nil {
		return fmt.Errorf("got error on GCPBucketCreds update. error: %v", errUpdate)
	}
	// create backing store
	r.DefaultBackingStore.Spec.Type = nbv1.StoreTypeGoogleCloudStorage
	r.DefaultBackingStore.Spec.GoogleCloudStorage = &nbv1.GoogleCloudStorageSpec{
		TargetBucket: bucketName,
		Secret: corev1.SecretReference{
			Name:      r.GCPBucketCreds.Name,
			Namespace: r.GCPBucketCreds.Namespace,
		},
	}
	return nil
}

func (r *Reconciler) prepareIBMBackingStore() error {
	r.Logger.Info("Preparing backing store in IBM Cloud")

	var (
		endpoint string
		location string
	)

	util.KubeCheck(r.IBMCloudCOSCreds)
	if r.IBMCloudCOSCreds.UID == "" {
		r.Logger.Errorf("Cloud credentials secret %q is not ready yet", r.IBMCloudCOSCreds.Name)
		return fmt.Errorf("Cloud credentials secret %q is not ready yet", r.IBMCloudCOSCreds.Name)
	}

	if val, ok := r.IBMCloudCOSCreds.StringData["IBM_COS_Endpoint"]; ok {
		// Use the endpoint provided in the secret
		endpoint = val
		r.Logger.Infof("Endpoint provided in secret: %q", endpoint)
		if val, ok := r.IBMCloudCOSCreds.StringData["IBM_COS_Location"]; ok {
			location = val
			r.Logger.Infof("Location provided in secret: %q", location)
		}
	} else {
		// Endpoint not provided in the secret, construct one based on the cluster's region
		// https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-endpoints#endpoints
		region, err := util.GetIBMRegion()
		if err != nil {
			r.Logger.Errorf("Failed to get IBM Region. %q", err)
			return fmt.Errorf("Failed to get IBM Region")
		}
		r.Logger.Infof("Constructing endpoint for region: %q", region)
		// https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-classes#classes-locationconstraint
		endpoint = fmt.Sprintf(ibmEndpoint, region)
		location = fmt.Sprintf(ibmLocation, region)
	}

	if _, err := url.Parse(endpoint); err != nil {
		r.Logger.Errorf("Invalid formate URL %q", endpoint)
		return fmt.Errorf("Invalid formate URL %q", endpoint)
	}

	r.Logger.Infof("IBM COS Endpoint: %s   LocationConstraint: %s", endpoint, location)

	var accessKeyID string
	if val, ok := r.IBMCloudCOSCreds.StringData["IBM_COS_ACCESS_KEY_ID"]; ok {
		accessKeyID = val
	} else {
		r.Logger.Errorf("Missing IBM_COS_ACCESS_KEY_ID in the secret")
		return fmt.Errorf("Missing IBM_COS_ACCESS_KEY_ID in the secret")
	}

	var secretAccessKey string
	if val, ok := r.IBMCloudCOSCreds.StringData["IBM_COS_SECRET_ACCESS_KEY"]; ok {
		secretAccessKey = val
	} else {
		r.Logger.Errorf("Missing IBM_COS_SECRET_ACCESS_KEY in the secret")
		return fmt.Errorf("Missing IBM_COS_SECRET_ACCESS_KEY in the secret")
	}

	bucketName := r.generateBackingStoreTargetName()
	r.Logger.Infof("IBM COS Bucket Name: %s", bucketName)

	s3Config := &aws.Config{
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(
			accessKeyID,
			secretAccessKey,
			"",
		),
		Region: &location,
	}
	if err := r.createS3BucketForBackingStore(s3Config, bucketName); err != nil {
		return err
	}
	r.Logger.Infof("Created bucket: %s", bucketName)

	// create backing store
	r.DefaultBackingStore.Spec.Type = nbv1.StoreTypeIBMCos
	r.DefaultBackingStore.Spec.IBMCos = &nbv1.IBMCosSpec{
		TargetBucket: bucketName,
		Secret: corev1.SecretReference{
			Name:      r.IBMCloudCOSCreds.Name,
			Namespace: r.IBMCloudCOSCreds.Namespace,
		},
		Endpoint:         endpoint,
		SignatureVersion: nbv1.S3SignatureVersionV2,
	}
	return nil
}

func (r *Reconciler) createGCPBucketForBackingStore(client *storage.Client, projectID, bucketName string) error {
	// [START create_bucket]
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	if err := client.Bucket(bucketName).Create(ctx, projectID, nil); err != nil {
		return err
	}
	// [END create_bucket]
	return nil
}

func (r *Reconciler) prepareCephBackingStore() error {
	util.KubeCheck(r.CephObjectStoreUser)
	if r.CephObjectStoreUser.UID == "" || r.CephObjectStoreUser.Status.Phase != "Ready" {
		r.Logger.Infof("Ceph objectstore user %q is not ready. retry on next reconcile..", r.CephObjectStoreUser.Name)
		return fmt.Errorf("Ceph objectstore user %q is not ready", r.CephObjectStoreUser.Name)
	}

	secretName := r.CephObjectStoreUser.Status.Info["secretName"]
	if secretName == "" {
		return util.NewPersistentError("InvalidCephObjectStoreUser",
			"Ceph objectstore user is ready but a secret name was not provided")
	}

	// get access\secret keys from user secret
	cephObjectStoreUserSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: options.Namespace,
		},
	}
	util.KubeCheck(cephObjectStoreUserSecret)
	if cephObjectStoreUserSecret.UID == "" {
		r.Logger.Infof("Ceph objectstore user secret %q was not created yet. retry on next reconcile..", secretName)
		return fmt.Errorf("Ceph objectstore user secret %q is not ready yet", secretName)
	}

	endpoint := cephObjectStoreUserSecret.StringData["Endpoint"]
	r.Logger.Infof("Will connect to RGW at %q", endpoint)

	region := "us-east-1"
	forcePathStyle := true
	client := &http.Client{
		Transport: util.InsecureHTTPTransport,
		Timeout:   10 * time.Second,
	}
	if r.ApplyCAsToPods != "" {
		client.Transport = util.SecureHTTPTransport
	}

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			cephObjectStoreUserSecret.StringData["AccessKey"],
			cephObjectStoreUserSecret.StringData["SecretKey"],
			"",
		),
		Endpoint:         &endpoint,
		Region:           &region,
		S3ForcePathStyle: &forcePathStyle,
		HTTPClient:       client,
	}

	bucketName := r.generateBackingStoreTargetName()
	if err := r.createS3BucketForBackingStore(s3Config, bucketName); err != nil {
		return err
	}

	// create backing store
	if r.DefaultBackingStore.ObjectMeta.Annotations == nil {
		r.DefaultBackingStore.ObjectMeta.Annotations = map[string]string{}
	}
	r.DefaultBackingStore.ObjectMeta.Annotations["rgw"] = ""
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

	r.DefaultBucketClass.Spec.PlacementPolicy = &nbv1.PlacementPolicy{
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

// ApplyMonitoringLabels function adds the name of the resource that manages
// noobaa, as a label on the noobaa metrics
func (r *Reconciler) ApplyMonitoringLabels(serviceMonitor *monitoringv1.ServiceMonitor) {
	if r.NooBaa.Spec.Labels != nil {
		if monitoringLabels, ok := r.NooBaa.Spec.Labels["monitoring"]; ok {
			if managedBy, ok := monitoringLabels["noobaa.io/managedBy"]; ok {
				relabelConfig := monitoringv1.RelabelConfig{
					TargetLabel: "managedBy",
					Replacement: managedBy,
				}
				serviceMonitor.Spec.Endpoints[0].RelabelConfigs = append(
					serviceMonitor.Spec.Endpoints[0].RelabelConfigs, &relabelConfig)
			} else {
				r.Logger.Info("noobaa.io/managedBy not specified in monitoring labels")
			}
		} else {
			r.Logger.Info("monitoring labels not specified")
		}
	}
}

// ReconcileServiceMonitors reconciles service monitors
func (r *Reconciler) ReconcileServiceMonitors() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

	r.ApplyMonitoringLabels(r.ServiceMonitorMgmt)

	if err := r.ReconcileObjectOptional(r.ServiceMonitorMgmt, nil); err != nil {
		return err
	}
	if err := r.ReconcileObjectOptional(r.ServiceMonitorS3, nil); err != nil {
		return err
	}
	return nil
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

	// creates namespace stores if sync is needed on upgrade
	if len(systemInfo.NamespaceResources) > 0 {
		if err := r.ReconcileNamespaceStores(systemInfo.NamespaceResources); err != nil {
			r.Logger.Infof("got error on ReconcileNamespaceStores, %+v", err)
			return err
		}
	}

	// update backingstores, namespacestores and bucketclass mode
	r.UpdateBackingStoresPhase(systemInfo.Pools)
	r.UpdateNamespaceStoresPhase(systemInfo.NamespaceResources)
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
				err := r.Client.Status().Update(r.Ctx, bs)
				if err != nil {
					logrus.Errorf("got error when trying to update status of backingstore %v. %v", bs.Name, err)
				}
			}
		}
	}
}

// UpdateNamespaceStoresPhase updates newPhase of namespace resource after readSystem
func (r *Reconciler) UpdateNamespaceStoresPhase(namespaceResources []nb.NamespaceResourceInfo) {

	nssList := &nbv1.NamespaceStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStoreList"},
	}
	if !util.KubeList(nssList, &client.ListOptions{Namespace: options.Namespace}) {
		logrus.Errorf("not found: Namespace Store list")
	}
	for i := range nssList.Items {
		nss := &nssList.Items[i]
		for _, namespaceResource := range namespaceResources {
			if namespaceResource.Name == nss.Name && nss.Status.Mode.ModeCode != namespaceResource.Mode {
				nss.Status.Mode.ModeCode = namespaceResource.Mode
				nss.Status.Mode.TimeStamp = fmt.Sprint(time.Now())
				r.NooBaa.Status.ObservedGeneration = r.NooBaa.Generation
				err := r.Client.Status().Update(r.Ctx, nss)
				if err != nil {
					logrus.Errorf("got error when trying to update status of namespacestore %v. %v", nss.Name, err)
				}
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

			// in case of a namespace bucket, we might not have bucket.Tiering. skip
			if bucket.Tiering == nil {
				continue
			}

			bucketTieringPolicyName := ""
			if bucket.BucketClaim != nil {
				bucketTieringPolicyName = bucket.BucketClaim.BucketClass
			}
			if bc.Name == bucketTieringPolicyName && bucket.Tiering.Mode != bc.Status.Mode {
				bc.Status.Mode = bucket.Tiering.Mode
				r.NooBaa.Status.ObservedGeneration = r.NooBaa.Generation
				err := r.Client.Status().Update(r.Ctx, bc)
				if err != nil {
					logrus.Errorf("got error when trying to update status of bucket class %v. %v ", bc.Name, err)
				}

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

// ReconcileNamespaceStores syncs between core namespace resources with namespace bucketclasses
func (r *Reconciler) ReconcileNamespaceStores(namespaceResources []nb.NamespaceResourceInfo) error {
	r.Logger.Infof("ReconcileNamespaceStores: %+v", namespaceResources)

	for _, nsr := range namespaceResources {
		r.Logger.Infof("ReconcileNamespaceStores: nsr: %+v", nsr)
		nsStore := &nbv1.NamespaceStore{
			TypeMeta: metav1.TypeMeta{Kind: "NamespaceStore"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      nsr.Name,
				Namespace: options.Namespace,
			},
		}
		if !util.KubeCheck(nsStore) {
			namespaceResourceOperatorInfo, err := r.NBClient.ReadNamespaceResourceOperatorInfoAPI(nb.ReadNamespaceResourceParams{Name: nsr.Name})
			if err != nil {
				logrus.Warnf(`❌ Failed to read NamespaceStore secrets: %s`, err)
				continue
			}
			if !namespaceResourceOperatorInfo.NeedK8sSync {
				continue
			}

			o := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
			secret := o.(*corev1.Secret)
			secret.Namespace = options.Namespace
			secret.Data = nil
			secret.StringData = map[string]string{}

			switch nsr.EndpointType {
			case "AWS":
				nsStore.Spec.Type = nbv1.NSStoreTypeAWSS3
				secret.StringData["AWS_ACCESS_KEY_ID"] = nsr.Identity
				secret.StringData["AWS_SECRET_ACCESS_KEY"] = namespaceResourceOperatorInfo.SecretKey
				secret.Name = fmt.Sprintf("namespace-store-%s-%s", nbv1.NSStoreTypeAWSS3, nsr.Name)
				nsStore.Spec.AWSS3 = &nbv1.AWSS3Spec{
					TargetBucket: nsr.TargetBucket,
					Secret: corev1.SecretReference{
						Name:      secret.Name,
						Namespace: secret.Namespace,
					},
				}
			case "AZURE":
				nsStore.Spec.Type = nbv1.NSStoreTypeAzureBlob
				secret.StringData["AccountName"] = nsr.Identity
				secret.StringData["AccountKey"] = namespaceResourceOperatorInfo.SecretKey
				secret.Name = fmt.Sprintf("namespace-store-%s-%s", nbv1.NSStoreTypeAzureBlob, nsr.Name)
				nsStore.Spec.AzureBlob = &nbv1.AzureBlobSpec{
					TargetBlobContainer: nsr.TargetBucket,
					Secret: corev1.SecretReference{
						Name:      secret.Name,
						Namespace: secret.Namespace,
					},
				}
			case "S3_COMPATIBLE":
				nsStore.Spec.Type = nbv1.NSStoreTypeS3Compatible
				secret.Name = fmt.Sprintf("namespace-store-%s-%s", nbv1.NSStoreTypeS3Compatible, nsr.Name)
				secret.StringData["AWS_ACCESS_KEY_ID"] = nsr.Identity
				secret.StringData["AWS_SECRET_ACCESS_KEY"] = namespaceResourceOperatorInfo.SecretKey

				var sig nbv1.S3SignatureVersion
				if nsr.AuthMethod == nb.CloudAuthMethodAwsV4 {
					sig = nbv1.S3SignatureVersionV4
				}
				if nsr.AuthMethod == nb.CloudAuthMethodAwsV2 {
					sig = nbv1.S3SignatureVersionV2
				}
				nsStore.Spec.S3Compatible = &nbv1.S3CompatibleSpec{
					TargetBucket: nsr.TargetBucket,
					Endpoint:     nsr.Endpoint,
					Secret: corev1.SecretReference{
						Name:      secret.Name,
						Namespace: secret.Namespace,
					},
					SignatureVersion: sig,
				}
			case "IBM_COS":
				nsStore.Spec.Type = nbv1.NSStoreTypeIBMCos
				secret.Name = fmt.Sprintf("namespace-store-%s-%s", nbv1.NSStoreTypeIBMCos, nsr.Name)
				secret.StringData["AWS_ACCESS_KEY_ID"] = nsr.Identity
				secret.StringData["AWS_SECRET_ACCESS_KEY"] = namespaceResourceOperatorInfo.SecretKey

				var sig nbv1.S3SignatureVersion
				if nsr.AuthMethod == nb.CloudAuthMethodAwsV4 {
					sig = nbv1.S3SignatureVersionV4
				}
				if nsr.AuthMethod == nb.CloudAuthMethodAwsV2 {
					sig = nbv1.S3SignatureVersionV2
				}
				nsStore.Spec.IBMCos = &nbv1.IBMCosSpec{
					TargetBucket: nsr.TargetBucket,
					Endpoint:     nsr.Endpoint,
					Secret: corev1.SecretReference{
						Name:      secret.Name,
						Namespace: secret.Namespace,
					},
					SignatureVersion: sig,
				}
			default:
				logrus.Errorf(`❌ Could not create NamespaceStore %q invalid endpoint type %q`, nsStore.Name, nsr.EndpointType)
				continue
			}

			// Create namespace store CR
			util.Panic(controllerutil.SetControllerReference(r.NooBaa, nsStore, scheme.Scheme))
			if !util.KubeCreateFailExisting(nsStore) {
				logrus.Errorf(`❌ Could not create NamespaceStore %q in Namespace %q (conflict)`, nsStore.Name, nsStore.Namespace)
				continue
			}

			if !util.KubeCheck(secret) {
				// Create secret
				util.Panic(controllerutil.SetControllerReference(nsStore, secret, scheme.Scheme))
				if !util.KubeCreateFailExisting(secret) {
					logrus.Errorf(`❌ Could not create Secret %q in Namespace %q (conflict)`, secret.Name, secret.Namespace)
					continue
				}
			}
			err = r.NBClient.SetNamespaceStoreInfo(nb.NamespaceStoreInfo{
				Name:      nsr.Name,
				Namespace: options.Namespace,
			})
			if err != nil {
				logrus.Infof("couldn't update namespace store info for namespace resource %q in namespace %q", nsr.Name, options.Namespace)
			}
		}
	}
	return nil
}

// ReconcileSystemConfigMap restarts the core and endpoint pods to recive updates from config map
func (r *Reconciler) ReconcileSystemConfigMap() error {
	sha256Hex := util.GetCmDataHash(r.CoreAppConfig.Data)

	if r.DeploymentEndpoint.ObjectMeta.Annotations["ConfigMapHash"] != sha256Hex {
		epPodList := &corev1.PodList{}
		epPodSelector, _ := labels.Parse("noobaa-s3=" + r.Request.Name)
		if !util.KubeList(epPodList, &client.ListOptions{Namespace: options.Namespace, LabelSelector: epPodSelector}) {
			r.Logger.Errorf("failed to list to endpoint pods")
		}
		r.DeploymentEndpoint.ObjectMeta.Annotations["ConfigMapHash"] = sha256Hex

		if !util.KubeUpdate(r.DeploymentEndpoint) {
			r.Logger.Infof("NooBaa %q failed to update Deployment Endpoint Config Map Hash", options.SystemName)
		}
		for _, pod := range epPodList.Items {
			util.KubeDeleteNoPolling(&pod)
		}
	}

	corePodList := &corev1.PodList{}
	corePodSelector, _ := labels.Parse("noobaa-core=" + r.Request.Name)
	if !util.KubeList(corePodList, &client.ListOptions{Namespace: options.Namespace, LabelSelector: corePodSelector}) {
		r.Logger.Errorf("failed to list to core pods")
	}

	if r.CoreApp.ObjectMeta.Annotations["ConfigMapHash"] != sha256Hex {
		r.CoreApp.ObjectMeta.Annotations["ConfigMapHash"] = sha256Hex

		if !util.KubeUpdate(r.CoreApp) {
			r.Logger.Infof("NooBaa %q failed to update Core STS Config Map Hash", options.SystemName)
		}

		for _, pod := range corePodList.Items {
			if pod.DeletionTimestamp == nil {
				util.KubeDeleteNoPolling(&pod)
			}
		}

	}

	return nil
}
