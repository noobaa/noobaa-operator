package cosi

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	namespacestore "github.com/noobaa/noobaa-operator/v5/pkg/namespacestore"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("COSI CLI tests", func() {
	os.Setenv("TEST_ENV", "true")
	options.Namespace = "test"

	CLIPath := "../../build/_output/bin/noobaa-operator-local"
	defaultBackingStoreName := "noobaa-default-backing-store"
	firstBucket := "first.bucket"
	pcBCName := "pc-bucket-class-cli"
	nsBCName := "ns-bucket-class-cli"
	acName1 := "access-class-cli1"
	acName2 := "access-class-cli2"
	aclaim := "access-claim-cli"
	aclaimSecret := "access-claim-secret-cli"
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

	Context("Bucket Access class CLI functions", func() {

		It("cosi bucketAccessclass create", func() {
			cmd := exec.Command(CLIPath, "cosi", "accessclass", "create", acName1, "-n", options.Namespace)
			log.Printf("Running command: create %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: access class create out %s ", string(out))
			Expect(err).To(BeNil())
			Expect(strings.Contains(string(out), "✅ Created: BucketAccessClass")).To(BeTrue())
			expectedName := fmt.Sprintf("Name:\n %s", acName1)
			expectedDriverName := fmt.Sprintf("Driver Name:\n %s", options.COSIDriverName())
			expectedAuthenticationType := fmt.Sprintf("Authentication Type:\n %s", nbv1.COSIKEYAuthenticationType)
			Expect(strings.Contains(string(out), expectedName)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedDriverName)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedAuthenticationType)).To(BeTrue())
		})

		It("cosi bucketAccessclass status", func() {
			cmd := exec.Command(CLIPath, "cosi", "accessclass", "status", acName1, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: access class status out %s ", string(out))
			Expect(err).To(BeNil())
			expectedName := fmt.Sprintf("Name:\n %s", acName1)
			expectedDriverName := fmt.Sprintf("Driver Name:\n %s", options.COSIDriverName())
			expectedAuthenticationType := fmt.Sprintf("Authentication Type:\n %s", nbv1.COSIKEYAuthenticationType)
			Expect(strings.Contains(string(out), expectedName)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedDriverName)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedAuthenticationType)).To(BeTrue())

		})

		It("cosi bucketAccessclass list", func() {
			cmd := exec.Command(CLIPath, "cosi", "accessclass", "list", "-n", options.Namespace)
			log.Printf("Running command: list %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: access class list out %s ", string(out))
			Expect(err).To(BeNil())
		})

		It("cosi bucketAccessclass delete", func() {
			cmd := exec.Command(CLIPath, "cosi", "accessclass", "delete", acName1, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: access class delete out %s ", string(out))
			Expect(err).To(BeNil())
			expectedDeleteMsg := fmt.Sprintf("Deleted : BucketAccessClass \\\"%s\\\"", acName1)
			Expect(strings.Contains(string(out), expectedDeleteMsg)).To(BeTrue())
		})
	})

	Context("Bucket Access claim CLI functions", func() {

		It("cosi bucketAccessclaim create", func() {

			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "create", "placement-bucketclass", pcBCName, "--backingstores", defaultBackingStoreName, "--deletion-policy", "retain", "-n", options.Namespace)
			_, err := cmd.CombinedOutput()
			Expect(err).To(BeNil())
			cmd = exec.Command(CLIPath, "cosi", "bucketclaim", "create", claimName, "--bucketclass", pcBCName, "-n", options.Namespace)
			_, err = cmd.CombinedOutput()
			Expect(err).To(BeNil())
			cmd = exec.Command(CLIPath, "cosi", "accessclass", "create", acName2, "-n", options.Namespace)
			_, err = cmd.CombinedOutput()
			Expect(err).To(BeNil())

			cmd = exec.Command(CLIPath, "cosi", "accessclaim", "create", aclaim, "--bucket-claim", claimName, "--bucket-access-class", acName2, "--creds-secret-name", aclaimSecret, "-n", options.Namespace)
			log.Printf("Running command: create access claim %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: access claim create out %s ", string(out))
			Expect(err).To(BeNil())
			Expect(strings.Contains(string(out), "✅ Created: BucketAccess")).To(BeTrue())

			cosiBucketAccessClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_access_claim_yaml).(*nbv1.COSIBucketAccessClaim)
			cosiBucketAccessClaim.Name = aclaim
			cosiBucketAccessClaim.Namespace = options.Namespace
			Expect(util.KubeCheck(cosiBucketAccessClaim)).To(BeTrue())
			Expect(cosiBucketAccessClaim.Status.AccessGranted).To(BeTrue())
			Expect(cosiBucketAccessClaim.Status.AccountID).NotTo(BeNil())

			expectedCreateBucket := fmt.Sprintf("  %-22s : %s\n", "S3Access", "true")
			expectedAccessGranted := fmt.Sprintf("  %-22s : %s\n", "AllowBucketCreate", "false")
			expectedDefaultResource := fmt.Sprintf("  %-22s : %s", "DefaultResource", "system-internal-storage-pool")
			log.Printf("create status prints expectedCreateBucket=%s \n expectedAccessGranted=%s \n expectedDefaultResource=%s \n", expectedCreateBucket, expectedAccessGranted, expectedDefaultResource)

			expectedName := fmt.Sprintf("  %-22s : %s\n", "Name", cosiBucketAccessClaim.Status.AccountID)

			Expect(strings.Contains(string(out), expectedName)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedAccessGranted)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedDefaultResource)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedCreateBucket)).To(BeTrue())
		})

		It("cosi bucketAccessclass status", func() {
			cmd := exec.Command(CLIPath, "cosi", "accessclaim", "status", aclaim, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: access claim status out %s ", string(out))
			Expect(err).To(BeNil())
			cosiBucketAccessClaim := util.KubeObject(bundle.File_deploy_cosi_bucket_access_claim_yaml).(*nbv1.COSIBucketAccessClaim)
			cosiBucketAccessClaim.Name = aclaim
			cosiBucketAccessClaim.Namespace = options.Namespace
			Expect(util.KubeCheck(cosiBucketAccessClaim)).To(BeTrue())

			expectedName := fmt.Sprintf("  %-22s : %s\n", "Name", cosiBucketAccessClaim.Status.AccountID)
			expectedCreateBucket := fmt.Sprintf("  %-22s : %s\n", "S3Access", "true")
			expectedAccessGranted := fmt.Sprintf("  %-22s : %s\n", "AllowBucketCreate", "false")
			expectedDefaultResource := fmt.Sprintf("  %-22s : %s", "DefaultResource", "system-internal-storage-pool")
			log.Printf("status prints expectedCreateBucket=%s \n expectedAccessGranted=%s \n expectedDefaultResource=%s \n", expectedCreateBucket, expectedAccessGranted, expectedDefaultResource)

			Expect(strings.Contains(string(out), expectedName)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedAccessGranted)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedDefaultResource)).To(BeTrue())
			Expect(strings.Contains(string(out), expectedCreateBucket)).To(BeTrue())
		})

		It("cosi bucketAccessclass list", func() {
			cmd := exec.Command(CLIPath, "cosi", "accessclaim", "list", "-n", options.Namespace)
			log.Printf("Running command: list %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: access claim list out %s ", string(out))
			Expect(err).To(BeNil())
			Expect(strings.Contains(string(out), aclaim)).To(BeTrue())

		})

		It("cosi bucketAccessclass delete", func() {
			cmd := exec.Command(CLIPath, "cosi", "accessclaim", "delete", aclaim, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err := cmd.CombinedOutput()
			log.Printf("Running command: access claim delete out %s ", string(out))
			Expect(err).To(BeNil())
			expectedDeleteMsg := fmt.Sprintf("Deleted : BucketAccess \\\"%s\\\"", aclaim)
			Expect(strings.Contains(string(out), expectedDeleteMsg)).To(BeTrue())
			cmd = exec.Command("kubectl", "get", "accessclaim", aclaim, "-n", options.Namespace)
			log.Printf("Running command: status %v args %v ", cmd.Path, cmd.Args)
			out, err = cmd.CombinedOutput()
			log.Printf("Running command: kubectl access claim delete out %s ", string(out))
			Expect(strings.Contains(string(out), "error: the server doesn't have a resource type \"accessclaim\"")).To(BeTrue())
			Expect(err).NotTo(BeNil())
		})
		It("Cleanup", func() {

			cmd := exec.Command(CLIPath, "cosi", "bucketclass", "delete", pcBCName, "-n", options.Namespace)
			_, err := cmd.CombinedOutput()
			log.Printf("Running command: bucketclass access claim delete err %q ", err)
			Expect(err).To(BeNil())

			cmd = exec.Command(CLIPath, "cosi", "bucketclaim", "delete", claimName, "-n", options.Namespace)
			log.Printf("Running command: bucketclaim access claim delete err %q ", err)
			_, err = cmd.CombinedOutput()
			Expect(err).To(BeNil())

			cmd = exec.Command(CLIPath, "cosi", "accessclass", "delete", acName2, "-n", options.Namespace)
			log.Printf("Running command: accessclass access claim delete err %q ", err)
			_, err = cmd.CombinedOutput()
			Expect(err).To(BeNil())
		})
	})
})
