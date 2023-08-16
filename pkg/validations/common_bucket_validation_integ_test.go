package validations

import (
	"fmt"

	. "github.com/onsi/ginkgo"
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
		emptyReplicationPolicy := getEmptyReplicationPolicy()
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

func getEmptyReplicationPolicy() string {
	return ``
}
