package validations

import (
	"encoding/json"
	"fmt"
	"regexp"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

var linuxUsernameRegex = regexp.MustCompile(`^[^-:\s][^\s:]{0,30}[^-:\s]$`)

// ValidateNSFSAccountConfig validates that the provided NSFS config is valid
func ValidateNSFSAccountConfig(NSFSConfig string) error {
	log := util.Logger()

	if NSFSConfig == "" {
		return nil
	}

	var configObj nbv1.AccountNsfsConfig

	err := json.Unmarshal([]byte(NSFSConfig), &configObj)
	if err != nil {
		return fmt.Errorf("failed to parse NSFS account config %q: %v", NSFSConfig, err)
	}

	log.Infof("Validating NSFS config: %+v", NSFSConfig)
	// Check if no UID, GID or distinguished name were provided
	if configObj.UID == nil && configObj.GID == nil && configObj.DistinguishedName == "" {
		return fmt.Errorf("UID and GID, or DistinguishedName must be provided")
	}
	// Check UID/GID cases only in case they're defined
	if configObj.UID != nil || configObj.GID != nil {
		// Check whether only UID or only GID were provided
		if *configObj.UID < 0 || *configObj.GID < 0 {
			return fmt.Errorf("UID and GID must be positive integers")
		// Check whether a distinguished name was provided alongside UID or GID
		} else if configObj.DistinguishedName != "" && (*configObj.GID > -1 || *configObj.UID > -1) {
			return fmt.Errorf(`NSFS account config cannot include both distinguished name and UID/GID`)
		}
	// Otherwise, validate the distinguished name
	} else if configObj.DistinguishedName != "" {
		if !linuxUsernameRegex.MatchString(configObj.DistinguishedName) {
			return fmt.Errorf("DistinguishedName must be a valid username by Linux standards")
		}
	}

	return nil
}

// ValidateReplicationPolicy validates and replication params and returns the replication policy object
func ValidateReplicationPolicy(bucketName string, replicationPolicy string, update bool, isCLI bool) error {
	log := util.Logger()
	if replicationPolicy == "" {
		return nil
	}

	var replicationRules nb.ReplicationPolicy
	err := json.Unmarshal([]byte(replicationPolicy), &replicationRules)
	if err != nil {
		return fmt.Errorf("Failed to parse replication json %q: %v", replicationRules, err)
	}
	log.Infof("ValidateReplicationPolicy: newReplication %+v", replicationRules)

	if len(replicationRules.Rules) == 0 {
		if update {
			return nil
		}
		return fmt.Errorf("replication rules array of bucket %q is empty %q", bucketName, replicationRules)
	}

	replicationParams := &nb.BucketReplicationParams{
		Name:              bucketName,
		ReplicationPolicy: replicationRules,
	}

	log.Infof("ValidateReplicationPolicy: validating replication: replicationParams: %+v", replicationParams)
	IsExternalRPCConnection := false
	if util.IsTestEnv() || isCLI {
		IsExternalRPCConnection = true
	}
	sysClient, err := system.Connect(IsExternalRPCConnection)
	if err != nil {
		return fmt.Errorf("Provisioner Failed to validate replication of bucket %q with error: %v", bucketName, err)
	}

	err = sysClient.NBClient.ValidateReplicationAPI(*replicationParams)
	if err != nil {
		rpcErr, isRPCErr := err.(*nb.RPCError)
		if isRPCErr {
			if rpcErr.RPCCode == "INVALID_REPLICATION_POLICY" {
				return fmt.Errorf("Bucket replication configuration is invalid")
			}
			if rpcErr.RPCCode == "INVALID_LOG_REPLICATION_INFO" {
				return fmt.Errorf("Bucket log replication info configuration is invalid")
			}
		}
		return fmt.Errorf("Provisioner Failed to validate replication of bucket %q with error: %v", bucketName, err)
	}
	log.Infof("ValidateReplicationPolicy: validated replication successfully")
	return nil
}
