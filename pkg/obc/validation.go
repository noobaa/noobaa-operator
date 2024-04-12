package obc

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
)

// ValidateOBC validate object bucket claim
func ValidateOBC(obc *nbv1.ObjectBucketClaim, isCLI bool) error {
	if obc == nil {
		return nil
	}
	return validateAdditionalConfig(obc.Name, obc.Spec.AdditionalConfig, false, isCLI)
}

// ValidateOB validate object bucket
func ValidateOB(ob *nbv1.ObjectBucket, isCLI bool) error {
	if ob == nil {
		return nil
	}
	return validateAdditionalConfig(ob.Name, ob.Spec.Endpoint.AdditionalConfigData, true, isCLI)
}

// Validate additional config
func validateAdditionalConfig(objectName string, additionalConfig map[string]string, update bool, isCLI bool) error {
	if additionalConfig == nil {
		return nil
	}

	obcMaxSize := additionalConfig["maxSize"]
	obcMaxObjects := additionalConfig["maxObjects"]
	replicationPolicy := additionalConfig["replicationPolicy"]
	NSFSAccountConfig := additionalConfig["NSFSAccountConfig"]

	if err := util.ValidateQuotaConfig(objectName, obcMaxSize, obcMaxObjects); err != nil {
		return err
	}

	if err := validations.ValidateReplicationPolicy(objectName, replicationPolicy, update, isCLI); err != nil {
		return err
	}

	if err := validations.ValidateAccountNSFSConfig(NSFSAccountConfig); err != nil {
		return err
	}

	return nil
}
