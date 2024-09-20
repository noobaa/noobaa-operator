package system

import (
	"database/sql"
	"fmt"
	"os"

	// this is the driver we are using for psql
	_ "github.com/lib/pq"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/asaskevich/govalidator"
	semver "github.com/coreos/go-semver/semver"
	dockerref "github.com/docker/distribution/reference"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

// ReconcilePhaseVerifying runs the reconcile verify phase
func (r *Reconciler) ReconcilePhaseVerifying() error {

	r.SetPhase(
		nbv1.SystemPhaseVerifying,
		"SystemPhaseVerifying",
		"noobaa operator started phase 1/4 - \"Verifying\"",
	)

	if err := r.CheckSystemCR(); err != nil {
		return err
	}

	if r.JoinSecret != nil {
		if err := r.CheckJoinSecret(); err != nil {
			return err
		}
	}

	if r.ExternalPgSecret != nil {
		if r.ExternalPgSecret.StringData["db_url"] == "" {
			return util.NewPersistentError("InvalidExternalPgSecert",
				"ExternalPgSecret is missing db_url")
		}
		if r.ExternalPgSSLSecret != nil {
			if r.ExternalPgSSLSecret.StringData["tls.key"] == "" ||
				r.ExternalPgSSLSecret.StringData["tls.crt"] == "" {
				return util.NewPersistentError("InvalidExternalPgCert",
					fmt.Sprintf("%q is missing private key (must be tls.key)"+
						" or missing cert key (must be tls.cert)", r.ExternalPgSSLSecret.Name))
			}
			err := os.WriteFile("/tmp/tls.key", []byte(r.ExternalPgSSLSecret.StringData["tls.key"]), 0600)
			if err != nil {
				return fmt.Errorf("failed to write k8s secret tls.key content to a file %v", err)
			}
			err = os.WriteFile("/tmp/tls.crt", []byte(r.ExternalPgSSLSecret.StringData["tls.crt"]), 0644)
			if err != nil {
				return fmt.Errorf("failed to write k8s secret tls.key content to a file %v", err)
			}
			os.Setenv("PGSSLKEY", "/tmp/tls.key")
			os.Setenv("PGSSLCERT", "/tmp/tls.crt")
		}
		if err := r.checkExternalPg(r.ExternalPgSecret.StringData["db_url"]); err != nil {
			return err
		}
	}

	if r.NooBaa.Spec.BucketLogging.LoggingType == nbv1.BucketLoggingTypeGuaranteed {
		if err := r.checkPersistentLoggingPVC(r.NooBaa.Spec.BucketLogging.BucketLoggingPVC, r.BucketLoggingPVC, "InvalidBucketLoggingConfiguration"); err != nil {
			return err
		}
	}

	if r.NooBaa.Spec.BucketNotifications.Enabled {
		if err := r.checkPersistentLoggingPVC(r.NooBaa.Spec.BucketNotifications.PVC, r.BucketNotificationsPVC, "InvalidBucketNotificationConfiguration"); err != nil {
			return err
		}
	}

	return nil
}

// CheckSystemCR checks the validity of the system CR
// (i.e system.metadata.name and system.spec.image)
// and updates the status accordingly
func (r *Reconciler) CheckSystemCR() error {

	log := r.Logger.WithField("func", "CheckSystemCR")

	// we assume a single system per ns here
	if r.NooBaa.Name != options.SystemName {
		return util.NewPersistentError("InvalidSystemName",
			fmt.Sprintf("Invalid system name %q expected %q", r.NooBaa.Name, options.SystemName))
	}

	specImage := options.ContainerImage
	if os.Getenv("NOOBAA_CORE_IMAGE") != "" {
		specImage = os.Getenv("NOOBAA_CORE_IMAGE")
	}
	if r.NooBaa.Spec.Image != nil {
		specImage = *r.NooBaa.Spec.Image
	}

	// Parse the image spec as a docker image url
	imageRef, err := dockerref.Parse(specImage)

	// If the image cannot be parsed log the incident and mark as persistent error
	// since we don't need to retry until the spec is updated.
	if err != nil {
		return util.NewPersistentError("InvalidImage",
			fmt.Sprintf(`Invalid image requested %q %v`, specImage, err))
	}

	// Get the image name and tag
	imageName := ""
	imageTag := ""
	switch image := imageRef.(type) {
	case dockerref.NamedTagged:
		log.Infof("Parsed image (NamedTagged) %v", image)
		imageName = image.Name()
		imageTag = image.Tag()
	case dockerref.Tagged:
		log.Infof("Parsed image (Tagged) %v", image)
		imageTag = image.Tag()
	case dockerref.Named:
		log.Infof("Parsed image (Named) %v", image)
		imageName = image.Name()
	default:
		log.Infof("Parsed image (unstructured) %v", image)
	}

	if imageName == options.ContainerImageName {
		version, err := semver.NewVersion(imageTag)
		if err == nil {
			log.Infof("Parsed version %q from image tag %q", version.String(), imageTag)
			minver := semver.New(options.ContainerImageSemverLowerBound)
			maxver := semver.New(options.ContainerImageSemverUpperBound)
			if version.Compare(*minver) != 1 || version.Compare(*maxver) != -1 {
				return util.NewPersistentError("InvalidImageVersion",
					fmt.Sprintf(`Invalid image version %q not matching version constraints: >=%v, <%v`,
						imageRef, minver, maxver))
			}
		} else {
			log.Infof("Using custom image %q", imageRef.String())
		}
	} else {
		log.Infof("Using custom image name %q the default is %q", imageRef.String(), options.ContainerImageName)
	}

	// Set ActualImage to be updated in the noobaa status
	r.NooBaa.Status.ActualImage = specImage

	// Verify the endpoints spec
	endpointsSpec := r.NooBaa.Spec.Endpoints
	if endpointsSpec != nil {
		if endpointsSpec.MinCount <= 0 {
			return util.NewPersistentError("InvalidEndpointsConfiguration",
				"Invalid endpoint min count (must be greater than 0)")
		}
		// Validate bounds on endpoint counts
		if endpointsSpec.MinCount > endpointsSpec.MaxCount {
			return util.NewPersistentError("InvalidEndpointsConfiguration",
				"Invalid endpoint maximum count (must be higher than or equal to minimum count)")
		}

		// Validate that all virtual hosts are in FQDN format
		for _, virtualHost := range endpointsSpec.AdditionalVirtualHosts {
			if !govalidator.IsDNSName(virtualHost) {
				return util.NewPersistentError("InvalidEndpointsConfiguration",
					fmt.Sprintf(`Invalid virtual host %s, not a fully qualified DNS name`, virtualHost))
			}
		}
	}

	// Validate the DefaultBackingStore Spec
	// nolint: staticcheck
	if r.NooBaa.Spec.DefaultBackingStoreSpec != nil {
		return util.NewPersistentError("Invalid, DefaultBackingStoreSpec was deprecated", fmt.Sprintf(`%s`, err))
	}

	if (r.NooBaa.Spec.Autoscaler.AutoscalerType == nbv1.AutoscalerTypeKeda ||
		r.NooBaa.Spec.Autoscaler.AutoscalerType == nbv1.AutoscalerTypeHPAV2) &&
		r.NooBaa.Spec.Autoscaler.PrometheusNamespace == "" {
		return util.NewPersistentError("InvalidEndpointsAutoscalerConfiguration",
			fmt.Sprintf("Autoscaler %s missing prometheusNamespace property ", r.NooBaa.Spec.Autoscaler.AutoscalerType))
	}

	return nil
}

// CheckJoinSecret checks that all need information to allow to join
// another noobaa clauster is specified in the join secret
func (r *Reconciler) CheckJoinSecret() error {
	if r.JoinSecret.StringData["mgmt_addr"] == "" {
		return util.NewPersistentError("InvalidJoinSecert",
			"JoinSecret is missing mgmt_addr")
	}
	if r.JoinSecret.StringData["auth_token"] == "" {
		return util.NewPersistentError("InvalidJoinSecert",
			"JoinSecret is missing auth_token")
	}

	if !util.IsRemoteClientNoobaa(r.NooBaa.GetAnnotations()) {
		if r.JoinSecret.StringData["bg_addr"] == "" {
			return util.NewPersistentError("InvalidJoinSecert",
				"JoinSecret is missing bg_addr")
		}
		if r.JoinSecret.StringData["md_addr"] == "" {
			return util.NewPersistentError("InvalidJoinSecert",
				"JoinSecret is missing md_addr")
		}
		if r.JoinSecret.StringData["hosted_agents_addr"] == "" {
			return util.NewPersistentError("InvalidJoinSecert",
				"JoinSecret is missing hosted_agents_addr")
		}
	}
	return nil
}

func (r *Reconciler) checkExternalPg(postgresDbURL string) error {
	dbURL := r.ExternalPgSecret.StringData["db_url"]
	if r.NooBaa.Spec.ExternalPgSSLRequired {
		if !r.NooBaa.Spec.ExternalPgSSLUnauthorized {
			dbURL += "?sslmode=verify-full" // won't allow self-signed certs
		} // when we want to allow self-signed we will use the default sslmode=require
	} else {
		dbURL += "?sslmode=disable" // don't use ssl - the default is to use it
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return util.NewPersistentError("InvalidExternalPgUrl",
			fmt.Sprintf("failed openning a connection to external DB url: %q, error: %s",
				dbURL, err))
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		return util.NewPersistentError("InvalidExternalPgUrl",
			fmt.Sprintf("failed pinging external DB url: %q, error: %s",
				dbURL, err))
	}
	// Query the PostgreSQL version
	var version string
	err = db.QueryRow("SELECT current_setting('server_version_num')::integer / 10000").Scan(&version)
	if err != nil {
		return util.NewPersistentError("InvalidExternalPgVersion",
			fmt.Sprintf("failed getting version of external DB url: %q, error: %s",
				dbURL, err))
	}
	// Check if the version is 15
	if version != "15" {
		return util.NewPersistentError("InvalidExternalPgVersion",
			fmt.Sprintf("version of external DB %q, is not supported: %s",
				dbURL, version))
	}
	// Query the database's collation
	var collation string
	err = db.QueryRow("SELECT datcollate FROM pg_database WHERE datname = current_database()").Scan(&collation)
	if err != nil {
		return util.NewPersistentError("InvalidExternalPgCollation",
			fmt.Sprintf("failed getting database collation of external DB url: %q, error: %s",
				dbURL, err))
	}
	// Check if the collation is "C"
	if collation != "C" {
		return util.NewPersistentError("InvalidExternalPgCollation",
			fmt.Sprintf("collation of external DB url: %q, is not supported: %s",
				dbURL, collation))
	}
	return nil
}

// checkPersistentLoggingPVC validates the configuration of pvc for persistent logging
func (r *Reconciler) checkPersistentLoggingPVC(
	pvcName *string,
	pvc *corev1.PersistentVolumeClaim,
	errorName string) error {
	if pvcName == nil {
		sc := &storagev1.StorageClass{
			TypeMeta:   metav1.TypeMeta{Kind: "StorageClass"},
			ObjectMeta: metav1.ObjectMeta{Name: "ocs-storagecluster-cephfs"},
		}
		// Return nil as the operator running in ODF environment
		if util.KubeCheck(sc) {
			return nil
		}
		return util.NewPersistentError(errorName,
			"Persistent Volume Claim (PVC) was not specified (and CephFS was not found for a defualt PVC)")
	}

	// Check if pvc exists in the cluster
	PersistentLoggingPVC := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{Kind: "PersistenVolumeClaim"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      *pvcName,
			Namespace: r.Request.Namespace,
		},
	}
	if !util.KubeCheck(PersistentLoggingPVC) {
		return util.NewPersistentError(errorName,
			fmt.Sprintf("The specified persistent logging pvc '%s' was not found", *pvcName))
	}

	// Check if pvc supports RWX access mode
	for _, accessMode := range PersistentLoggingPVC.Spec.AccessModes {
		if accessMode == corev1.ReadWriteMany {
			return nil
		}
	}
	return util.NewPersistentError(errorName,
		fmt.Sprintf("The specified persistent logging pvc '%s' does not support RWX access mode", *pvcName))
}
