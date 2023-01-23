package system

import (
	"fmt"
	"os"

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

	err = CheckMongoURL(r.NooBaa)
	if err != nil {
		return util.NewPersistentError("InvalidMongoDbURL", fmt.Sprintf(`%s`, err))
	}
	// Validate the DefaultBackingStore Spec
	// nolint: staticcheck
	if r.NooBaa.Spec.DefaultBackingStoreSpec != nil {
		return util.NewPersistentError("Invalid, DefaultBackingStoreSpec was deprecated", fmt.Sprintf(`%s`, err))
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
	if r.JoinSecret.StringData["auth_token"] == "" {
		return util.NewPersistentError("InvalidJoinSecert",
			"JoinSecret is missing auth_token")
	}
	return nil
}
