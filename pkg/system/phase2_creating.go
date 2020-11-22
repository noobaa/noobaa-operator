package system

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	cloudcredsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcilePhaseCreating runs the reconcile phase
func (r *Reconciler) ReconcilePhaseCreating() error {

	r.SetPhase(
		nbv1.SystemPhaseCreating,
		"SystemPhaseCreating",
		"noobaa operator started phase 2/4 - \"Creating\"",
	)

	if err := r.ReconcileObject(r.ServiceAccount, r.SetDesiredServiceAccount); err != nil {
		return err
	}
	if err := r.ReconcilePhaseCreatingForMainClusters(); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceS3, r.SetDesiredServiceS3); err != nil {
		return err
	}
	if err := r.ReconcileObjectOptional(r.RouteS3, nil); err != nil {
		return err
	}
	// the credentials that are created by cloud-credentials-operator sometimes take time
	// to be valid (requests sometimes returns InvalidAccessKeyId for 1-2 minutes)
	// creating the credential request as early as possible to try and avoid it
	if err := r.ReconcileBackingStoreCredentials(); err != nil {
		r.Logger.Errorf("failed to create CredentialsRequest. will retry in phase 4. error: %v", err)
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

	// A failure to discover OAuth endpoints should not fail the entire reconcile phase.
	oAuthEndpoints, err := util.DiscoverOAuthEndpoints()
	if err != nil {
		r.Logger.Warnf("Discovery of OAuth endpoints failed, got: %v", err)
	}
	r.OAuthEndpoints = oAuthEndpoints

	if err := r.ReconcileObject(r.SecretServer, nil); err != nil {
		return err
	}
	if r.NooBaa.Spec.DBType == "postgres" {
		if err := r.ReconcileObject(r.SecretDB, nil); err != nil {
			return err
		}
	}
	if err := r.ReconcileObject(r.SecretRootMasterKey, nil); err != nil {
		return err
	}
	if err := r.UpgradeSplitDB(); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.CoreApp, r.SetDesiredCoreApp); err != nil {
		return err
	}
	if err := r.ReconcileDB(); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceMgmt, r.SetDesiredServiceMgmt); err != nil {
		return err
	}
	if r.NooBaa.Spec.DBType == "postgres" {
		if err := r.ReconcileObject(r.ServiceDbPg, r.SetDesiredServiceDB); err != nil {
			return err
		}
	} else {
		if err := r.ReconcileObject(r.ServiceDb, r.SetDesiredServiceDB); err != nil {
			return err
		}
	}
	if r.NooBaa.Spec.DBType == "postgres" {
		if err := r.UpgradeMigrateDB(); err != nil {
			return err
		}
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
	return nil
}

// SetDesiredServiceS3 updates the ServiceS3 as desired for reconciling
func (r *Reconciler) SetDesiredServiceS3() error {
	r.ServiceS3.Spec.Selector["noobaa-s3"] = r.Request.Name
	return nil
}

// SetDesiredServiceDB updates the ServiceS3 as desired for reconciling
func (r *Reconciler) SetDesiredServiceDB() error {
	if r.NooBaa.Spec.DBType == "postgres" {
		r.ServiceDbPg.Spec.Selector["noobaa-db"] = "postgres"
		r.ServiceDbPg.Spec.Ports[0].Name = "postgres"
		r.ServiceDbPg.Spec.Ports[0].Port = 5432
		r.ServiceDbPg.Spec.Ports[0].TargetPort = intstr.FromInt(5432)
	} else {
		r.ServiceDb.Spec.Selector["noobaa-db"] = r.Request.Name
		r.ServiceDb.Spec.Ports[0].Name = "mongodb"
		r.ServiceDb.Spec.Ports[0].Port = 27017
		r.ServiceDb.Spec.Ports[0].TargetPort = intstr.FromInt(27017)
	}
	return nil
}

