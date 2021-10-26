package bucketclass

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

// ValidateBucketClass bucket class
func ValidateBucketClass(bc *nbv1.BucketClass) error {
	if bc == nil {
		return nil
	}
	return validateQuotaConfig(bc.Name, bc.Spec.Quota)
}

// Validate quota config
func validateQuotaConfig(bcName string, quota *nbv1.Quota) error {
	if quota == nil {
		return nil
	}

	return util.ValidateQuotaConfig(bcName, quota.MaxSize, quota.MaxObjects)
}
