package system

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/google/uuid"
	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	cloudcredsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// ReconcilePhaseCreating runs the reconcile phase
func (r *Reconciler) ReconcilePhaseCreating() error {

	r.SetPhase(
		nbv1.SystemPhaseCreating,
		"SystemPhaseCreating",
		"noobaa operator started phase 2/4 - \"Creating\"",
	)

	// the credentials that are created by cloud-credentials-operator sometimes take time
	// to be valid (requests sometimes returns InvalidAccessKeyId for 1-2 minutes)
	// creating the credential request as early as possible to try and avoid it
	err := r.ReconcileCredentialsRequest()
	if err != nil {
		r.Logger.Errorf("failed to create CredentialsRequest. will retry in phase 4. error: %v", err)
		return err
	}

	r.SecretServer.StringData["jwt"] = util.RandomBase64(16)
	r.SecretServer.StringData["server_secret"] = util.RandomHex(4)
	if err := r.ReconcileObject(r.SecretServer, nil); err != nil {
		return err
	}

	if err := r.ReconcileObject(r.CoreApp, r.SetDesiredCoreApp); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceMgmt, r.SetDesiredServiceMgmt); err != nil {
		return err
	}
	if err := r.ReconcileObject(r.ServiceS3, r.SetDesiredServiceS3); err != nil {
		return err
	}

	return nil
}

// SetDesiredServiceMgmt updates the ServiceMgmt as desired for reconciling
func (r *Reconciler) SetDesiredServiceMgmt() {
	r.ServiceMgmt.Spec.Selector["noobaa-mgmt"] = r.Request.Name
}

// SetDesiredServiceS3 updates the ServiceS3 as desired for reconciling
func (r *Reconciler) SetDesiredServiceS3() {
	r.ServiceS3.Spec.Selector["noobaa-s3"] = r.Request.Name
}

// SetDesiredCoreApp updates the CoreApp as desired for reconciling
func (r *Reconciler) SetDesiredCoreApp() {
	r.CoreApp.Spec.Template.Labels["noobaa-core"] = r.Request.Name
	r.CoreApp.Spec.Template.Labels["noobaa-mgmt"] = r.Request.Name
	r.CoreApp.Spec.Template.Labels["noobaa-s3"] = r.Request.Name
	r.CoreApp.Spec.Selector.MatchLabels["noobaa-core"] = r.Request.Name
	r.CoreApp.Spec.ServiceName = r.ServiceMgmt.Name

	podSpec := &r.CoreApp.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa"
	for i := range podSpec.InitContainers {
		c := &podSpec.InitContainers[i]
		switch c.Name {
		case "init":
			c.Image = r.NooBaa.Status.ActualImage
		}
	}
	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		switch c.Name {
		case "core":
			c.Image = r.NooBaa.Status.ActualImage
			for j := range c.Env {
				switch c.Env[j].Name {
				// case "ENDPOINT_FORKS_NUMBER":
				// 	c.Env[j].Value = "1" // TODO recalculate
				case "AGENT_PROFILE":
					c.Env[j].Value = r.SetDesiredAgentProfile(c.Env[j].Value)
				}
			}
			if r.NooBaa.Spec.CoreResources != nil {
				c.Resources = *r.NooBaa.Spec.CoreResources
			}
		case "db":
			if r.NooBaa.Spec.DBImage == nil {
				c.Image = options.DBImage
			} else {
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

	if r.CoreApp.UID == "" {
		for i := range r.CoreApp.Spec.VolumeClaimTemplates {
			pvc := &r.CoreApp.Spec.VolumeClaimTemplates[i]
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
		for i := range r.CoreApp.Spec.VolumeClaimTemplates {
			pvc := &r.CoreApp.Spec.VolumeClaimTemplates[i]
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

// ReconcileCredentialsRequest creates a CredentialsRequest resource if neccesary and returns
// the bucket name allowed for the credentials. nil is returned if cloud credentials are not supported
func (r *Reconciler) ReconcileCredentialsRequest() error {
	if !util.IsAWSPlatform() {
		r.Logger.Info("not running in AWS. skipping ReconcileCredentialsRequest")
		return nil
	}
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
