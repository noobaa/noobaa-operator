package validations

import (
	"encoding/json"
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

// ValidateNSFSAccountConfig validates that the provided NSFS config is valid
func ValidateNSFSAccountConfig(NSFSConfig string) error {
	log := util.Logger()

	if NSFSConfig == "" {
		return nil
	}

	var configObj nbv1.AccountNsfsConfig

	err := json.Unmarshal([]byte(NSFSConfig), &configObj)
	if err != nil {
		return fmt.Errorf("failed to parse NSFS config %q: %v", NSFSConfig, err)
	}

	log.Infof("Validating NSFS config: %+v", NSFSConfig)
	// UID validation
	if configObj.UID < 0 {
		return fmt.Errorf("UID must be a whole positive number")
	}

	// GID validation
	if configObj.GID < 0 {
		return fmt.Errorf("GID must be a whole positive number")
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
