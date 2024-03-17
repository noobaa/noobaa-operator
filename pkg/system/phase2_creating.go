package system

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/libopenstorage/secrets"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/util/kms"
	secv1 "github.com/openshift/api/security/v1"
	cloudcredsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/robfig/cron/v3"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	webIdentityTokenPath string = "/var/run/secrets/openshift/serviceaccount/token"
	roleARNEnvVar        string = "ROLEARN"
)

// ReconcilePhaseCreating runs the reconcile phase
func (r *Reconciler) ReconcilePhaseCreating() error {

	r.SetPhase(
		nbv1.SystemPhaseCreating,
		"SystemPhaseCreating",
		"noobaa operator started phase 2/4 - \"Creating\"",
	)

	if err := r.ReconcileCAInject(); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceAccount, r.SetDesiredServiceAccount); err != nil {
		return err
	}
	if err := r.ReconcilePhaseCreatingForMainClusters(); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceS3, r.SetDesiredServiceS3); err != nil {
		return err
	}
	if err := r.ReconcileObjectOptional(r.RouteS3, r.SetDesiredRouteS3); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceSts, r.SetDesiredServiceSts); err != nil {
		return err
	}
	if err := r.ReconcileObjectOptional(r.RouteSts, nil); err != nil {
		return err
	}
	// the credentials that are created by cloud-credentials-operator sometimes take time
	// to be valid (requests sometimes returns InvalidAccessKeyId for 1-2 minutes)
	// creating the credential request as early as possible to try and avoid it
	if err := r.ReconcileBackingStoreCredentials(); err != nil {
		r.Logger.Errorf("failed to create CredentialsRequest. will retry in phase 4. error: %v", err)
		return err
	}
	if err := r.ReconcileObjectOptional(r.ServiceSyslog, r.SetDesiredServiceSyslog); err != nil {
		return err
	}

	return nil
}

// ReconcilePhaseCreatingForMainClusters reconcile all object for full deployment clusters
func (r *Reconciler) ReconcilePhaseCreatingForMainClusters() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

	if err := r.ReconcileObject(r.CoreAppConfig, r.SetDesiredCoreAppConfig); err != nil {
		return err
	}

	if r.CoreAppConfig.Data["NOOBAA_LOG_LEVEL"] == "warn" {
		r.Logger.Infof("Setting operator log level to Warn")
		util.InitLogger(logrus.WarnLevel)
	} else {
		util.InitLogger(logrus.DebugLevel)
	}

	// A failure to discover OAuth endpoints should not fail the entire reconcile phase.
	oAuthEndpoints, err := util.DiscoverOAuthEndpoints()
	if err != nil {
		r.Logger.Warnf("Discovery of OAuth endpoints failed, got: %v", err)
	}
	r.OAuthEndpoints = oAuthEndpoints

	if err := r.ReconcileObject(r.SecretServer, nil); err != nil {
		return err
	}
	if r.NooBaa.Spec.ExternalPgSecret == nil {
		if err := r.ReconcileObject(r.SecretDB, nil); err != nil {
			return err
		}
	}
	if err := r.ReconcileRootSecret(); err != nil {
		return err
	}

	// create the db only if postgres secret is not given
	if r.NooBaa.Spec.ExternalPgSecret == nil {
		if err := r.UpgradePostgresDB(); err != nil {
			return err
		}

		if err := r.ReconcileDB(); err != nil {
			return err
		}

		if err := r.ReconcileObject(r.ServiceDbPg, r.SetDesiredServiceDBForPostgres); err != nil {
			return err
		}
	}
	if err := r.ReconcileObject(r.ServiceMgmt, r.SetDesiredServiceMgmt); err != nil {
		return err
	}

	if err := r.ReconcileObject(r.CoreApp, r.SetDesiredCoreApp); err != nil {
		return err
	}

	if err := r.ReconcileObjectOptional(r.RouteMgmt, nil); err != nil {
		return err
	}

	return nil
}

// SetDesiredServiceAccount updates the ServiceAccount as desired for reconciling
func (r *Reconciler) SetDesiredServiceAccount() error {
	if r.ServiceAccount.Annotations == nil {
		r.ServiceAccount.Annotations = map[string]string{}
	}
	r.ServiceAccount.Annotations["serviceaccounts.openshift.io/oauth-redirectreference.noobaa-mgmt"] =
		`{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"` + r.RouteMgmt.Name + `"}}`
	return nil
}

// SetDesiredServiceMgmt updates the ServiceMgmt as desired for reconciling
func (r *Reconciler) SetDesiredServiceMgmt() error {
	r.ServiceMgmt.Spec.Selector["noobaa-mgmt"] = r.Request.Name
	r.ServiceMgmt.Labels["noobaa-mgmt-svc"] = "true"
	return nil
}

// SetDesiredServiceSyslog updates the ServiceSyslog as desired for reconciling
func (r *Reconciler) SetDesiredServiceSyslog() error {
	r.ServiceSyslog.Spec.Selector["noobaa-mgmt"] = r.Request.Name
	r.ServiceSyslog.Labels["noobaa-syslog-svc"] = "true"
	return nil
}

// SetDesiredServiceS3 updates the ServiceS3 as desired for reconciling
func (r *Reconciler) SetDesiredServiceS3() error {
	if r.NooBaa.Spec.DisableLoadBalancerService {
		r.ServiceS3.Spec.Type = corev1.ServiceTypeClusterIP
		r.ServiceS3.Spec.LoadBalancerSourceRanges = []string{}
	} else {
		// It is here in case disableLoadBalancerService is removed from the crd or changed to false
		r.ServiceS3.Spec.Type = corev1.ServiceTypeLoadBalancer
		r.ServiceS3.Spec.LoadBalancerSourceRanges = r.NooBaa.Spec.LoadBalancerSourceSubnets.S3
	}
	r.ServiceS3.Spec.Selector["noobaa-s3"] = r.Request.Name
	r.ServiceS3.Labels["noobaa-s3-svc"] = "true"
	return nil
}

// SetDesiredRouteS3 updates the RouteS3 as desired for reconciling
func (r *Reconciler) SetDesiredRouteS3() error {
	r.RouteS3.Spec.TLS.InsecureEdgeTerminationPolicy = "Allow"
	if r.NooBaa.Spec.DenyHTTP {
		r.RouteS3.Spec.TLS.InsecureEdgeTerminationPolicy = "None"
	}
	return nil
}

// SetDesiredServiceSts updates the ServiceSts as desired for reconciling
func (r *Reconciler) SetDesiredServiceSts() error {
	if r.NooBaa.Spec.DisableLoadBalancerService {
		r.ServiceSts.Spec.Type = corev1.ServiceTypeClusterIP
		r.ServiceSts.Spec.LoadBalancerSourceRanges = []string{}
	} else {
		// It is here in case disableLoadBalancerService is removed from the crd or changed to false
		r.ServiceSts.Spec.Type = corev1.ServiceTypeLoadBalancer
		r.ServiceSts.Spec.LoadBalancerSourceRanges = r.NooBaa.Spec.LoadBalancerSourceSubnets.STS
	}
	r.ServiceSts.Spec.Selector["noobaa-s3"] = r.Request.Name
	return nil
}

// SetDesiredServiceDBForPostgres updates the postgres service
func (r *Reconciler) SetDesiredServiceDBForPostgres() error {
	r.ServiceDbPg.Spec.Selector["noobaa-db"] = "postgres"
	r.ServiceDbPg.Spec.Ports[0].Name = "postgres"
	r.ServiceDbPg.Spec.Ports[0].Port = 5432
	r.ServiceDbPg.Spec.Ports[0].TargetPort = intstr.FromInt(5432)
	return nil
}

