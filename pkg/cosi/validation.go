package cosi

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
)

// ValidateCOSIBucketClaim validate COSI bucket claim
func ValidateCOSIBucketClaim(objectName string, req *APIRequest) error {
	return validateAdditionalParameters(objectName, req)
}

// Validate additional parameters of the cosi bucket
func validateAdditionalParameters(bucketName string, req *APIRequest) error {
	placementPolicy := req.BucketClass.PlacementPolicy
	if err := validations.ValidatePlacementPolicy(placementPolicy, req.Provisioner.Namespace); err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("cosi bucket claim %q validation error: invalid placementPolicy %v, %v", bucketName, placementPolicy, err),
		}
	}

	namespacePolicy := req.BucketClass.NamespacePolicy
	if err := validations.ValidateNamespacePolicy(namespacePolicy, req.Provisioner.Namespace); err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("cosi bucket claim %q validation error: invalid namespacePolicy %v, %v", bucketName, namespacePolicy, err),
		}
	}

	replicationPolicy := req.BucketClass.ReplicationPolicy
	if err := validations.ValidateReplicationPolicy(bucketName, replicationPolicy, false); err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("cosi bucket claim %q validation error: invalid replicationPolicy %v, %v", bucketName, replicationPolicy, err),
		}
	}

	quota := req.BucketClass.Quota
	if err := validations.ValidateQuotaConfig(bucketName, quota); err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("cosi bucket claim %q validation error: invalid quota %v, %v", bucketName, quota, err),
		}
	}
	return nil
}
