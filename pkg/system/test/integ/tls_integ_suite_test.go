package tlsintegtests

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes"
)

var clientset *kubernetes.Clientset
var logger *log.Logger

func TestTLSIntegration(t *testing.T) {
	_, ok := os.LookupEnv("OPERATOR_IMAGE")
	if !ok {
		t.Skip("Skipping integration test: OPERATOR_IMAGE not set")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "TLS Integration Suite")
}

var _ = BeforeSuite(func(ctx context.Context) {
	By("Connecting to K8S cluster")
	logger = log.New(GinkgoWriter, "INFO: ", log.Lshortfile)

	var err error
	clientset, err = kubernetes.NewForConfig(util.KubeConfig())
	Expect(err).ToNot(HaveOccurred())
	Expect(clientset).ToNot(BeNil())
}, NodeTimeout(60*time.Second))