// SetDesiredNooBaaDB updates the NooBaaDB as desired for reconciling
func (r *Reconciler) SetDesiredNooBaaDB() error {
	var NooBaaDBTemplate *appsv1.StatefulSet = nil

	var NooBaaDB = r.NooBaaPostgresDB
	if dbLabels, ok := r.NooBaa.Spec.Labels["db"]; ok {
		NooBaaDB.Spec.Template.Labels = dbLabels
	}
	if dbAnnotations, ok := r.NooBaa.Spec.Annotations["db"]; ok {
		NooBaaDB.Spec.Template.Annotations = dbAnnotations
	}
	NooBaaDB.Spec.Template.Labels["noobaa-db"] = "postgres"
	NooBaaDB.Spec.Selector.MatchLabels["noobaa-db"] = "postgres"
	NooBaaDB.Spec.ServiceName = r.ServiceDbPg.Name
	NooBaaDBTemplate = util.KubeObject(bundle.File_deploy_internal_statefulset_postgres_db_yaml).(*appsv1.StatefulSet)

	podSpec := &NooBaaDB.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa-db"
	defaultUID := int64(10001)
	defaulfGID := int64(0)
	defaultFSGroup := int64(0)
	defaultFSGroupChangePolicy := corev1.FSGroupChangeOnRootMismatch

	podSpec.SecurityContext.RunAsUser = &defaultUID
	podSpec.SecurityContext.RunAsGroup = &defaulfGID
	podSpec.SecurityContext.FSGroup = &defaultFSGroup
	podSpec.SecurityContext.FSGroupChangePolicy = &defaultFSGroupChangePolicy
	podSpec.InitContainers = util.FilterSlice(
		podSpec.InitContainers,
		func(c corev1.Container) bool {
			// remove the init container that is not needed
			return c.Name != "init"
		},
	)

	for i := range podSpec.InitContainers {
		c := &podSpec.InitContainers[i]
		if c.Name == "initialize-database" {
			c.Image = GetDesiredDBImage(r.NooBaa)
		}
	}
	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		if c.Name == "db" {
			c.Image = GetDesiredDBImage(r.NooBaa)
			if r.NooBaa.Spec.DBResources != nil {
				c.Resources = *r.NooBaa.Spec.DBResources
			}

			c.Lifecycle = &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"/bin/sh", "-c", "pg_ctl -D /var/lib/pgsql/data/userdata/ -w -t 60 -m fast stop"},
					},
				},
			}

		}
	}

	if r.NooBaa.Spec.ImagePullSecret == nil {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{}
	} else {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{*r.NooBaa.Spec.ImagePullSecret}
	}
	podSpec.Tolerations = r.NooBaa.Spec.Tolerations
	podSpec.Affinity = r.NooBaa.Spec.Affinity

	if NooBaaDB.UID == "" {
		for i := range NooBaaDB.Spec.VolumeClaimTemplates {
			pvc := &NooBaaDB.Spec.VolumeClaimTemplates[i]
			pvc.Namespace = NooBaaDB.Namespace
			r.Own(pvc)
			// unsetting BlockOwnerDeletion to avoid error when trying to own pvc:
			// "cannot set blockOwnerDeletion if an ownerReference refers to a resource you can't set finalizers on"
			pvc.OwnerReferences[0].BlockOwnerDeletion = nil
			switch pvc.Name {
			case "db":
				if r.NooBaa.Spec.DBStorageClass != nil {
					pvc.Spec.StorageClassName = r.NooBaa.Spec.DBStorageClass
				} else {
					storageClassName, err := r.findLocalStorageClass()
					if err != nil {
						r.Logger.Errorf("got error finding a default/local storage class. error: %v", err)
						return err
					}
					pvc.Spec.StorageClassName = &storageClassName
				}
				if r.NooBaa.Spec.DBVolumeResources != nil {
					pvc.Spec.Resources = *r.NooBaa.Spec.DBVolumeResources
				}
			}
		}

	} else {
		// upgrade path add new resources,
		// merge the volumes & volumeMounts from the template
		util.MergeVolumeList(&NooBaaDB.Spec.Template.Spec.Volumes, &NooBaaDBTemplate.Spec.Template.Spec.Volumes)
		for i := range NooBaaDB.Spec.Template.Spec.Containers {
			util.MergeVolumeMountList(&NooBaaDB.Spec.Template.Spec.Containers[i].VolumeMounts, &NooBaaDBTemplate.Spec.Template.Spec.Containers[i].VolumeMounts)
		}

		// when already exists we check that there is no update requested to the volumes
		// otherwise we report that volume update is unsupported
		for i := range NooBaaDB.Spec.VolumeClaimTemplates {
			pvc := &NooBaaDB.Spec.VolumeClaimTemplates[i]
			switch pvc.Name {
			case "db":
				currentClass := ""
				desiredClass := ""
				if pvc.Spec.StorageClassName != nil {
					currentClass = *pvc.Spec.StorageClassName
				}
				if r.NooBaa.Spec.DBStorageClass != nil {
					desiredClass = *r.NooBaa.Spec.DBStorageClass
					if desiredClass != currentClass {
						r.Logger.Infof("No match between desired DB storage class in noobaa %s and current class in pvc %s",
							desiredClass, currentClass)
						r.Recorder.Eventf(r.NooBaa, corev1.EventTypeWarning, "DBStorageClassIsImmutable",
							"spec.dbStorageClass is immutable and cannot be updated for volume %q in existing %s %q"+
								" since it requires volume recreate and migrate which is unsupported by the operator",
							pvc.Name, r.CoreApp.TypeMeta.Kind, r.CoreApp.Name)
					}
				}
				if r.NooBaa.Spec.DBVolumeResources != nil &&
					!reflect.DeepEqual(pvc.Spec.Resources, *r.NooBaa.Spec.DBVolumeResources) {
					r.Logger.Infof("No match between DB volume resources")
					r.Recorder.Eventf(r.NooBaa, corev1.EventTypeWarning, "DBVolumeResourcesIsImmutable",
						"spec.dbVolumeResources is immutable and cannot be updated for volume %q in existing %s %q"+
							" since it requires volume recreate and migrate which is unsupported by the operator",
						pvc.Name, r.CoreApp.TypeMeta.Kind, r.CoreApp.Name)
				}
			}
		}
	}

	return nil
}

func (r *Reconciler) setDesiredCoreEnv(c *corev1.Container) {
	for j := range c.Env {
		switch c.Env[j].Name {
		case "AGENT_PROFILE":
			c.Env[j].Value = r.SetDesiredAgentProfile(c.Env[j].Value)

		case "POSTGRES_HOST":
			if r.NooBaa.Spec.ExternalPgSecret == nil {
				c.Env[j].Value = r.NooBaaPostgresDB.Name + "-0." + r.NooBaaPostgresDB.Spec.ServiceName
			}

		case "DB_TYPE":
			c.Env[j].Value = "postgres"

		case "OAUTH_AUTHORIZATION_ENDPOINT":
			if r.OAuthEndpoints != nil {
				c.Env[j].Value = r.OAuthEndpoints.AuthorizationEndpoint
			}

		case "OAUTH_TOKEN_ENDPOINT":
			if r.OAuthEndpoints != nil {
				c.Env[j].Value = r.OAuthEndpoints.TokenEndpoint
			}
		case "POSTGRES_USER":
			if r.NooBaa.Spec.ExternalPgSecret == nil {
				if c.Env[j].Value != "" {
					c.Env[j].Value = ""
				}
				c.Env[j].ValueFrom = &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: r.SecretDB.Name,
						},
						Key: "user",
					},
				}
			}
		case "POSTGRES_PASSWORD":
			if r.NooBaa.Spec.ExternalPgSecret == nil {
				if c.Env[j].Value != "" {
					c.Env[j].Value = ""
				}
				c.Env[j].ValueFrom = &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: r.SecretDB.Name,
						},
						Key: "password",
					},
				}
			}
		case "POSTGRES_CONNECTION_STRING":
			if r.NooBaa.Spec.ExternalPgSecret != nil {
				if c.Env[j].Value != "" {
					c.Env[j].Value = ""
				}
				c.Env[j].ValueFrom = &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: r.NooBaa.Spec.ExternalPgSecret.Name,
						},
						Key: "db_url",
					},
				}
			}
		case "POSTGRES_SSL_REQUIRED":
			if r.NooBaa.Spec.ExternalPgSSLRequired {
				c.Env[j].Value = "true"
			}
		case "POSTGRES_SSL_UNAUTHORIZED":
			if r.NooBaa.Spec.ExternalPgSSLUnauthorized {
				c.Env[j].Value = "true"
			}
		case "NOOBAA_ROOT_SECRET":
			c.Env[j].Value = r.SecretRootMasterKey
		case "NODE_EXTRA_CA_CERTS":
			c.Env[j].Value = r.ApplyCAsToPods
		}
	}
}