// SetDesiredNooBaaDB updates the NooBaaDB as desired for reconciling
func (r *Reconciler) SetDesiredNooBaaDB() error {
	var NooBaaDB *appsv1.StatefulSet = nil
	if r.NooBaa.Spec.DBType == "postgres" {
		NooBaaDB = r.NooBaaPostgresDB
		NooBaaDB.Spec.Template.Labels["noobaa-db"] = "postgres"
		NooBaaDB.Spec.Selector.MatchLabels["noobaa-db"] = "postgres"
		NooBaaDB.Spec.ServiceName = r.ServiceDbPg.Name
	} else {
		NooBaaDB = r.NooBaaMongoDB
		NooBaaDB.Spec.Template.Labels["noobaa-db"] = r.Request.Name
		NooBaaDB.Spec.Selector.MatchLabels["noobaa-db"] = r.Request.Name
		NooBaaDB.Spec.ServiceName = r.ServiceDb.Name
	}

	podSpec := &NooBaaDB.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa"
	for i := range podSpec.InitContainers {
		c := &podSpec.InitContainers[i]
		if c.Name == "init" {
			c.Image = r.NooBaa.Status.ActualImage
		}
	}
	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		if c.Name == "db" {

			c.Image = options.DBImage
			if os.Getenv("NOOBAA_DB_IMAGE") != "" {
				c.Image = os.Getenv("NOOBAA_DB_IMAGE")
			}
			if r.NooBaa.Spec.DBImage != nil {
				c.Image = *r.NooBaa.Spec.DBImage
			}

			if r.NooBaa.Spec.DBResources != nil {
				c.Resources = *r.NooBaa.Spec.DBResources
			}
			if r.NooBaa.Spec.DBType == "postgres" {
				for j := range c.Env {
					switch c.Env[j].Name {
					case "POSTGRESQL_USER":
						c.Env[j].ValueFrom = &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: r.SecretDB.Name,
								},
								Key: "user",
							},
						}
					case "POSTGRESQL_PASSWORD":
						c.Env[j].ValueFrom = &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: r.SecretDB.Name,
								},
								Key: "password",
							},
						}
					}
					
				}
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
	if r.NooBaa.Spec.Tolerations != nil {
		podSpec.Tolerations = r.NooBaa.Spec.Tolerations
	}
	if r.NooBaa.Spec.Affinity != nil {
		podSpec.Affinity = r.NooBaa.Spec.Affinity
	}

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
				}
				if r.NooBaa.Spec.DBVolumeResources != nil {
					pvc.Spec.Resources = *r.NooBaa.Spec.DBVolumeResources
				}
			}
		}

	} else {

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
				}
				if desiredClass != currentClass {
					r.Recorder.Eventf(r.NooBaa, corev1.EventTypeWarning, "DBStorageClassIsImmutable",
						"spec.dbStorageClass is immutable and cannot be updated for volume %q in existing %s %q"+
							" since it requires volume recreate and migrate which is unsupported by the operator",
						pvc.Name, r.CoreApp.TypeMeta.Kind, r.CoreApp.Name)
				}
				if r.NooBaa.Spec.DBVolumeResources != nil &&
					!reflect.DeepEqual(pvc.Spec.Resources, *r.NooBaa.Spec.DBVolumeResources) {
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

		case "MONGODB_URL":
			c.Env[j].Value = "mongodb://" + r.NooBaaMongoDB.Name + "-0." + r.NooBaaMongoDB.Spec.ServiceName + "/nbcore"

		case "POSTGRES_HOST":
			c.Env[j].Value = r.NooBaaPostgresDB.Name + "-0." + r.NooBaaPostgresDB.Spec.ServiceName

		case "DB_TYPE":
			if r.NooBaa.Spec.DBType == "postgres" {
				c.Env[j].Value = "postgres"
			}

		case "OAUTH_AUTHORIZATION_ENDPOINT":
			if r.OAuthEndpoints != nil {
				c.Env[j].Value = r.OAuthEndpoints.AuthorizationEndpoint
			}

		case "OAUTH_TOKEN_ENDPOINT":
			if r.OAuthEndpoints != nil {
				c.Env[j].Value = r.OAuthEndpoints.TokenEndpoint
			}
		case "POSTGRES_USER":
			if r.NooBaa.Spec.DBType == "postgres" {
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
			if r.NooBaa.Spec.DBType == "postgres" {
				c.Env[j].ValueFrom = &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: r.SecretDB.Name,
						},
						Key: "password",
					},
				}
			}
		}
				
	}
}

