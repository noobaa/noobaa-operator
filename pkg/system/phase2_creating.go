package system

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcilePhaseCreating runs the reconcile phase
func (r *Reconciler) ReconcilePhaseCreating() error {

	r.SetPhase(
		nbv1.SystemPhaseCreating,
		"SystemPhaseCreating",
		"noobaa operator started phase 2/4 - \"Creating\"",
	)

	// A failure to discover OAuth endpoints should not fail the entire reconcile phase.
	oAuthEndpoints, err := util.DiscoverOAuthEndpoints()
	if err != nil {
		r.Logger.Warnf("Discovery of OAuth endpoints failed, got: %v", err)
	}
	r.OAuthEndpoints = oAuthEndpoints
	// the credentials that are created by cloud-credentials-operator sometimes take time
	// to be valid (requests sometimes returns InvalidAccessKeyId for 1-2 minutes)
	// creating the credential request as early as possible to try and avoid it
	if err := r.ReconcileBackingStoreCredentials(); err != nil {
		r.Logger.Errorf("failed to create CredentialsRequest. will retry in phase 4. error: %v", err)
		return err
	}
	if err := r.ReconcileObject(r.ServiceAccount, r.SetDesiredServiceAccount); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.SecretServer, nil); err != nil {
		return err
	}

	if err := r.UpgradeSplitDB(); err != nil {
		return err
	}

	if err := r.ReconcileObject(r.CoreApp, r.SetDesiredCoreApp); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.NooBaaDB, r.SetDesiredNooBaaDB); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceMgmt, r.SetDesiredServiceMgmt); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceDb, r.SetDesiredServiceDB); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceS3, r.SetDesiredServiceS3); err != nil {
		return err
	}
	if err := r.ReconcileObjectOptional(r.RouteMgmt, nil); err != nil {
		return err
	}
	if err := r.ReconcileObjectOptional(r.RouteS3, nil); err != nil {
		return err
	}

	return nil
}

// SetDesiredServiceAccount updates the ServiceAccount as desired for reconciling
func (r *Reconciler) SetDesiredServiceAccount() {
	if r.ServiceAccount.Annotations == nil {
		r.ServiceAccount.Annotations = map[string]string{}
	}
	r.ServiceAccount.Annotations["serviceaccounts.openshift.io/oauth-redirectreference.noobaa-mgmt"] =
		`{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"` + r.RouteMgmt.Name + `"}}`
}

// SetDesiredServiceMgmt updates the ServiceMgmt as desired for reconciling
func (r *Reconciler) SetDesiredServiceMgmt() {
	r.ServiceMgmt.Spec.Selector["noobaa-mgmt"] = r.Request.Name
}

// SetDesiredServiceS3 updates the ServiceS3 as desired for reconciling
func (r *Reconciler) SetDesiredServiceS3() {
	r.ServiceS3.Spec.Selector["noobaa-s3"] = r.Request.Name
}

// SetDesiredServiceDB updates the ServiceS3 as desired for reconciling
func (r *Reconciler) SetDesiredServiceDB() {
	r.ServiceDb.Spec.Selector["noobaa-db"] = r.Request.Name
}