// SetDesiredCoreApp updates the CoreApp as desired for reconciling
func (r *Reconciler) SetDesiredCoreApp() error {
	if coreLabels, ok := r.NooBaa.Spec.Labels["core"]; ok {
		r.CoreApp.Spec.Template.Labels = coreLabels
	}
	if coreAnnotations, ok := r.NooBaa.Spec.Annotations["core"]; ok {
		r.CoreApp.Spec.Template.Annotations = coreAnnotations
	}
	r.CoreApp.Spec.Template.Labels["noobaa-core"] = r.Request.Name
	r.CoreApp.Spec.Template.Labels["noobaa-mgmt"] = r.Request.Name
	r.CoreApp.Spec.Selector.MatchLabels["noobaa-core"] = r.Request.Name
	r.CoreApp.Spec.ServiceName = r.ServiceMgmt.Name

	podSpec := &r.CoreApp.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa"
	coreImageChanged := false

	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		switch c.Name {
		case "core":
			if c.Image != r.NooBaa.Status.ActualImage {
				coreImageChanged = true
				c.Image = r.NooBaa.Status.ActualImage
			}
			// adding the missing Env variable from default container
			util.MergeEnvArrays(&c.Env, &r.DefaultCoreApp.Env)
			r.setDesiredCoreEnv(c)
			r.setDesiredRootMasterKeyMounts(podSpec, c)

			util.ReflectEnvVariable(&c.Env, "HTTP_PROXY")
			util.ReflectEnvVariable(&c.Env, "HTTPS_PROXY")
			util.ReflectEnvVariable(&c.Env, "NO_PROXY")

			if r.NooBaa.Spec.CoreResources != nil {
				c.Resources = *r.NooBaa.Spec.CoreResources
			}
			if util.KubeCheckQuiet(r.CaBundleConf) {
				configMapVolumeMounts := []corev1.VolumeMount{{
					Name:      r.CaBundleConf.Name,
					MountPath: "/etc/pki/ca-trust/extracted/pem",
					ReadOnly:  true,
				}}
				util.MergeVolumeMountList(&c.VolumeMounts, &configMapVolumeMounts)
			}
			if r.ExternalPgSSLSecret != nil && util.KubeCheckQuiet(r.ExternalPgSSLSecret) {
				secretVolumeMounts := []corev1.VolumeMount{{
					Name:      r.ExternalPgSSLSecret.Name,
					MountPath: "/etc/external-db-secret",
					ReadOnly:  true,
				}}
				util.MergeVolumeMountList(&c.VolumeMounts, &secretVolumeMounts)
			}
		case "noobaa-log-processor":
			if c.Image != r.NooBaa.Status.ActualImage {
				coreImageChanged = true
				c.Image = r.NooBaa.Status.ActualImage
			}
			// adding the missing Env variable from default container
			util.MergeEnvArrays(&c.Env, &r.DefaultCoreApp.Env)
			r.setDesiredCoreEnv(c)

			if r.NooBaa.Spec.LogResources != nil {
				c.Resources = *r.NooBaa.Spec.LogResources
			} else {
				var reqCPU, reqMem resource.Quantity
				reqCPU, _ = resource.ParseQuantity("200m")
				reqMem, _ = resource.ParseQuantity("500Mi")

				logResourceList := corev1.ResourceList{
					corev1.ResourceCPU:    reqCPU,
					corev1.ResourceMemory: reqMem,
				}
				c.Resources = corev1.ResourceRequirements{
					Requests: logResourceList,
					Limits:   logResourceList,
				}
			}
			if util.KubeCheckQuiet(r.CaBundleConf) {
				configMapVolumeMounts := []corev1.VolumeMount{{
					Name:      r.CaBundleConf.Name,
					MountPath: "/etc/pki/ca-trust/extracted/pem",
					ReadOnly:  true,
				}}
				util.MergeVolumeMountList(&c.VolumeMounts, &configMapVolumeMounts)
			}
		}
	}
	if r.NooBaa.Spec.ImagePullSecret == nil {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{}
	} else {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{*r.NooBaa.Spec.ImagePullSecret}
	}
	podSpec.Tolerations = r.NooBaa.Spec.Tolerations
	podSpec.Affinity = r.NooBaa.Spec.Affinity

	if r.CoreApp.UID == "" {
		// generate info event for the first creation of noobaa
		if r.Recorder != nil {
			r.Recorder.Eventf(r.NooBaa, corev1.EventTypeNormal,
				"NooBaaImage", `Using NooBaa image %q for the creation of %q`, r.NooBaa.Status.ActualImage, r.NooBaa.Name)
		}
	} else {
		if coreImageChanged {
			// generate info event for the first creation of noobaa
			if r.Recorder != nil {
				r.Recorder.Eventf(r.NooBaa, corev1.EventTypeNormal,
					"NooBaaImage", `Updating NooBaa image to %q for %q`, r.NooBaa.Status.ActualImage, r.NooBaa.Name)
			}
		}

	}

	if r.CoreApp.Spec.Template.Annotations == nil {
		r.CoreApp.Spec.Template.Annotations = make(map[string]string)
	}

	r.CoreApp.Spec.Template.Annotations["noobaa.io/configmap-hash"] = r.CoreAppConfig.Annotations["noobaa.io/configmap-hash"]

	replicas := int32(1)
	phase := r.NooBaa.Status.UpgradePhase
	if phase == nbv1.UpgradePhasePrepare || phase == nbv1.UpgradePhaseMigrate {
		replicas = int32(0)
	}
	pgUpdatePhase := r.NooBaa.Status.PostgresUpdatePhase
	if pgUpdatePhase == nbv1.UpgradePhasePrepare || pgUpdatePhase == nbv1.UpgradePhaseUpgrade || pgUpdatePhase == nbv1.UpgradePhaseClean {
		replicas = int32(0)
	}
	r.CoreApp.Spec.Replicas = &replicas

	if util.KubeCheckQuiet(r.CaBundleConf) {
		configMapVolumes := []corev1.Volume{{
			Name: r.CaBundleConf.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.CaBundleConf.Name,
					},
					Items: []corev1.KeyToPath{{
						Key:  "ca-bundle.crt",
						Path: "tls-ca-bundle.pem",
					}},
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &configMapVolumes)
	}
	if r.ExternalPgSSLSecret != nil && util.KubeCheckQuiet(r.ExternalPgSSLSecret) {
		secretVolumes := []corev1.Volume{{
			Name: r.ExternalPgSSLSecret.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: r.ExternalPgSSLSecret.Name,
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &secretVolumes)
	}
	return nil
}

// ReconcileBackingStoreCredentials creates a CredentialsRequest resource if necessary and returns
// the bucket name allowed for the credentials. nil is returned if cloud credentials are not supported
func (r *Reconciler) ReconcileBackingStoreCredentials() error {
	// Skip if joining another NooBaa
	r.Logger.Info("Reconciling Backing Store Credentials")
	if r.JoinSecret != nil {
		return nil
	}
	// If default backing store is disabled
	if r.NooBaa.Spec.ManualDefaultBackingStore {
		r.Logger.Info("ManualDefaultBackingStore is true, Skip Reconciling Backing Store Credentials")
		return nil
	}

	if util.IsAWSPlatform() {
		return r.ReconcileAWSCredentials()
	}
	if util.IsAzurePlatformNonGovernment() {
		return r.ReconcileAzureCredentials()
	}
	if util.IsGCPPlatform() {
		return r.ReconcileGCPCredentials()
	}
	if util.IsIBMPlatform() {
		return r.ReconcileIBMCredentials()
	}
	return r.ReconcileRGWCredentials()
}

