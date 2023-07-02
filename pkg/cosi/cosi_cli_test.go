package cosi

import (
	"os"
	"os/exec"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	namespacestore "github.com/noobaa/noobaa-operator/v5/pkg/namespacestore"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("COSI driver/provisioner tests", func() {
	os.Setenv("TEST_ENV", "true")
	options.Namespace = "test"

	CLIPath := "../../build/_output/bin/noobaa-operator-local"
	defaultBackingStoreName := "noobaa-default-backing-store"
	firstBucket := "first.bucket"
	pcBCName := "pc-bucket-class-cli"
	nsBCName := "ns-bucket-class-cli"
	invalidNsBCName := "invalid-ns-bucket-class-cli"
	claimName := "cosi-claim-cli"
	cosiNSResourceName := "cosi-nsr-cli"
	var nsResource *nbv1.NamespaceStore
	log := util.Logger()

	sysClient, err := system.Connect(true)
	if err != nil {
		log.Infof("COSI bucket creation tests, can't create sysclient: err= %q", err)
	}

	// create namespace resource on top of first.bucket
	nsResource = getNamespaceResourceObj(cosiNSResourceName, firstBucket, sysClient)
	Expect(util.KubeCreateSkipExisting(nsResource)).To(BeTrue())
	namespacestore.WaitReady(nsResource)

	Context("Bucket class CLI functions", func() {

		It("cosi placement bucketclass create", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "create", "placement-bucketclass", pcBCName, "--backingstores", defaultBackingStoreName, "--deletion-policy", "retain", "-n", options.Namespace)
			log.Printf("Running command: create %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: class pc create out %s ", string(out))
			Expect(err).To(BeNil())
			Expect(strings.Contains(string(out), "✅ Created: BucketClass")).To(BeTrue())
		})

		It("cosi namespace bucketclass create", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "create", "namespace-bucketclass", "single", nsBCName, "--resource", cosiNSResourceName, "--deletion-policy", "delete", "-n", options.Namespace)
			log.Printf("Running command: create %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: class ns create out %s ", string(out))
			Expect(err).To(BeNil())
			Expect(strings.Contains(string(out), "✅ Created: BucketClass")).To(BeTrue())
		})

		It("cosi namespace bucketclass create invalid deletion policy", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "create", "namespace-bucketclass", "single", invalidNsBCName, "--resource", cosiNSResourceName, "--deletion-policy", "invalid-deletion-policy", "-n", options.Namespace)
			log.Printf("Running command: create %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: class ns create out %s %+v", string(out), err)
			Expect(err).NotTo(BeNil())
		})

		It("cosi namespace bucketclass status", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "status", pcBCName, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: class pc status out %s ", string(out))
			Expect(err).To(BeNil())
			Expectedstatus := "Spec:\n map[placementPolicy:{\"tiers\":[{\"backingStores\":[\"noobaa-default-backing-store\"]}]}]"
			Expect(strings.Contains(string(out), Expectedstatus)).To(BeTrue())
		})

		It("cosi namespace bucketclass list", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "list", "-n", options.Namespace)
			log.Printf("Running command: list %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: class pc list out %s ", string(out))
			Expect(err).To(BeNil())
			cleanupKubeResources(nsResource)
		})

		It("cosi namespace bucketclass delete", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "delete", nsBCName, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: class pc delete out %s ", string(out))
			Expect(err).To(BeNil())
			expectedDeleteMsg := "Deleted : BucketClass \\\"ns-bucket-class-cli\\\""
			Expect(strings.Contains(string(out), expectedDeleteMsg)).To(BeTrue())
		})
	})

	Context("Bucket claim CLI functions", func() {

		It("cosi bucketclaim create", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclaim", "create", claimName, "--bucketclass", pcBCName, "-n", options.Namespace)
			log.Printf("Running command: claim create %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: claim create out=%s \n err=%+v", string(out), err)
			Expect(err).To(BeNil())
			// TODO - add read bucket and compare
		})

		It("cosi bucketclaim status", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclaim", "status", claimName, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: claim pc status out %s ", string(out))
			Expect(err).To(BeNil())
		})

		It("cosi bucketclaim list", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclaim", "list", "-n", options.Namespace)
			log.Printf("Running command: list %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: claim list out %s ", string(out))
			Expect(err).To(BeNil())
		})

		It("cosi bucketclaim delete", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclaim", "delete", claimName, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: claim delete out %s ", string(out))
			Expect(err).To(BeNil())
		})

		It("cosi placement bucketclass delete", func() {
			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "delete", pcBCName, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: class pc delete out %s ", string(out))
			Expect(err).To(BeNil())
		})
	})
})
