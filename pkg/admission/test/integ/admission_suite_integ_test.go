package admissionintegtests

import (
	"log"
	"os"
	"testing"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes"
)

var clientset *kubernetes.Clientset
var logger *log.Logger

func TestAdmission(t *testing.T) {
	// this variable is defined in .github/workflows/run_admission_test.yml
	// indication of running in integration test environment
	_, ok := os.LookupEnv("OPERATOR_IMAGE")
	if !ok {
		t.Skip() // Not an integration test, skip
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Admission Suite")
}

func connectToK8s() (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(util.KubeConfig())
	if err != nil {
		logger.Printf("failed to create K8s clientset")
		return nil, err
	}

	return clientset, nil
}

var _ = BeforeSuite(func() {
	By("Connecting to K8S cluster")
	logger = log.New(GinkgoWriter, "INFO: ", log.Lshortfile)

	var err error
	clientset, err = connectToK8s()
	Expect(err).ToNot(HaveOccurred())
	Expect(clientset).ToNot(BeNil())
}, 60)