// ReconcileRGWCredentials creates a ceph objectstore user if a ceph objectstore exists in the same namespace
func (r *Reconciler) ReconcileRGWCredentials() error {
	r.Logger.Info("Not running in AWS. will attempt to create a ceph objectstore user")
	util.KubeCheck(r.CephObjectStoreUser)
	if r.CephObjectStoreUser.UID != "" {
		return nil
	}

	// Try to list ceph object store users to validate that the CRD is installed in the cluster.
	cephObjectStoreUserList := &cephv1.CephObjectStoreUserList{}
	if !util.KubeList(cephObjectStoreUserList, &client.ListOptions{Namespace: options.Namespace}) {
		r.Logger.Info("failed to list ceph objectstore user, the scrd might not be installed in the cluster")
		return nil
	}

	// Try to list the ceph object stores.
	cephObjectStoreList := &cephv1.CephObjectStoreList{}
	if !util.KubeList(cephObjectStoreList, &client.ListOptions{Namespace: options.Namespace}) {
		r.Logger.Info("failed to list ceph objectstore to use as backing store")
		return nil
	}
	if len(cephObjectStoreList.Items) == 0 {
		r.Logger.Info("did not find any ceph objectstore to use as backing store")
		return nil
	}

	// Log all stores and take the first one for not.
	// TODO: need to decide what to do if multiple objectstores in one namespace
	r.Logger.Infof("found %d ceph objectstores: %v", len(cephObjectStoreList.Items), cephObjectStoreList.Items)
	storeName := cephObjectStoreList.Items[0].ObjectMeta.Name
	r.Logger.Infof("using objectstore %q as a default backing store", storeName)

	// create ceph objectstore user
	r.CephObjectStoreUser.Spec.Store = storeName
	r.Own(r.CephObjectStoreUser)
	err := r.Client.Create(r.Ctx, r.CephObjectStoreUser)
	if err != nil {
		r.Logger.Errorf("got error on CephObjectStoreUser creation. error: %v", err)
		return err
	}
	return nil
}

// ReconcileAWSCredentials creates a CredentialsRequest resource if cloud credentials operator is available
func (r *Reconciler) ReconcileAWSCredentials() error {
	// check if we have the env var ROLEARN that indicates that this is an OpenShift AWS STS cluster
	// cluster admin set this env (either in the UI in ARN details or via Subscription yaml) and set the mode to manual
	// olm will then copy the env from the subscription to the operator deployment (which is where your operator can pick it up from)
	roleARN := os.Getenv(roleARNEnvVar)
	r.Logger.Infof("Getting role ARN: %s = %s", roleARNEnvVar, roleARN)
	if roleARN != "" {
		if !arn.IsARN(roleARN) {
			r.Logger.Errorf("error with cloud credentials request, provided role ARN is invalid: %q", roleARN)
			return fmt.Errorf("provided role ARN is invalid: %q", roleARN)
		}
		r.IsAWSSTSCluster = true
	}

	arnPrefix := "arn:aws:s3:::"
	awsRegion, err := util.GetAWSRegion()
	if err != nil {
		r.Logger.Errorf("Got error from util.GetAWSRegion(). will use arnPrefix=%q. error=%v", arnPrefix, err)
	} else if awsRegion == "us-gov-east-1" || awsRegion == "us-gov-west-1" {
		arnPrefix = "arn:aws-us-gov:s3:::"
	}
	r.Logger.Info("Running in AWS. will create a CredentialsRequest resource")
	var bucketName string
	err = r.Client.Get(r.Ctx, util.ObjectKey(r.AWSCloudCreds), r.AWSCloudCreds)
	if err == nil {
		// credential request already exist. get the bucket name
		codec := cloudcredsv1.Codec
		awsProviderSpec := &cloudcredsv1.AWSProviderSpec{}
		err = codec.DecodeProviderSpec(r.AWSCloudCreds.Spec.ProviderSpec, awsProviderSpec)
		if err != nil {
			r.Logger.Error("error decoding providerSpec from cloud credentials request")
			return err
		}
		bucketName = strings.TrimPrefix(awsProviderSpec.StatementEntries[0].Resource, arnPrefix)
		r.Logger.Infof("found existing credential request for bucket %s, role ARN: %s", bucketName, roleARN)
		// since AWSSTSRoleARN is *string we will add the adderss of the variable roleARN only if it is not empty
		if r.IsAWSSTSCluster {
			r.DefaultBackingStore.Spec.AWSS3 = &nbv1.AWSS3Spec{
				TargetBucket:  bucketName,
				AWSSTSRoleARN: &roleARN,
			}
		} else {
			r.DefaultBackingStore.Spec.AWSS3 = &nbv1.AWSS3Spec{
				TargetBucket: bucketName,
			}
		}
		return nil
	}
	if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		// cloud credentials crd is missing. skip this stage
		return nil
	}
	if errors.IsNotFound(err) {
		// credential request does not exist. create one
		r.Logger.Info("Creating CredentialsRequest resource")
		bucketName = r.generateBackingStoreTargetName()
		codec := cloudcredsv1.Codec
		awsProviderSpec := &cloudcredsv1.AWSProviderSpec{}
		err = codec.DecodeProviderSpec(r.AWSCloudCreds.Spec.ProviderSpec, awsProviderSpec)
		if err != nil {
			r.Logger.Error("error decoding providerSpec from cloud credentials request")
			return err
		}
		// fix creds request according to bucket name
		awsProviderSpec.StatementEntries[0].Resource = arnPrefix + bucketName
		awsProviderSpec.StatementEntries[1].Resource = arnPrefix + bucketName + "/*"
		// add fields related to STS to creds request (role ARN)
		if r.IsAWSSTSCluster {
			awsProviderSpec.STSIAMRoleARN = roleARN
		}

		updatedProviderSpec, err := codec.EncodeProviderSpec(awsProviderSpec)
		if err != nil {
			r.Logger.Error("error encoding providerSpec for cloud credentials request")
			return err
		}
		r.AWSCloudCreds.Spec.ProviderSpec = updatedProviderSpec
		// add fields related to STS to creds request (path)
		if r.IsAWSSTSCluster {
			r.AWSCloudCreds.Spec.CloudTokenPath = webIdentityTokenPath
		}
		r.Own(r.AWSCloudCreds)
		err = r.Client.Create(r.Ctx, r.AWSCloudCreds)
		if err != nil {
			r.Logger.Errorf("got error when trying to create credentials request for bucket %s (STSIAMRoleARN %s). %v",
				bucketName, roleARN, err)
			return err
		}
		r.DefaultBackingStore.Spec.AWSS3 = &nbv1.AWSS3Spec{
			TargetBucket: bucketName,
		}
		return nil
	}
	return err
}

// ReconcileAzureCredentials creates a CredentialsRequest resource if cloud credentials operator is available
func (r *Reconciler) ReconcileAzureCredentials() error {
	r.Logger.Info("Running in Azure. will create a CredentialsRequest resource")
	err := r.Client.Get(r.Ctx, util.ObjectKey(r.AzureCloudCreds), r.AzureCloudCreds)
	if err == nil || meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		return nil
	}
	if errors.IsNotFound(err) {
		// credential request does not exist. create one
		r.Logger.Info("Creating CredentialsRequest resource")
		r.Own(r.AzureCloudCreds)
		err = r.Client.Create(r.Ctx, r.AzureCloudCreds)
		if err != nil {
			r.Logger.Errorf("got error when trying to create credentials request for azure. %v", err)
			return err
		}
		return nil
	}
	return err
}

// ReconcileGCPCredentials creates a CredentialsRequest resource if cloud credentials operator is available
func (r *Reconciler) ReconcileGCPCredentials() error {
	r.Logger.Info("Running on GCP. will create a CredentialsRequest resource")
	err := r.Client.Get(r.Ctx, util.ObjectKey(r.GCPCloudCreds), r.GCPCloudCreds)
	if err == nil || meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
		return nil
	}
	if errors.IsNotFound(err) {
		// credential request does not exist. create one
		r.Logger.Info("Creating CredentialsRequest resource")
		r.Own(r.GCPCloudCreds)
		err = r.Client.Create(r.Ctx, r.GCPCloudCreds)
		if err != nil {
			r.Logger.Errorf("got error when trying to create credentials request for GCP. %v", err)
			return err
		}
		return nil
	}
	return err
}