// SetDesiredCoreApp updates the CoreApp as desired for reconciling
func (r *Reconciler) SetDesiredCoreApp() error {
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
			// adding the missing Env varibale from default container
			util.MergeEnvArrays(&c.Env, &r.DefaultCoreApp.Env);
			r.setDesiredCoreEnv(c)

			util.ReflectEnvVariable(&c.Env, "HTTP_PROXY")
			util.ReflectEnvVariable(&c.Env, "HTTPS_PROXY")
			util.ReflectEnvVariable(&c.Env, "NO_PROXY")

			if r.NooBaa.Spec.CoreResources != nil {
				c.Resources = *r.NooBaa.Spec.CoreResources
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
	if r.NooBaa.Spec.Tolerations != nil {
		podSpec.Tolerations = r.NooBaa.Spec.Tolerations
	}
	if r.NooBaa.Spec.Affinity != nil {
		podSpec.Affinity = r.NooBaa.Spec.Affinity
	}

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

	phase := r.NooBaa.Status.UpgradePhase
	replicas := int32(1);
	if (phase == nbv1.UpgradePhasePrepare || phase == nbv1.UpgradePhaseMigrate) {
		replicas = int32(0)
	}
	r.CoreApp.Spec.Replicas = &replicas
	return nil
}

// ReconcileBackingStoreCredentials creates a CredentialsRequest resource if necessary and returns
// the bucket name allowed for the credentials. nil is returned if cloud credentials are not supported
func (r *Reconciler) ReconcileBackingStoreCredentials() error {
	// Skip if joining another NooBaa
	if r.JoinSecret != nil {
		return nil
	}

	if util.IsAWSPlatform() {
		return r.ReconcileAWSCredentials()
	}
	if util.IsAzurePlatform() {
		return r.ReconcileAzureCredentials()
	}
	if util.IsGCPPlatform() {
		return r.ReconcileGCPCredentials()
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
	r.Logger.Info("Running in AWS. will create a CredentialsRequest resource")
	var bucketName string
	err := r.Client.Get(r.Ctx, util.ObjectKey(r.AWSCloudCreds), r.AWSCloudCreds)
	if err == nil {
		// credential request already exist. get the bucket name
		codec, err := cloudcredsv1.NewCodec()
		if err != nil {
			r.Logger.Error("error creating codec for cloud credentials providerSpec")
			return err
		}
		awsProviderSpec := &cloudcredsv1.AWSProviderSpec{}
		err = codec.DecodeProviderSpec(r.AWSCloudCreds.Spec.ProviderSpec, awsProviderSpec)
		if err != nil {
			r.Logger.Error("error decoding providerSpec from cloud credentials request")
			return err
		}
		bucketName = strings.TrimPrefix(awsProviderSpec.StatementEntries[0].Resource, "arn:aws:s3:::")
		r.Logger.Infof("found existing credential request for bucket %s", bucketName)
		r.DefaultBackingStore.Spec.AWSS3 = &nbv1.AWSS3Spec{
			TargetBucket: bucketName,
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
		codec, err := cloudcredsv1.NewCodec()
		if err != nil {
			r.Logger.Error("error creating codec for cloud credentials providerSpec")
			return err
		}
		awsProviderSpec := &cloudcredsv1.AWSProviderSpec{}
		err = codec.DecodeProviderSpec(r.AWSCloudCreds.Spec.ProviderSpec, awsProviderSpec)
		if err != nil {
			r.Logger.Error("error decoding providerSpec from cloud credentials request")
			return err
		}
		// fix creds request according to bucket name
		awsProviderSpec.StatementEntries[0].Resource = "arn:aws:s3:::" + bucketName
		awsProviderSpec.StatementEntries[1].Resource = "arn:aws:s3:::" + bucketName + "/*"
		updatedProviderSpec, err := codec.EncodeProviderSpec(awsProviderSpec)
		if err != nil {
			r.Logger.Error("error encoding providerSpec for cloud credentials request")
			return err
		}
		r.AWSCloudCreds.Spec.ProviderSpec = updatedProviderSpec
		r.Own(r.AWSCloudCreds)
		err = r.Client.Create(r.Ctx, r.AWSCloudCreds)
		if err != nil {
			r.Logger.Errorf("got error when trying to create credentials request for bucket %s. %v", bucketName, err)
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

// ReconcileDB choose between different types of DB
func (r *Reconciler) ReconcileDB() error {
	var err error
	if r.NooBaa.Spec.DBType == "postgres" {
		err = r.ReconcileObject(r.NooBaaPostgresDB, r.SetDesiredNooBaaDB)
		// Making sure that previous CRs without the value will deploy MongoDB
	} else if r.NooBaa.Spec.DBType == "" || r.NooBaa.Spec.DBType == "mongodb" {
		err = r.ReconcileObject(r.NooBaaMongoDB, r.SetDesiredNooBaaDB)
	} else {
		err = util.NewPersistentError("UnknownDBType", "Unknown dbType is specified in NooBaa spec")
	}
	return err
}

// UpgradeSplitDB removes the old pvc and create a  new one with the same PV
func (r *Reconciler) UpgradeSplitDB() error {
	oldPvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-noobaa-core-0",
			Namespace: options.Namespace,
		},
	}
	if util.KubeCheckQuiet(oldPvc) {
		r.Logger.Infof("UpgradeSplitDB: Old OVC found, upgrading...")
		if err := r.UpgradeSplitDBSetReclaimPolicy(oldPvc, corev1.PersistentVolumeReclaimRetain); err != nil {
			return err
		}
		if err := r.UpgradeSplitDBCreateNewPVC(oldPvc); err != nil {
			return err
		}
		if err := r.UpgradeSplitDBSetReclaimPolicy(oldPvc, corev1.PersistentVolumeReclaimDelete); err != nil {
			return err
		}
		if err := r.UpgradeSplitDBDeleteOldSTS(); err != nil {
			return err
		}
		if err := r.UpgradeSplitDBDeleteOldPVC(oldPvc); err != nil {
			return err
		}
	}
	return nil
}

// UpgradeSplitDBSetReclaimPolicy sets the reclaim policy to reclaim parameter and checks it
func (r *Reconciler) UpgradeSplitDBSetReclaimPolicy(oldPvc *corev1.PersistentVolumeClaim, reclaim corev1.PersistentVolumeReclaimPolicy) error {
	pv := &corev1.PersistentVolume{
		TypeMeta:   metav1.TypeMeta{Kind: "PersistentVolume"},
		ObjectMeta: metav1.ObjectMeta{Name: oldPvc.Spec.VolumeName},
	}
	if !util.KubeCheck(pv) {
		return fmt.Errorf("UpgradeSplitDBSetReclaimPolicy(%s): PV not found", reclaim)
	}
	if pv.Spec.PersistentVolumeReclaimPolicy != reclaim {
		pv.Spec.PersistentVolumeReclaimPolicy = reclaim
		if pv.Spec.ClaimRef != nil &&
			pv.Spec.ClaimRef.Name == oldPvc.Name &&
			pv.Spec.ClaimRef.Namespace == oldPvc.Namespace {
			pv.Spec.ClaimRef = nil
		}
		util.KubeUpdate(pv)
		if !util.KubeCheck(pv) {
			return fmt.Errorf("UpgradeSplitDBSetReclaimPolicy(%s): PV not found after update", reclaim)
		}
		if pv.Spec.PersistentVolumeReclaimPolicy != reclaim {
			return fmt.Errorf("UpgradeSplitDBSetReclaimPolicy(%s): PV reclaim policy could not be updated", reclaim)
		}
	}
	return nil
}

// UpgradeSplitDBCreateNewPVC creates new pvc and checks it
func (r *Reconciler) UpgradeSplitDBCreateNewPVC(oldPvc *corev1.PersistentVolumeClaim) error {
	newPvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-" + r.NooBaaMongoDB.Name + "-0",
			Namespace: options.Namespace,
		},
		Spec: oldPvc.Spec,
	}
	util.KubeCreateSkipExisting(newPvc)
	time.Sleep(2 * time.Second)
	if !util.KubeCheck(newPvc) {
		return fmt.Errorf("UpgradeSplitDBCreateNewPVC: New PVC not found")
	}
	if newPvc.Status.Phase != corev1.ClaimBound {
		return fmt.Errorf("UpgradeSplitDBCreateNewPVC: New PVC not bound yet")
	}
	if newPvc.Spec.VolumeName != oldPvc.Spec.VolumeName {
		// TODO how to recover?? since this is not expected maybe just return persistent error and wait for manual fix
		return fmt.Errorf("UpgradeSplitDBCreateNewPVC: New PVC bound to another PV")
	}
	return nil
}

// UpgradeSplitDBDeleteOldSTS deletes old STS named noobaa-core and checks it
func (r *Reconciler) UpgradeSplitDBDeleteOldSTS() error {
	oldSts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{Kind: "StatefulSet"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "noobaa-core",
			Namespace: options.Namespace,
		},
	}
	util.KubeDelete(oldSts)
	if util.KubeCheck(oldSts) {
		return fmt.Errorf("UpgradeSplitDBDeleteOldSTS: Old STS still exists")
	}
	return nil
}

// UpgradeSplitDBDeleteOldPVC deletes the parameter oldPvc and checks it
func (r *Reconciler) UpgradeSplitDBDeleteOldPVC(oldPVC *corev1.PersistentVolumeClaim) error {
	util.KubeDelete(oldPVC)
	if util.KubeCheck(oldPVC) {
		return fmt.Errorf("UpgradeSplitDBDeleteOldPVC: Old PVC still exists")
	}
	return nil
}

// UpgradeMigrateDB performs a db upgrade between mongodb to postgres
func (r *Reconciler) UpgradeMigrateDB() error {
	phase := r.NooBaa.Status.UpgradePhase
	if (phase == nbv1.UpgradePhaseFinished || phase == nbv1.UpgradePhaseNone) {
		return nil
	}
	oldSts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{Kind: "StatefulSet"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "noobaa-db",
			Namespace: options.Namespace,
		},
	}
	if !util.KubeCheck(oldSts) {
		phase = nbv1.UpgradePhaseNone
	} else {
		r.Logger.Infof("UpgradeMigrateDB: Old STS found, upgrading...")
		if (phase == "") {
			phase = nbv1.UpgradePhasePrepare
		}
		switch phase {
		case nbv1.UpgradePhasePrepare:
			util.KubeCheckQuiet(r.CoreApp)
			if (*r.CoreApp.Spec.Replicas == 0) {
				phase = nbv1.UpgradePhaseMigrate
			}
		case nbv1.UpgradePhaseMigrate: 
			if err:= r.ReconcileObject(r.UpgradeJob, r.SetDesiredJobUpgradeDB); err != nil {
				return err
			}
			if (r.UpgradeJob.Status.Succeeded > 0) {
				phase = nbv1.UpgradePhaseClean
			}
		case nbv1.UpgradePhaseClean: 
			if (*oldSts.Spec.Replicas == 0) {
				phase = nbv1.UpgradePhaseFinished
			} else {
				if err:= r.ReconcileObject(oldSts, func () error {
					replicas := int32(0); // cleaning the pods
					oldSts.Spec.Replicas = &replicas;
					return nil;
				}); err != nil {
					return err
				}
			}
		}
	}
	r.NooBaa.Status.UpgradePhase = phase;
	if err := r.UpdateStatus(); err != nil {
		return err
	} 
	return nil
}

// SetDesiredJobUpgradeDB updates the UpgradeJob as desired for reconciling
func (r *Reconciler) SetDesiredJobUpgradeDB() error {
	r.UpgradeJob.Spec.Template.Spec.Containers[0].Image = r.NooBaa.Status.ActualImage
	r.UpgradeJob.Spec.Template.Spec.Containers[0].Command = []string{"/noobaa_init_files/noobaa_init.sh", "db_migrate"}
	r.setDesiredCoreEnv(&r.UpgradeJob.Spec.Template.Spec.Containers[0])
	return nil
}