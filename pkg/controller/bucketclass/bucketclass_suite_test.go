package bucketclass_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBucketclass(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bucketclass Suite")
}