// ReconcileIBMCredentials sets IsIBMCloud to indicate operator is running in IBM Cloud
func (r *Reconciler) ReconcileIBMCredentials() error {
	// Currently IBM Cloud is not supported by cloud credential operator
	// In IBM Cloud, the COS Creds will be provided through Secret.
	r.Logger.Info("Running in IBM Cloud")
	util.KubeCheck(r.IBMCosBucketCreds)
	if r.IBMCosBucketCreds.UID == "" {
		r.Logger.Infof("%q secret is not present", r.IBMCosBucketCreds.Name)
		return nil
	}
	return nil
}

// ReconcileCAInject checks if a namespace called openshift-config exist (OCP)
// if so creates a cofig map for OCP to inject supported CAs to
func (r *Reconciler) ReconcileCAInject() error {
	ocpConfigNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: "openshift-config",
		},
	}
	if util.KubeCheckQuiet(ocpConfigNamespace) {
		r.Logger.Infof("Found openshift-config ns - will reconcile CA inject configmap: %q", r.CaBundleConf.Name)
		if err := r.ReconcileObject(r.CaBundleConf, nil); err != nil {
			return err
		}
	}
	return nil
}

// SetDesiredAgentProfile updates the value of the AGENT_PROFILE env
func (r *Reconciler) SetDesiredAgentProfile(profileString string) string {
	agentProfile := map[string]interface{}{}
	err := json.Unmarshal([]byte(profileString), &agentProfile)
	if err != nil {
		r.Logger.Infof("SetDesiredAgentProfile: ignore non-json AGENT_PROFILE value %q: %v", profileString, err)
	}
	agentProfile["image"] = r.NooBaa.Status.ActualImage
	if r.NooBaa.Spec.PVPoolDefaultStorageClass != nil {
		agentProfile["storage_class"] = *r.NooBaa.Spec.PVPoolDefaultStorageClass
	} else {
		delete(agentProfile, "storage_class")
	}
	profileBytes, err := json.Marshal(agentProfile)
	util.Panic(err)
	return string(profileBytes)
}

func (r *Reconciler) setKMSConditionStatus(s corev1.ConditionStatus) {
	conditions := &r.NooBaa.Status.Conditions
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: metav1.NewTime(time.Now()),
		Type:              nbv1.ConditionTypeKMSStatus,
		Status:            s,
	})
	r.Logger.Infof("setKMSConditionStatus %v", s)
}

func (r *Reconciler) reportKMSConditionStatus() {
	conditions := &r.NooBaa.Status.Conditions
	cond := conditionsv1.FindStatusCondition(*conditions, nbv1.ConditionTypeKMSStatus)
	if cond == nil {
		util.Logger().Printf("❌ Missing KMS status: %q\n", nbv1.ConditionTypeKMSStatus)
		return
	}

	st := cond.Status
	var kmsStatusBadge string
	if kms.StatusValid(st) {
		kmsStatusBadge = "✅"
	} else {
		kmsStatusBadge = "❌"
	}
	util.Logger().Printf("%v KMS status: %q\n", kmsStatusBadge, st)
}

func (r *Reconciler) setKMSConditionType(t string) {
	conditions := &r.NooBaa.Status.Conditions
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		LastHeartbeatTime: metav1.NewTime(time.Now()),
		Type:              nbv1.ConditionTypeKMSType,
		Status:            corev1.ConditionStatus(t),
	})
	r.Logger.Infof("setKMSConditionType %v", t)
}

// ReconcileRootSecret choose KMS for root secret key
func (r *Reconciler) ReconcileRootSecret() error {
	// External KMS Spec
	connectionDetails := r.NooBaa.Spec.Security.KeyManagementService.ConnectionDetails
	authTokenSecretName := r.NooBaa.Spec.Security.KeyManagementService.TokenSecretName

	k, err := kms.NewKMS(connectionDetails, authTokenSecretName, r.Request.Name, r.Request.Namespace, string(r.NooBaa.UID))
	if err != nil {
		r.Logger.Errorf("ReconcileRootSecret, NewKMS error %v", err)
		r.setKMSConditionStatus(nbv1.ConditionKMSInvalid)
		return err
	}
	r.setKMSConditionType(k.Type)

	conditionStatus := nbv1.ConditionKMSSync // default condition status
	if err := k.Get(); err != nil {
		if err != secrets.ErrInvalidSecretId {
			// Unknown get error
			r.Logger.Errorf("ReconcileRootSecret, KMS Get error %v", err)
			r.setKMSConditionStatus(nbv1.ConditionKMSErrorRead)
			return err
		}

		// the KMS root key was empty
		// Initialize external KMS with a randomly generated key
		err = k.Set(util.RandomBase64(32))
		if err != nil {
			r.Logger.Errorf("ReconcileRootSecret, KMS Set error %v", err)
			r.setKMSConditionStatus(nbv1.ConditionKMSErrorWrite)
			return err
		}
		// The key was set the first time - status init
		conditionStatus = nbv1.ConditionKMSInit
	}

	// Set the value from KMS
	err = k.Reconcile(r)
	if err != nil {
		r.Logger.Errorf("ReconcileRootSecret, KMS reconcile error %v", err)
		r.setKMSConditionStatus(nbv1.ConditionKMSErrorSecretReconcile)
		return err
	}

	r.setKMSConditionStatus(conditionStatus)

	if err := r.ReconcileKeyRotation(); err != nil {
		return err
	}

	return nil
}

// ReconcileKeyRotation checks if key rotation is enabled
// if so creates a cron job to run every time set
func (r *Reconciler) ReconcileKeyRotation() error {
	r.Logger.Infof("ReconcileKeyRotation, KMS Starting")

	if len(r.SecretRootMasterKey) > 0 {
		// key rotation not supported by the KMS backend
		r.Logger.Infof("ReconcileKeyRotation, KMS skip reconcile, disabled by configuration")
		return nil
	}

	enabledKeyRotation := r.NooBaa.Spec.Security.KeyManagementService.EnableKeyRotation
	if !enabledKeyRotation {
		r.Logger.Infof("ReconcileKeyRotation, KMS skip reconcile, single master root mode")
		return nil
	}
	scheduleCronSpec := r.NooBaa.Spec.Security.KeyManagementService.Schedule
	if len(scheduleCronSpec) == 0 {
		r.Logger.Infof("ReconcileKeyRotation, KMS default monthly rotation schedule")
		scheduleCronSpec = "0 0 1 * *" // “At 00:00 on day-of-month 1.”
	}

	schedule, err := cron.ParseStandard(scheduleCronSpec)
	if err != nil {
		r.Logger.Errorf("ReconcileKeyRotation, KMS rotation schedule parse error: %v", err)
		return err
	}
	lastTime := r.NooBaa.Status.LastKeyRotateTime.Time
	now := time.Now()
	if lastTime.IsZero() {
		r.Logger.Infof("ReconcileKeyRotation, KMS skip reconcile: initial set LastKeyRotateTime %v", now)
		r.NooBaa.Status.LastKeyRotateTime = metav1.Time{Time: now}
		return nil
	}
	nextSchedule := schedule.Next(lastTime)

	r.Logger.Infof("ReconcileKeyRotation, KMS rotation nextSchedule: %v", nextSchedule)
	if nextSchedule.After(now) {
		r.Logger.Infof("ReconcileKeyRotation, KMS skip reconcile, now %v before nextSchedule %v", now, nextSchedule)
		return nil
	}
	err = r.keyRotate()
	if err != nil {
		r.Logger.Errorf("ReconcileKeyRotation, KMS keyRotate error %v", err)
		return err
	}
	r.Logger.Infof("ReconcileKeyRotation, KMS Updating Last Key Rotate time: %v", now)
	r.NooBaa.Status.LastKeyRotateTime = metav1.Time{Time: now}

	return nil
}