// SetDesiredNooBaaDB updates the NooBaaDB as desired for reconciling
func (r *Reconciler) SetDesiredNooBaaDB() {
	r.NooBaaDB.Spec.Template.Labels["noobaa-db"] = r.Request.Name
	r.NooBaaDB.Spec.Selector.MatchLabels["noobaa-db"] = r.Request.Name
	r.NooBaaDB.Spec.ServiceName = r.ServiceDb.Name

	podSpec := &r.NooBaaDB.Spec.Template.Spec
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

	if r.NooBaaDB.UID == "" {
		for i := range r.NooBaaDB.Spec.VolumeClaimTemplates {
			pvc := &r.NooBaaDB.Spec.VolumeClaimTemplates[i]
			r.Own(pvc)
			// unsetting BlockOwnerDeletion to acoid error when trying to own pvc:
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
		for i := range r.NooBaaDB.Spec.VolumeClaimTemplates {
			pvc := &r.NooBaaDB.Spec.VolumeClaimTemplates[i]
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
}

// SetDesiredCoreApp updates the CoreApp as desired for reconciling
func (r *Reconciler) SetDesiredCoreApp() {
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
			for j := range c.Env {
				switch c.Env[j].Name {
				// case "ENDPOINT_FORKS_NUMBER":
				// 	c.Env[j].Value = "1" // TODO recalculate
				case "AGENT_PROFILE":
					c.Env[j].Value = r.SetDesiredAgentProfile(c.Env[j].Value)

				case "MONGODB_URL":
					c.Env[j].Value = "mongodb://" + r.NooBaaDB.Name + "-0." + r.NooBaaDB.Spec.ServiceName + "/nbcore"

				case "OAUTH_AUTHORIZATION_ENDPOINT":
					if r.OAuthEndpoints != nil {
						c.Env[j].Value = r.OAuthEndpoints.AuthorizationEndpoint
					}

				case "OAUTH_TOKEN_ENDPOINT":
					if r.OAuthEndpoints != nil {
						c.Env[j].Value = r.OAuthEndpoints.TokenEndpoint
					}
				}
			}
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
}

// ReconcileBackingStoreCredentials creates a CredentialsRequest resource if neccesary and returns
// the bucket name allowed for the credentials. nil is returned if cloud credentials are not supported
func (r *Reconciler) ReconcileBackingStoreCredentials() error {
	if util.IsAWSPlatform() {
		return r.ReconcileAWSCredentials()
	}
	return r.ReconcileRGWCredentials()
}

// ReconcileRGWCredentials creates a ceph objectstore user if a ceph objectstore exists in the same namespace
func (r *Reconciler) ReconcileRGWCredentials() error {
	r.Logger.Info("Not running in AWS. will attempt to create a ceph objectstore user")
	util.KubeCheck(r.CephObjectstoreUser)
	if r.CephObjectstoreUser.UID != "" {
		return nil
	}

	// create user if not already exists
	// list ceph objectstores and pick the first one
	cephObjectStoresList := &cephv1.CephObjectStoreList{}
	if !util.KubeList(cephObjectStoresList, &client.ListOptions{Namespace: options.Namespace}) {
		r.Logger.Warn("failed to list ceph objectstore to use as backing store")
		// no object stores
		return nil
	}
	if len(cephObjectStoresList.Items) == 0 {
		r.Logger.Warn("did not find any ceph objectstore to use as backing store")
		// no object stores
		return nil
	}
	r.Logger.Infof("found %d ceph objectstores: %v", len(cephObjectStoresList.Items), cephObjectStoresList.Items)
	// for now take the first one. need to decide what to do if multiple objectstores in one namespace
	storeName := cephObjectStoresList.Items[0].ObjectMeta.Name
	r.Logger.Infof("using objectstore %q as a default backing store", storeName)
	r.CephObjectstoreUser.Spec.Store = storeName

	r.Own(r.CephObjectstoreUser)
	// create ceph objectstore user
	err := r.Client.Create(r.Ctx, r.CephObjectstoreUser)
	if err != nil {
		r.Logger.Errorf("got error on CephObjectstoreUser creation. error: %v", err)
		return err
	}
	return nil
}

// ReconcileAWSCredentials creates a CredentialsRequest resource if cloud credentials operator is available
func (r *Reconciler) ReconcileAWSCredentials() error {
	r.Logger.Info("Running in AWS. will create a CredentialsRequest resource")
	var bucketName string
	err := r.Client.Get(r.Ctx, util.ObjectKey(r.CloudCreds), r.CloudCreds)
	if err == nil {
		// credential request alread exist. get the bucket name
		codec, err := cloudcredsv1.NewCodec()
		if err != nil {
			r.Logger.Error("error creating codec for cloud credentials providerSpec")
			return err
		}
		awsProviderSpec := &cloudcredsv1.AWSProviderSpec{}
		err = codec.DecodeProviderSpec(r.CloudCreds.Spec.ProviderSpec, awsProviderSpec)
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
		bucketName = "noobaa-backing-store-" + uuid.New().String()
		codec, err := cloudcredsv1.NewCodec()
		if err != nil {
			r.Logger.Error("error creating codec for cloud credentials providerSpec")
			return err
		}
		awsProviderSpec := &cloudcredsv1.AWSProviderSpec{}
		err = codec.DecodeProviderSpec(r.CloudCreds.Spec.ProviderSpec, awsProviderSpec)
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
		r.CloudCreds.Spec.ProviderSpec = updatedProviderSpec
		r.Own(r.CloudCreds)
		err = r.Client.Create(r.Ctx, r.CloudCreds)
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

// UpgradeSplitDB removes the old pvc and create a  new one with the same PV
func (r *Reconciler) UpgradeSplitDB() error {
	oldPvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-noobaa-core-0",
			Namespace: options.Namespace,
		},
	}
	if util.KubeCheck(oldPvc) {
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
			Name:      "db-" + r.NooBaaDB.Name + "-0",
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
