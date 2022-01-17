package kmsibmkptest

import (
	"log"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/noobaa/noobaa-operator/v5/pkg/util/kms"
)

var logger *log.Logger

// KMS - IBM KP - integration test entry point
func TestIBMKPKMS(t *testing.T) {
	// this variable is defined in .github/workflows/run_kms_*_test.yml
	// indication of running in integration test environment
	_, ok := os.LookupEnv("OPERATOR_IMAGE")
	if !ok {
		t.Skip() // Not an integration test, skip
	}

	id, idOk := os.LookupEnv(kms.IbmInstanceIDKey)
	if !idOk || id == "" {
		t.Skip() // No access to secrets
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "KMS (IBM KP) Suite")
}

var _ = BeforeSuite(func() {
	logger = log.New(GinkgoWriter, "INFO: ", log.Lshortfile)
}, 60)