// keyRotate sets new master root key
// and starts change propogation to the nooba pods
func (r *Reconciler) keyRotate() error {
	r.Logger.Infof("Key rotation started at %v", time.Now())
	if len(r.SecretRootMasterKey) > 0 {
		return fmt.Errorf("Master root key rotation is not supported by the KMS backend")
	}

	// KMS Spec
	connectionDetails := r.NooBaa.Spec.Security.KeyManagementService.ConnectionDetails
	authTokenSecretName := r.NooBaa.Spec.Security.KeyManagementService.TokenSecretName

	k, err := kms.NewKMS(connectionDetails, authTokenSecretName, r.Request.Name, r.Request.Namespace, string(r.NooBaa.UID))
	if err != nil {
		r.Logger.Errorf("keyRotate, NewKMS error %v", err)
		r.setKMSConditionStatus(nbv1.ConditionKMSInvalid)
		return err
	}

	// Generate new random root key and set it in the KMS
	// Key - rotate begins
	err = k.Set(util.RandomBase64(32))
	if err != nil {
		r.Logger.Errorf("keyRotate, KMS Set error %v", err)
		r.setKMSConditionStatus(nbv1.ConditionKMSErrorWrite)
		return err
	}

	// Set the value from KMS
	// Set new root key with system reconciler
	// to propogate change to noobaa pods
	err = k.Reconcile(r)
	if err != nil {
		r.Logger.Errorf("keyRotate, KMS reconcile error %v", err)
		r.setKMSConditionStatus(nbv1.ConditionKMSErrorSecretReconcile)
		return err
	}

	r.setKMSConditionStatus(nbv1.ConditionKMSKeyRotate)

	return nil
}

// ReconcileSecretMap sets the root master key for rotating key / map
func (r *Reconciler) ReconcileSecretMap(data map[string]string) error {
	if data == nil {
		return fmt.Errorf("System Reconciler ReconcileSecretMap data is nil")
	}
	r.SecretRootMasterMap.StringData = data
	if err := r.ReconcileObject(r.SecretRootMasterMap, nil); err != nil {
		return err
	}
	return nil
}

// ReconcileSecretString sets the root master key for single string secret
func (r *Reconciler) ReconcileSecretString(data string) error {
	if len(data) == 0 {
		return fmt.Errorf("System Reconciler ReconcileSecreString data len is zero")
	}
	r.SecretRootMasterKey = data
	return nil
}

func (r *Reconciler) reconcileRbac(scc, sa, role, binding string) error {
	SCC := util.KubeObject(scc).(*secv1.SecurityContextConstraints)
	if ok := util.KubeApply(SCC); ok {
		r.Logger.Infof("ReconcileRbac: SCC %q", SCC.Name)
	}

	SA := util.KubeObject(sa).(*corev1.ServiceAccount)
	SA.Namespace = options.Namespace
	if err := r.ReconcileObject(SA, nil); err != nil {
		return err
	}
	Role := util.KubeObject(role).(*rbacv1.Role)
	Role.Namespace = options.Namespace
	if err := r.ReconcileObject(Role, nil); err != nil {
		return err
	}
	RoleBinding := util.KubeObject(binding).(*rbacv1.RoleBinding)
	RoleBinding.Namespace = options.Namespace
	if err := r.ReconcileObject(RoleBinding, nil); err != nil {
		return err
	}

	return nil
}

// reconcileDBRBAC creates DB scc, role, rolebinding and service account
func (r *Reconciler) reconcileDBRBAC() error {
	return r.reconcileRbac(
		bundle.File_deploy_scc_db_yaml,
		bundle.File_deploy_service_account_db_yaml,
		bundle.File_deploy_role_db_yaml,
		bundle.File_deploy_role_binding_db_yaml)
}

// ReconcileDB choose between different types of DB
func (r *Reconciler) ReconcileDB() error {

	if r.NooBaa.Spec.ExternalPgSecret != nil {
		return nil
	}

	if err := r.reconcileDBRBAC(); err != nil {
		return err
	}

	// those are config maps required by the NooBaaPostgresDB StatefulSet,
	// if the configMap was not created at this step, NooBaaPostgresDB
	// would fail to start.

	isDBInitUpdated, reconcileDbError := r.ReconcileDBConfigMap(r.PostgresDBInitDb, r.SetDesiredPostgresDBInitDb)
	if reconcileDbError != nil {
		return reconcileDbError
	}

	isDBConfUpdated, reconcileDbError := r.ReconcileDBConfigMap(r.PostgresDBConf, r.SetDesiredPostgresDBConf)
	if reconcileDbError != nil {
		return reconcileDbError
	}

	phase := r.NooBaa.Status.PostgresUpdatePhase
	// during postgres upgrade, don't reconcile DB
	if phase != nbv1.UpgradePhaseFinished && phase != nbv1.UpgradePhaseNone {
		r.Logger.Infof("during postgres db upgrade")
		return nil
	}

	result, reconcilePostgresError := r.reconcileObjectAndGetResult(r.NooBaaPostgresDB, r.SetDesiredNooBaaDB, false)
	if reconcilePostgresError != nil {
		return reconcilePostgresError
	}
	if !r.isObjectUpdated(result) && (isDBInitUpdated || isDBConfUpdated) {
		r.Logger.Warn("One of the db configMap was updated but not postgres db")
		restartError := r.RestartDbPods()
		if restartError != nil {
			r.Logger.Warn("Unable to restart db pods")
		}

	}
	return nil
}

// ReconcileDBConfigMap reconcile provided postgres db config map
func (r *Reconciler) ReconcileDBConfigMap(cm *corev1.ConfigMap, desiredFunc func() error) (bool, error) {
	r.Own(cm)
	result, error := r.reconcileObjectAndGetResult(cm, desiredFunc, false)
	if error != nil {
		return false, fmt.Errorf("could not update Postgres DB configMap %q in Namespace %q", cm.Name, cm.Namespace)
	}
	return r.isObjectUpdated(result), nil
}

// SetDesiredPostgresDBConf fill desired postgres db config map
func (r *Reconciler) SetDesiredPostgresDBConf() error {
	dbConfigYaml := util.KubeObject(bundle.File_deploy_internal_configmap_postgres_db_yaml).(*corev1.ConfigMap)
	r.PostgresDBConf.Data = dbConfigYaml.Data

	overrideField := "noobaa-postgres.conf"
	operator := r.NooBaa
	if operator.Spec.DBConf != nil {
		// If the user has specified a custom "dbConf" in the NooBaa CR then proceed to append that configuration
		// to the pre-defined configuration
		r.PostgresDBConf.Data[overrideField] = dbConfigYaml.Data[overrideField] + "\n" + *operator.Spec.DBConf
	}

	return nil
}

// SetDesiredPostgresDBInitDb fill desired postgres db init config map
func (r *Reconciler) SetDesiredPostgresDBInitDb() error {
	postgresDBInitDbYaml := util.KubeObject(bundle.File_deploy_internal_configmap_postgres_initdb_yaml).(*corev1.ConfigMap)
	r.PostgresDBInitDb.Data = postgresDBInitDbYaml.Data
	return nil
}

// RestartDbPods restart db pods
func (r *Reconciler) RestartDbPods() error {
	r.Logger.Warn("Restarting postgres db pods")

	dbPodList := &corev1.PodList{}
	dbPodSelector, _ := labels.Parse("noobaa-db=postgres")
	if !util.KubeList(dbPodList, &client.ListOptions{Namespace: options.Namespace, LabelSelector: dbPodSelector}) {
		return fmt.Errorf("failed to list db pods in Namespace %q", options.Namespace)
	}
	for _, pod := range dbPodList.Items {
		if pod.DeletionTimestamp == nil {
			util.KubeDeleteNoPolling(&pod)
		}
	}
	return nil
}

// ReconcileSetDbImageAndInitCode changes postgres image and upgrade-db command script
func (r *Reconciler) ReconcileSetDbImageAndInitCode(targetImage string, initScript string, addInit bool) error {
	r.Logger.Infof("ReconcileSetDbImageAndInitCode:: changing DB image: %s and init contatiners script: %s",
		targetImage, initScript)
	if err := r.ReconcileObject(r.NooBaaPostgresDB, func() error {
		podSpec := &r.NooBaaPostgresDB.Spec.Template.Spec
		podSpec.ServiceAccountName = "noobaa"
		if addInit {
			upgradeContainer := podSpec.InitContainers[0]
			upgradeContainer.Name = "upgrade-db"
			upgradeContainer.Command = []string{"sh", "-x", initScript}
			upgradeContainer.Image = targetImage
			podSpec.InitContainers = append(podSpec.InitContainers, upgradeContainer)
		} else {
			for i := range podSpec.InitContainers {
				c := &podSpec.InitContainers[i]
				if c.Name == "upgrade-db" {
					c.Command = []string{"sh", "-x", initScript}
					c.Image = targetImage
				}
			}
		}
		for i := range podSpec.Containers {
			c := &podSpec.Containers[i]
			if c.Name == "db" {
				c.Image = targetImage
			}
		}
		return nil
	}); err != nil {
		r.Logger.Errorf("got error on postgres STS reconcile %v", err)
		return err
	}
	return nil
}

