package kmsazurevaulttest

import (
	"log"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var logger *log.Logger

// KMS - Azure Vault - integration test entry point
func TestAzureVaultKMS(t *testing.T) {
	// this variable is defined in .github/workflows/run_kms_*_test.yml
	// indication of running in integration test environment
	_, ok := os.LookupEnv("OPERATOR_IMAGE")
	if !ok {
		t.Skip() // Not an integration test, skip
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "KMS (Azure Vault) Suite")
}

var _ = BeforeSuite(func() {
	logger = log.New(GinkgoWriter, "INFO: ", log.Lshortfile)
}, 60)
