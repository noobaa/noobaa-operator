package validations

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	firstBucket  = "first.bucket"
	secondBucket = "second.bucket"
	// setting it to false since this is already testEnv
	isCLI = false
)

var _ = Describe("Replication Validation tests", func() {

	It("validate valid replication policy", func() {
		validReplicationPolicy := getValidReplicationPolicy()
		err := ValidateReplicationPolicy(secondBucket, validReplicationPolicy, false, isCLI)
		Expect(err).To(BeNil())
	})

	It("validate invalid replication policy - should fail", func() {
		invalidReplicationPolicy := getInvalidReplicationPolicy()
		err := ValidateReplicationPolicy(firstBucket, invalidReplicationPolicy, false, isCLI)
		expectedErrMsg := fmt.Sprintf("Provisioner Failed to validate replication of bucket \"%s\" with error: INVALID_SCHEMA_PARAMS SERVER bucket_api#/methods/validate_replication", firstBucket)
		Expect(err.Error()).To(Equal(expectedErrMsg))
	})

	It("validate invalid replication policy json - should fail", func() {
		invalidReplicationPolicyJSON := getInvalidReplicationPolicyJSON()
		err := ValidateReplicationPolicy(firstBucket, invalidReplicationPolicyJSON, false, isCLI)
		expectedErrMsg := "Failed to parse replication json {[] <nil>}: unexpected end of JSON input"
		Expect(err.Error()).To(Equal(expectedErrMsg))
	})

	It("validate empty replication policy valid", func() {
		emptyReplicationPolicy := getEmptyString()
		err := ValidateReplicationPolicy(firstBucket, emptyReplicationPolicy, false, isCLI)
		Expect(err).To(BeNil())
	})

	It("validate empty replication rules array on create bucket - should fail", func() {
		emptyRulesArrReplicationPolicy := getEmptyRulesArrReplicationPolicy()
		err := ValidateReplicationPolicy(firstBucket, emptyRulesArrReplicationPolicy, false, isCLI)
		expectedErrMsg := fmt.Sprintf("replication rules array of bucket \"%s\" is empty {[] <nil>}", firstBucket)
		Expect(err.Error()).To(Equal(expectedErrMsg))
	})

	It("validate empty replication rules array on create bucket", func() {
		emptyRulesArrReplicationPolicy := getEmptyRulesArrReplicationPolicy()
		err := ValidateReplicationPolicy(firstBucket, emptyRulesArrReplicationPolicy, true, isCLI)
		Expect(err).To(BeNil())

	})
})

var _ = Describe("NSFS account config validation tests", func() {

	It("validate valid nsfs account config", func() {
		err := ValidateNSFSAccountConfig(getValidNsfsAccountConfig(), getValidBucketclassName())
		Expect(err).To(BeNil())
	})

	It("validate valid nsfs account config with empty bucketclass name - should fail", func() {
		err := ValidateNSFSAccountConfig(getValidNsfsAccountConfig(), getEmptyString())
		expectedErrMsg := "a bucketclass backed by an NSFS namespacestore is required"
		Expect(err.Error()).To(ContainSubstring(expectedErrMsg))
	})

	It("validate invalid nsfs account config - should fail", func() {
		err := ValidateNSFSAccountConfig(getInvalidNSFSAccountConfig(), getValidBucketclassName())
		expectedErrMsg := "NSFS account config must include both"
		Expect(err.Error()).To(ContainSubstring(expectedErrMsg))
	})

	It("validate invalid nsfs account config json - should fail", func() {
		err := ValidateNSFSAccountConfig(getInvalidNSFSAccountConfigJSON(), getValidBucketclassName())
		expectedErrMsg := "failed to parse NSFS account config"
		Expect(err.Error()).To(ContainSubstring(expectedErrMsg))
	})

	It("validate empty nsfs account config", func() {
		err := ValidateNSFSAccountConfig(getEmptyString(), getValidBucketclassName())
		Expect(err).To(BeNil())
	})
})

func getValidReplicationPolicy() string {
	return `{"rules":[{"rule_id":"rule-1","destination_bucket":"first.bucket","filter":{"prefix":"a"}}]}`
}

func getInvalidReplicationPolicy() string {
	return `{"rules":[{"destination_bucket":"second.bucket","filter":{"prefix":"a"}}]}`
}

func getInvalidReplicationPolicyJSON() string {
	// missing } at the end of the string
	return `{"rules":[{"destination_bucket":"second.bucket","filter":{"prefix":"a"}}]`
}

func getEmptyRulesArrReplicationPolicy() string {
	return `{"rules":[]}`
}

func getEmptyString() string {
	return ``
}

func getValidNsfsAccountConfig() string {
	return `{"uid": 42, "gid": 505, "path": "/path/to/nsfs"}`
}

func getInvalidNSFSAccountConfigJSON() string {
	// curly bracket replaced with regular one
	return `{"distinguished_name": "someuser")`
}

func getInvalidNSFSAccountConfig() string {
	// Invalid because it contains both distinguished_name as well as uid+gid
	return `{"distinguished_name": "someuser", "uid": 123, "gid": 123}`
}

func getValidBucketclassName() string {
	return "bucketclass-1"
}