// HasUpgradeDbContainerFailed checks if postgres upgrade container failed
func (r *Reconciler) HasUpgradeDbContainerFailed(dbPod *corev1.Pod) string {
	//when init container restart policy is Always, pod state will not be set to failed on container error.
	//check for persistent error in container
	for i := range dbPod.Status.InitContainerStatuses {
		if dbPod.Status.InitContainerStatuses[i].Name == "upgrade-db" {
			r.Logger.Infof("HasUpgradeDbContainerFailed: Checking state of upgrade-db container: %v", dbPod.Status.InitContainerStatuses[i].State)
			if dbPod.Status.InitContainerStatuses[i].State.Terminated != nil {
				return dbPod.Status.InitContainerStatuses[i].State.Terminated.Reason
			}
			if dbPod.Status.InitContainerStatuses[i].State.Waiting != nil {
				return dbPod.Status.InitContainerStatuses[i].State.Waiting.Reason
			}
			if dbPod.Status.InitContainerStatuses[i].State.Running != nil {
				return "Running"
			}
		}
	}
	return "NotFound"
}

// RevertPostgresUpgrade reverts the postgres version to original version after failed upgrade
func (r *Reconciler) RevertPostgresUpgrade() error {
	r.Logger.Infof("RevertPostgresUpgrade reverting postgres upgrade to original database image")
	if err := r.ReconcileSetDbImageAndInitCode(*r.NooBaa.Status.BeforeUpgradeDbImage, "/init/revertdb.sh", false); err != nil {
		r.Logger.Errorf("got error on postgres STS reconcile %v", err)
		return err
	}
	//since DB pod is in failed state need to restart it
	restartError := r.RestartDbPods()
	if restartError != nil {
		r.Logger.Warn("Unable to restart db pods")
	}
	return nil
}

// UpgradePostgresDB upgrade postgres image to newer verision
func (r *Reconciler) UpgradePostgresDB() error {
	phase := r.NooBaa.Status.PostgresUpdatePhase
	if phase == nbv1.UpgradePhaseFinished || phase == nbv1.UpgradePhaseNone {
		return nil
	}

	var err error = nil
	hasInitContainerFailed := false
	r.Logger.Infof("UpgradePostgresDB: current phase is %s", phase)
	switch phase {

	case nbv1.UpgradePhaseFailed:
		if _, ok := r.NooBaa.Annotations["manual_upgrade_completed"]; ok {
			r.Logger.Infof("UpgradePostgresDB: annotation for manual upgrade completed was set, marking upgrade as finished")
			phase = nbv1.UpgradePhaseFinished
			break
		}
		if _, ok := r.NooBaa.Annotations["retry_upgrade"]; !ok {
			return nil
		}
		r.Logger.Infof("UpgradePostgresDB: annotation for retrying the upgrade was set, restarting upgrade")
		fallthrough
	case "":
		sts := util.KubeObject(bundle.File_deploy_internal_statefulset_postgres_db_yaml).(*appsv1.StatefulSet)
		sts.Name = "noobaa-db-pg"
		sts.Namespace = options.Namespace
		//postgres database doesn't exists, no need to upgrade.
		if !util.KubeCheckQuiet(sts) {
			r.Logger.Infof("UpgradePostgresDB: old STS doesn't exist - no need for upgrade")
			phase = nbv1.UpgradePhaseNone
			break
		}

		oldImage, ok := os.LookupEnv("NOOBAA_PSQL_12_IMAGE")
		if !ok {
			r.Logger.Errorf("UpgradePostgresDB: Missing critical env variable for upgrade - NOOBAA_PSQL_12_IMAGE")
			return util.NewPersistentError("MissingEnvVariable",
				"Missing critical env variable for pg upgrade - NOOBAA_PSQL_12_IMAGE")
		}
		r.Logger.Infof("UpgradePostgresDB: found ENV of pgsql version 12: %s", oldImage)
		if *r.NooBaa.Spec.DBImage != os.Getenv("NOOBAA_DB_IMAGE") {
			r.Logger.Infof("UpgradePostgresDB: NooBaa CR DB image: %s and operator DB image: %s are not the same, waiting...",
				*r.NooBaa.Spec.DBImage, os.Getenv("NOOBAA_DB_IMAGE"))
			return nil
		}
		r.Recorder.Eventf(r.NooBaa, corev1.EventTypeNormal, "DbUpgrade",
			"Configuring DB pod to run dbdump of PostgreSQL data")
		r.Logger.Infof("UpgradePostgresDB: setting phase to %s", nbv1.UpgradePhasePrepare)
		phase = nbv1.UpgradePhasePrepare
		r.NooBaa.Status.BeforeUpgradeDbImage = &oldImage

	case nbv1.UpgradePhasePrepare:
		if err := r.SetEndpointsDeploymentReplicas(0); err != nil {
			r.Logger.Errorf("UpgradePostgresDB::got error on endpoints deployment reconcile %v", err)
			return err
		}
		if err = r.ReconcileSetDbImageAndInitCode(*r.NooBaa.Status.BeforeUpgradeDbImage, "/init/dumpdb.sh", true); err != nil {
			r.Logger.Errorf("UpgradePostgresDB: got error on postgres STS reconcile %v", err)
			break
		}
		r.Logger.Infof("UpgradePostgresDB: restarting DB pods after updating pod's init contatiner for db-dump")
		restartError := r.RestartDbPods() // make sure new pods will start
		if restartError != nil {
			r.Logger.Warnf("UpgradePostgresDB: Unable to restart db pods %v", restartError)
			return restartError
		}
		r.Recorder.Eventf(r.NooBaa, corev1.EventTypeNormal, "DbUpgrade",
			"Starting dump. Start running DB dump in version 12 format")
		phase = nbv1.UpgradePhaseUpgrade

	case nbv1.UpgradePhaseUpgrade:
		dbPod := &corev1.Pod{}
		dbPod.Name = "noobaa-db-pg-0"
		dbPod.Namespace = r.NooBaaPostgresDB.Namespace
		if !util.KubeCheckQuiet(dbPod) {
			return nil
		}
		// make sure previous step has finished
		if dbPod.ObjectMeta.DeletionTimestamp != nil {
			r.Logger.Infof("UpgradePostgresDB: upgrade-db is not yet running, phase is: %s and deletion time stamp is %v",
				dbPod.Status.Phase, dbPod.ObjectMeta.DeletionTimestamp)
			return nil
		}
		status := r.HasUpgradeDbContainerFailed(dbPod)
		if status == "PodInitializing" || status == "Running" {
			r.Logger.Infof("UpgradePostgresDB: upgrade-db container didn't finished dump yet, waiting...")
			return nil
		}
		if status == "CrashLoopBackOff" || status == "Error" || status == "NotFound" {
			r.Logger.Errorf("UpgradePostgresDB: upgrade-db container failed dump DB. set to revert the image")
			hasInitContainerFailed = true
			break
		}
		r.Logger.Infof("UpgradePostgresDB: upgrade-db container didn't fail and finished dumping the old data [status=%s]", status)
		if err = r.ReconcileSetDbImageAndInitCode(GetDesiredDBImage(r.NooBaa), "/init/upgradedb.sh", false); err != nil {
			r.Logger.Errorf("UpgradePostgresDB: got error on postgres STS reconcile %v", err)
			break
		}
		r.Logger.Infof("UpgradePostgresDB: restarting DB pods after updating pod's init contatiner for upgrade-db")
		restartError := r.RestartDbPods() // make sure new pods will start
		if restartError != nil {
			r.Logger.Warn("UpgradePostgresDB: Unable to restart db pods")
		}
		r.Recorder.Eventf(r.NooBaa, corev1.EventTypeNormal, "DbUpgrade",
			"Finished dump. Start running upgrade of DB data from version 12 format to version 15")
		phase = nbv1.UpgradePhaseClean

	case nbv1.UpgradePhaseClean:
		dbPod := &corev1.Pod{}
		dbPod.Name = "noobaa-db-pg-0"
		dbPod.Namespace = r.NooBaaPostgresDB.Namespace
		if !util.KubeCheckQuiet(dbPod) {
			return nil
		}
		// make sure previous step has finished
		if dbPod.ObjectMeta.DeletionTimestamp != nil {
			r.Logger.Infof("UpgradePostgresDB: upgrade-db is not yet running, phase is: %s and deletion time stamp is %v",
				dbPod.Status.Phase, dbPod.ObjectMeta.DeletionTimestamp)
			return nil
		}
		r.Logger.Info("UpgradePostgresDB: DB pod is found")
		status := r.HasUpgradeDbContainerFailed(dbPod)
		if status == "PodInitializing" || status == "Running" {
			r.Logger.Infof("UpgradePostgresDB: upgrade-db container didn't finished upgrading data, waiting...")
			return nil
		}
		if status == "CrashLoopBackOff" || status == "Error" || status == "NotFound" {
			r.Logger.Errorf("UpgradePostgresDB: upgrade-db container failed upgrading data. set to revert the image")
			hasInitContainerFailed = true
			break
		}
		r.Logger.Infof("UpgradePostgresDB: upgrade-db container didn't fail and finished moving data to new version [status=%s]", status)
		if dbPod.Status.Phase != "Running" && dbPod.ObjectMeta.DeletionTimestamp == nil {
			r.Logger.Infof("UpgradePostgresDB: upgrade-db is not yet running, phase is: %s and deletion time stamp is %v",
				dbPod.Status.Phase, dbPod.ObjectMeta.DeletionTimestamp)
			return nil
		}
		//remove upgrade-db container
		if err = r.ReconcileObject(r.NooBaaPostgresDB, r.removeUpgradeContainer); err != nil {
			r.Logger.Errorf("got error on postgres STS reconcile %v", err)
			break
		}
		r.Logger.Infof("UpgradePostgresDB: upgrade-db container removed with success")
		if err := r.SetEndpointsDeploymentReplicas(1); err != nil {
			r.Logger.Errorf("UpgradePostgresDB::got error on endpoints deployment reconcile %v", err)
			return err
		}
		r.Recorder.Eventf(r.NooBaa, corev1.EventTypeNormal, "DbUpgrade",
			"Finished data upgrade after upgrade of data to PostgreSQL 15 format")
		phase = nbv1.UpgradePhaseFinished

	case nbv1.UpgradePhaseReverting:
		dbPod := &corev1.Pod{}
		dbPod.Name = "noobaa-db-pg-0"
		dbPod.Namespace = r.NooBaaPostgresDB.Namespace
		if !util.KubeCheckQuiet(dbPod) {
			return nil
		}
		// make sure previous step has finished
		if dbPod.ObjectMeta.DeletionTimestamp != nil {
			r.Logger.Infof("UpgradePostgresDB: upgrade-db is not yet running, phase is: %s and deletion time stamp is %v",
				dbPod.Status.Phase, dbPod.ObjectMeta.DeletionTimestamp)
			return nil
		}
		r.Logger.Info("UpgradePostgresDB: DB pod is found")
		status := r.HasUpgradeDbContainerFailed(dbPod)
		if status == "PodInitializing" || status == "Running" {
			r.Logger.Infof("UpgradePostgresDB: upgrade-db container didn't finished upgrading data, waiting...")
			return nil
		}
		if status == "CrashLoopBackOff" || status == "Error" {
			r.Logger.Errorf("UpgradePostgresDB: upgrade-db container failed upgrading data. set to revert the image")
			hasInitContainerFailed = true
			break
		}
		r.Logger.Infof("UpgradePostgresDB: upgrade-db container didn't fail and finished reverting data to old version [status=%s]", status)
		if err = r.ReconcileObject(r.NooBaaPostgresDB, r.removeUpgradeContainer); err != nil {
			r.Logger.Errorf("UpgradePostgresDB: got error on postgres STS reconcile %v", err)
			return err
		}
		if err := r.SetEndpointsDeploymentReplicas(1); err != nil {
			r.Logger.Errorf("UpgradePostgresDB::got error on endpoints deployment reconcile %v", err)
			return err
		}
		phase = nbv1.UpgradePhaseFailed
		r.Logger.Infof("UpgradePostgresDB: revert has succeeded, need to fix issue and re-attempt upgrade")
		r.Recorder.Eventf(r.NooBaa, corev1.EventTypeWarning, "DbUpgradeFailure",
			"Failed to update DB version to pgsql15. DB reverted back to 12.")
	}

	if (err != nil && util.IsPersistentError(err)) || hasInitContainerFailed {
		err = r.RevertPostgresUpgrade()
		if err == nil {
			r.Logger.Infof("UpgradePostgresDB: restarting DB pods after updating pod's init contatiner for revert-db")
		} else {
			r.Logger.Errorf("UpgradePostgresDB:failed reverting the DB %v, will retry revert", err)
		}
		phase = nbv1.UpgradePhaseReverting
	}

	r.NooBaa.Status.PostgresUpdatePhase = phase
	err = r.UpdateStatus()

	return err
}

