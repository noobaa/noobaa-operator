package system

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/util"
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

	if err := r.ReconcileObject(r.PrometheusRule, nil); err != nil {
		if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
			r.Logger.Printf("No PrometheusRule CRD existing, skip creating PrometheusRules\n")
		} else {
			return err
		}
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
	podSpec.ServiceAccountName = "noobaa-operator" // TODO do we use the same SA?
	for i := range podSpec.InitContainers {
		c := &podSpec.InitContainers[i]
		if c.Name == "init-mongo" {
			c.Image = r.NooBaa.Status.ActualImage
		}
	}
	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		if c.Name == "noobaa-server" {
			c.Image = r.NooBaa.Status.ActualImage
			for j := range c.Env {
				if c.Env[j].Name == "AGENT_PROFILE" {
					c.Env[j].Value = fmt.Sprintf(`{ "image": "%s" }`, r.NooBaa.Status.ActualImage)
				}
			}
			if r.NooBaa.Spec.CoreResources != nil {
				c.Resources = *r.NooBaa.Spec.CoreResources
			}
		} else if c.Name == "mongodb" {
			if r.NooBaa.Spec.MongoImage == nil {
				c.Image = options.MongoImage
			} else {
				c.Image = *r.NooBaa.Spec.MongoImage
			}
			if r.NooBaa.Spec.MongoResources != nil {
				c.Resources = *r.NooBaa.Spec.MongoResources
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

	for i := range r.CoreApp.Spec.VolumeClaimTemplates {
		pvc := &r.CoreApp.Spec.VolumeClaimTemplates[i]
		pvc.Spec.StorageClassName = r.NooBaa.Spec.StorageClassName

		// TODO we want to own the PVC's by NooBaa system but get errors on openshift:
		//   Warning  FailedCreate  56s  statefulset-controller
		//   create Pod noobaa-core-0 in StatefulSet noobaa-core failed error:
		//   Failed to create PVC mongo-datadir-noobaa-core-0:
		//   persistentvolumeclaims "mongo-datadir-noobaa-core-0" is forbidden:
		//   cannot set blockOwnerDeletion if an ownerReference refers to a resource
		//   you can't set finalizers on: , <nil>, ...

		// r.Own(pvc)
	}
}

// ReconcileCredentialsRequest creates a CredentialsRequest resource if neccesary and returns
// the bucket name allowed for the credentials. nil is returned if cloud credentials are not supported
func (r *Reconciler) ReconcileCredentialsRequest() error {
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
		r.DefaultBackingStore.Spec.S3Options = &nbv1.S3Options{
			BucketName: bucketName,
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
		r.DefaultBackingStore.Spec.S3Options = &nbv1.S3Options{
			BucketName: bucketName,
		}
		return nil
	}
	return err
}
