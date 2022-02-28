package validations

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

// ValidateRemoveNSFSConfig validates that the NSFS config was not removed
func ValidateRemoveNSFSConfig(na nbv1.NooBaaAccount, oldNa nbv1.NooBaaAccount) error {
	if oldNa.Spec.NsfsAccountConfig != nil && na.Spec.NsfsAccountConfig == nil {
		return util.ValidationError{
			Msg: "Removing the NsfsAccountConfig is unsupported",
		}
	}
	return nil
}

// ValidateNSFSConfig validates that the NSFS config files were set properly
func ValidateNSFSConfig(na nbv1.NooBaaAccount) error {
	nsfsConf := na.Spec.NsfsAccountConfig

	if nsfsConf == nil {
		return nil
	}

	//UID validation
	if nsfsConf.UID < 0 {
		return util.ValidationError{
			Msg: "UID must be a whole positive number",
		}
	}

	//GID validation
	if nsfsConf.GID < 0 {
		return util.ValidationError{
			Msg: "GID must be a whole positive number",
		}
	}

	return nil
}
