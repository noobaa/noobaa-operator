package obc

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
)

// ValidateOBC validate object bucket claim
func ValidateOBC(obc *nbv1.ObjectBucketClaim) error {
	if obc == nil {
		return nil
	}
	return validateAdditionalConfig(obc.Name, obc.Spec.AdditionalConfig, false)
}

// ValidateOB validate object bucket
func ValidateOB(ob *nbv1.ObjectBucket) error {
	if ob == nil {
		return nil
	}
	return validateAdditionalConfig(ob.Name, ob.Spec.Endpoint.AdditionalConfigData, true)
}

// Validate additional config
func validateAdditionalConfig(objectName string, additionalConfig map[string]string, update bool) error {
	if additionalConfig == nil {
		return nil
	}

	obcMaxSize := additionalConfig["maxSize"]
	obcMaxObjects := additionalConfig["maxObjects"]
	replicationPolicy := additionalConfig["replicationPolicy"]

	if err := util.ValidateQuotaConfig(objectName, obcMaxSize, obcMaxObjects); err != nil {
		return err
	}

	if err := validations.ValidateReplicationPolicy(objectName, replicationPolicy, update); err != nil {
		return err
	}

	return nil
}
