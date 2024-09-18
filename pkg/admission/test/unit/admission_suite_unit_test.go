package admissionunittests

import (
	"context"
	"log"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var logger *log.Logger

func TestAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Admission Suite")
}

var _ = BeforeSuite(func(ctx context.Context) {
	logger = log.New(GinkgoWriter, "INFO: ", log.Lshortfile)
}, NodeTimeout(60*time.Second))
