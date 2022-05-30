package validations

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

// ValidateNoobaaDeletion check the existence of AllowNoobaaDeletion in Noobaa CR
func ValidateNoobaaDeletion(nb nbv1.NooBaa) error {
	if !nb.Spec.CleanupPolicy.AllowNoobaaDeletion {
		return util.ValidationError{
			Msg: "Noobaa cleanup policy is not set, blocking Noobaa deletion",
		}
	}
	return nil
}
