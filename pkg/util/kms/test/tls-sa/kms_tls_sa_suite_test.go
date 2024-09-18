package kmstlstestsa

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var logger *log.Logger

// KMS integration - TLS Vault - SA authentication test entry point
func TestTlsSAKMS(t *testing.T) {
	// this variable is defined in .github/workflows/run_kms_*_test.yml
	// indication of running in integration test environment
	_, ok := os.LookupEnv("OPERATOR_IMAGE")
	if !ok {
		t.Skip() // Not an integration test, skip
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "External KMS (TLS Vault) Suite")
}

var _ = BeforeSuite(func(ctx context.Context) {
	logger = log.New(GinkgoWriter, "INFO: ", log.Lshortfile)
}, NodeTimeout(60*time.Second))
