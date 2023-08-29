package cosi

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
)

// ValidateCOSIBucketClaim validate COSI bucket claim
func ValidateCOSIBucketClaim(objectName string, namespace string, spec nbv1.BucketClassSpec, isCLI bool) error {
	return validateAdditionalParameters(objectName, namespace, spec, isCLI)
}

// Validate additional parameters of the cosi bucket
func validateAdditionalParameters(bucketName string, namespace string, spec nbv1.BucketClassSpec, isCLI bool) error {
	placementPolicy := spec.PlacementPolicy
	if err := validations.ValidatePlacementPolicy(placementPolicy, namespace); err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("cosi bucket claim %q validation error: invalid placementPolicy %v, %v", bucketName, placementPolicy, err),
		}
	}

	namespacePolicy := spec.NamespacePolicy
	if err := validations.ValidateNamespacePolicy(namespacePolicy, namespace); err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("cosi bucket claim %q validation error: invalid namespacePolicy %v, %v", bucketName, namespacePolicy, err),
		}
	}

	replicationPolicy := spec.ReplicationPolicy
	if err := validations.ValidateReplicationPolicy(bucketName, replicationPolicy, false, isCLI); err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("cosi bucket claim %q validation error: invalid replicationPolicy %v, %v", bucketName, replicationPolicy, err),
		}
	}

	quota := spec.Quota
	if err := validations.ValidateQuotaConfig(bucketName, quota); err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("cosi bucket claim %q validation error: invalid quota %v, %v", bucketName, quota, err),
		}
	}
	return nil
}