func (r *Reconciler) removeUpgradeContainer() error {
	podSpec := &r.NooBaaPostgresDB.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa"
	for i := range podSpec.InitContainers {
		c := &podSpec.InitContainers[i]
		if c.Name == "upgrade-db" {
			podSpec.InitContainers = append(podSpec.InitContainers[:i], podSpec.InitContainers[i+1:]...)
		}
	}
	return nil
}

// SetEndpointsDeploymentReplicas updates the number of replicas on the endpoints deployment
func (r *Reconciler) SetEndpointsDeploymentReplicas(replicas int32) error {
	r.Logger.Infof("SetEndpointsDeploymentReplicas:: setting endpoints replica count to %d", replicas)
	return r.ReconcileObject(r.DeploymentEndpoint, func() error {
		r.DeploymentEndpoint.Spec.Replicas = &replicas
		return nil
	})
}

// SetDesiredCoreAppConfig initiate the config map with predifined environment variables and their values
func (r *Reconciler) SetDesiredCoreAppConfig() error {
	// Reowning the ConfigMap, incase the CreateOrUpdate removed the OwnerRefernce
	r.Own(r.CoreAppConfig)

	// When adding a new env var, make sure to add it to the sts/deployment as well
	// see "NOOBAA_DISABLE_COMPRESSION" as an example
	DefaultConfigMapData := map[string]string{
		"NOOBAA_DISABLE_COMPRESSION": "false",
		"DISABLE_DEV_RANDOM_SEED":    "true",
		"NOOBAA_LOG_LEVEL":           "default_level",
	}
	for key, value := range DefaultConfigMapData {
		if _, ok := r.CoreAppConfig.Data[key]; !ok {
			r.CoreAppConfig.Data[key] = value
		}
	}

	if r.CoreAppConfig.Annotations == nil {
		r.CoreAppConfig.Annotations = make(map[string]string)
	}
	r.CoreAppConfig.Annotations["noobaa.io/configmap-hash"] = util.GetCmDataHash(r.CoreAppConfig.Data)

	return nil
}

func (r *Reconciler) findLocalStorageClass() (string, error) {
	lsoStorageClassNames := []string{}
	scList := &storagev1.StorageClassList{}
	util.KubeList(scList)
	for _, sc := range scList.Items {
		if sc.ObjectMeta.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			return sc.Name, nil
		}
		if sc.Provisioner == "kubernetes.io/no-provisioner" {
			lsoStorageClassNames = append(lsoStorageClassNames, sc.Name)
		}
	}
	if len(lsoStorageClassNames) == 0 {
		return "", fmt.Errorf("Error: found no LSO storage class and no storage class was marked as default")
	}
	if len(lsoStorageClassNames) > 1 {
		return "", fmt.Errorf("Error: found more than one LSO storage class and none was marked as default")
	}
	return lsoStorageClassNames[0], nil
}
