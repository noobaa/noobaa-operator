package upgrade_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	operatorRepo         = "noobaa/noobaa-operator:"
	operatorQuayRepo     = "quay.io/noobaa/noobaa-operator:"
	ns                   = "noobaa-upgrade-test-ns"
	obcNameBeforeUpgrade = "upgrade-test-obc-before-upgrade"
	obcNameAfterUpgrade  = "upgrade-test-obc-after-upgrade"
	numObjectsToPut      = 5
	objectKeyLength      = 5
	objectContentLength  = 5 * 1024 * 1024         // 5 MB
	s3Endpoint           = "http://localhost:8080" // Assuming port-forwarding is set up to localhost:8080
	CLIpath              = "../../build/_output/bin/noobaa-operator-local"
	charset              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var accessKey, secretKey string
var log = util.Logger()

// Notice that the upgrade tests are serial, so they run one after another
var _ = Describe("NooBaa Operator Upgrade Tests", Serial, func() {
	KeyToContentBeforeUpgradeMap := make(map[string]string)
	It("should install NooBaa operator at source version", func() {
		operatorSourceVersion, err := getOperatorSourceVersion()
		Expect(err).NotTo(HaveOccurred())
		installNooBaa(operatorSourceVersion)
	})

	It("should port forward NooBaa service", func() {
		portForwardNoobaa()
	})

	It("should create OBC before upgrade", func() {
		createObc(obcNameBeforeUpgrade, "noobaa-default-bucket-class")
	})

	It("should put/get objects before the upgrade", func() {
		s3Client := getS3ClientForOBC(obcNameBeforeUpgrade)
		for i := 0; i < numObjectsToPut; i++ {
			objectKey := generateRandomString(objectKeyLength)
			objectContent := generateRandomString(objectContentLength)
			putObject(s3Client, obcNameBeforeUpgrade, objectKey, objectContent)
			KeyToContentBeforeUpgradeMap[objectKey] = objectContent
		}
		for objectKey, objectContent := range KeyToContentBeforeUpgradeMap {
			getObject(s3Client, obcNameBeforeUpgrade, objectKey, objectContent)
		}
	})

	It("should stop port forward NooBaa service", func() {
		stopPortForwardNoobaa()
	})

	It("should upgrade correctly", func() {
		operatorTargetVersion, err := getOperatorTargetVersion()
		Expect(err).NotTo(HaveOccurred())
		upgradeNooBaa(operatorTargetVersion)
	})

	It("should port forward NooBaa service", func() {
		portForwardNoobaa()
	})

	It("should get objects that were uploaded before the upgrade correctly", func() {
		s3Client := getS3ClientForOBC(obcNameBeforeUpgrade)
		for objectKey, objectContent := range KeyToContentBeforeUpgradeMap {
			getObject(s3Client, obcNameBeforeUpgrade, objectKey, objectContent)
		}
	})

	It("should put/get objects after the upgrade on bucket before upgrade correctly", func() {
		s3Client := getS3ClientForOBC(obcNameBeforeUpgrade)
		for i := 0; i < numObjectsToPut; i++ {
			objectKey := generateRandomString(objectKeyLength)
			objectContent := generateRandomString(objectContentLength)
			putObject(s3Client, obcNameBeforeUpgrade, objectKey, objectContent)
			getObject(s3Client, obcNameBeforeUpgrade, objectKey, objectContent)
		}
	})

	It("should put/get bucket after the upgrade correctly", func() {
		createObc(obcNameAfterUpgrade, "noobaa-default-bucket-class")
	})

	It("should put/get objects after the upgrade on bucket after upgrade correctly", func() {
		s3Client := getS3ClientForOBC(obcNameAfterUpgrade)
		for i := 0; i < numObjectsToPut; i++ {
			objectKey := generateRandomString(objectKeyLength)
			objectContent := generateRandomString(objectContentLength)
			putObject(s3Client, obcNameAfterUpgrade, objectKey, objectContent)
			getObject(s3Client, obcNameAfterUpgrade, objectKey, objectContent)
		}
	})

	It("should stop port forward NooBaa service", func() {
		stopPortForwardNoobaa()
	})
})

// portForwardNoobaa sets up port forwarding for the NooBaa service to allow access to S3.
func portForwardNoobaa() {
	cmd := exec.Command("kubectl", "port-forward", "-n", ns, "service/s3", "8080:80")
	Expect(cmd.Start()).To(Succeed())
	time.Sleep(5 * time.Second) // wait for port forward to establish
}

// stopPortForwardNoobaa stops port forwarding for the NooBaa S3 service.
func stopPortForwardNoobaa() {
	cmd := exec.Command("pkill", "-f", "kubectl port-forward")
	Expect(cmd.Start()).To(Succeed())
	time.Sleep(5 * time.Second) // wait for port forward to establish
}

// getOBCSpec returns a new ObjectBucketClaim (OBC) specification with the given name and bucket class.
func getOBCSpec(obcName string, bucketclass string) *nbv1.ObjectBucketClaim {
	obc := &nbv1.ObjectBucketClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obcName,
			Namespace: ns,
		},
		Spec: nbv1.ObjectBucketClaimSpec{
			BucketName:       obcName,
			StorageClassName: fmt.Sprintf("%s.noobaa.io", ns),
			AdditionalConfig: map[string]string{
				"bucketclass": bucketclass,
			},
		},
	}
	return obc
}

// createObc creates an Object Bucket Claim (OBC) with the specified name and bucket class.
// TODO - support more types of bucketclasses/obcs
func createObc(obcName string, bucketclass string) {
	obc := getOBCSpec(obcName, bucketclass)
	ok := util.KubeApply(obc)
	Expect(ok).To(BeTrue())
	validateOBCIsBound(obcName)
}

var _ = BeforeSuite(func(ctx context.Context) {
	util.InitLogger(logrus.DebugLevel)
	_ = exec.Command("kubectl", "delete", "ns", ns, "--ignore-not-found").Run()
	Expect(exec.Command("kubectl", "create", "ns", ns).Run()).To(Succeed())
}, NodeTimeout(7*60*time.Second))

var _ = AfterSuite(func(ctx context.Context) {
	cmd := exec.Command(CLIpath, "uninstall", "--cleanup", "--cleanup_data", "-n", ns)
	cmd.Stdin = strings.NewReader("y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	Expect(cmd.Start()).To(Succeed())
}, NodeTimeout(5*60*time.Second))

///////////////////////////////
//        CLI FUNCTIONS      //
///////////////////////////////

// installNooBaa installs the NooBaa operator at the specified source version.
func installNooBaa(operatorSourceVersion string) {
	Expect(exec.Command(CLIpath, "--dev", fmt.Sprintf("--operator-image=%s%s", operatorRepo, operatorSourceVersion), "install", "-n", ns).Run()).To(Succeed())
	validateNooBaaCR()
}

// upgradeNooBaa upgrades the NooBaa operator to the specified target version.
// It uses the CLI tool to perform the upgrade and validates the NooBaa CR status
// TODO - add validation of the new version of core & db images This can be done by checking the version in the NooBaa CR
func upgradeNooBaa(upgradeTargetVersion string) {
	Expect(exec.Command(CLIpath, fmt.Sprintf("--operator-image=%s%s", operatorQuayRepo, upgradeTargetVersion), "upgrade", "-n", ns).Run()).To(Succeed())
	validateNooBaaOperatorImage(upgradeTargetVersion)
	validateNooBaaCR()
}

//////////////////////////////
//        S3 FUNCTIONS      //
//////////////////////////////

func getS3ClientForOBC(obcName string) *s3.S3 {
	accessKey, secretKey = getAccountCredentials(obcName)
	Expect(accessKey).ToNot(BeEmpty())
	Expect(secretKey).ToNot(BeEmpty())
	s3Client, err := getS3Client(accessKey, secretKey)
	Expect(err).NotTo(HaveOccurred())
	Expect(s3Client).NotTo(BeNil())
	return s3Client
}

// getS3Client creates an S3 client using the provided access key and secret key.
func getS3Client(accessKey string, secretKey string) (*s3.S3, error) {
	client := &http.Client{Transport: util.InsecureHTTPTransport}
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			"",
		),
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(s3Endpoint),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       client,
	}
	s3Session, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}
	s3Client := s3.New(s3Session)
	return s3Client, nil
}

// getAccountCredentials retrieves the AWS access key and secret key from the OBC secret.
func getAccountCredentials(bucketName string) (string, string) {
	get := func(key string) string {
		out, err := exec.Command("kubectl", "-n", ns, "get", "secret", bucketName, "-o", fmt.Sprintf("jsonpath={.data.%s}", key)).Output()
		Expect(err).NotTo(HaveOccurred())
		decoded, err := base64.StdEncoding.DecodeString(string(out))
		Expect(err).NotTo(HaveOccurred())
		return string(decoded)
	}
	accessKey := get("AWS_ACCESS_KEY_ID")
	secretKey := get("AWS_SECRET_ACCESS_KEY")
	return accessKey, secretKey
}

// putObject uploads an object to the specified bucket.
func putObject(s3Client *s3.S3, bucketName string, objectKey string, objectContent string) {
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
		Body:   strings.NewReader(objectContent),
	})

	Expect(err).To(BeNil())
	log.Infof("Successfully put object %s to bucket %s\n", objectKey, bucketName)
}

// getObject retrieves an object from the specified bucket and validates its content.
func getObject(s3Client *s3.S3, bucketName string, objectKey string, objectContent string) {
	getObjectOutput, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	Expect(err).NotTo(HaveOccurred())
	buf := new(bytes.Buffer)
	buf.Reset()
	_, err = buf.ReadFrom(getObjectOutput.Body)
	Expect(err).NotTo(HaveOccurred())
	validateObjectContent(buf.String(), objectContent)
	log.Infof("Successfully get object %s to bucket %s\n", objectKey, bucketName)
}

//////////////////////////////
//    VALIDATION FUNCTIONS  //
//////////////////////////////

// validateNooBaaCR waits for the NooBaa CR to be in Ready state.
// TODO - validate noobaa core & db images
func validateNooBaaCR() {
	Eventually(func() string {
		out, _ := exec.Command("kubectl", "-n", ns, "get", "noobaa", "noobaa", "-o", "jsonpath={.status.phase}").Output()
		return string(out)
	}, 3*time.Minute, 5*time.Second).Should(Equal("Ready"))
}

// validateNooBaaOperatorImage asserts that the NooBaa operator image is running the expected version.
func validateNooBaaOperatorImage(expectedVersion string) {
	Eventually(func() string {
		out, _ := exec.Command("kubectl", "-n", ns, "get", "deployment", "noobaa-operator", "-o", "jsonpath={.spec.template.spec.containers[0].image}").Output()
		log.Infof("Current NooBaa Operator image: %s\n expected version: %s\n", string(out), expectedVersion)
		return string(out)
	}, 3*time.Minute, 5*time.Second).Should(ContainSubstring(expectedVersion))

	Eventually(func() string {
		out, _ := exec.Command("kubectl", "-n", ns, "get", "deployment", "noobaa-operator", "-o", "jsonpath={.status.readyReplicas}").Output()
		log.Infof("NooBaa Operator deployment ready replicas: %s\n", string(out))
		return string(out)
	}, 3*time.Minute, 5*time.Second).Should(Equal("1"))
}

// validateObjectContent checks if the actual content matches the expected content.
func validateObjectContent(actualContent string, expectedContent string) {
	Expect(actualContent).To(Equal(expectedContent))
}

// validateOBCIsBound checks if the OBC is in Bound state.
func validateOBCIsBound(obcName string) {
	Eventually(func() (string, error) {
		out, err := exec.Command("kubectl", "-n", ns, "get", "obc", obcName, "-o", "json").Output()
		if err != nil {
			return "", err
		}
		var obc struct {
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		}
		if json.Unmarshal(out, &obc) != nil {
			return "", fmt.Errorf("failed to unmarshal OBC status")
		}
		return obc.Status.Phase, nil
	}, 4*time.Minute, 5*time.Second).Should(Equal("Bound"))
}

////////////////////////////
//    VERSIONS FUNCTIONS  //
////////////////////////////

// getLastNightlyOperatorBuild retrieves the latest nightly operator build tag from Quay.io API.
func getLastNightlyOperatorBuild() (string, error) {
	operatorTagsQuayPath := "https://quay.io/api/v1/repository/noobaa/noobaa-operator/tag/"
	cmd := exec.Command("curl", "-s", operatorTagsQuayPath)
	out, err := cmd.CombinedOutput()
	Expect(err).To(BeNil())
	var tagsMap map[string]interface{}
	err = json.Unmarshal(out, &tagsMap)
	Expect(err).To(BeNil())
	tags := tagsMap["tags"]
	tagsSlice, ok := tags.([]interface{})
	Expect(ok).To(BeTrue())
	Expect(len(tagsSlice)).To(BeNumerically(">", 0))
	tagInfo, ok := tagsSlice[0].(map[string]interface{})
	Expect(ok).To(BeTrue())
	name, ok := tagInfo["name"].(string)
	Expect(ok).To(BeTrue())
	Expect(err).To(BeNil())
	log.Infof("Latest nightly operator build tag: %s\n", name)
	return name, nil
}

// getLatestOperatorRelease retrieves the latest operator release tag from GitHub API.
func getLatestOperatorRelease() (string, error) {
	releaseTagsQuayPath := "https://api.github.com/repos/noobaa/noobaa-operator/releases/latest"
	cmd := exec.Command("curl", "-s", releaseTagsQuayPath)
	out, err := cmd.CombinedOutput()
	Expect(err).To(BeNil())
	var latestRelease map[string]interface{}
	err = json.Unmarshal(out, &latestRelease)
	Expect(err).To(BeNil())
	log.Infof("latestTag %s\n", latestRelease["tag_name"])
	Expect(err).To(BeNil())
	Expect(latestRelease["tag_name"]).ToNot(BeNil())
	tagName, ok := latestRelease["tag_name"].(string)
	Expect(ok).ToNot(BeNil())
	return tagName, nil
}

// getOperatorSourceVersion retrieves the operator source version from environment variables or defaults to the latest release.
func getOperatorSourceVersion() (string, error) {
	versionFromEnv := os.Getenv("UPGRADE_TEST_OPERATOR_SOURCE_VERSION")
	if versionFromEnv != "" {
		return versionFromEnv, nil
	}

	latestTag, err := getLatestOperatorRelease()
	if err == nil && latestTag != "" {
		trimmedTag := strings.TrimPrefix(latestTag, "v")
		return trimmedTag, nil
	}
	return "", err
}

// getOperatorTargetVersion retrieves the operator target version from environment variables or defaults to the latest nightly build.
// If the environment variable is not set, it fetches the latest nightly build tag.
func getOperatorTargetVersion() (string, error) {
	versionFromEnv := os.Getenv("UPGRADE_TEST_OPERATOR_TARGET_VERSION")
	if versionFromEnv != "" {
		return versionFromEnv, nil
	}
	latestNightlyBuildTag, err := getLastNightlyOperatorBuild()
	if err == nil && latestNightlyBuildTag != "" {
		return latestNightlyBuildTag, nil
	}
	return "", err
}

// generateRandomString generates a random string of the specified length using alphanumeric characters.
func generateRandomString(length int) string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// TODO - enable noobaa-core source and target versions functionality
// getCoreSourceVersion retrieves the core source version from environment variables or defaults to an empty string.
// func getCoreSourceVersion() string {
// 	versionFromEnv := os.Getenv("UPGRADE_TEST_CORE_SOURCE_VERSION")
// 	if versionFromEnv != "" {
// 		return versionFromEnv
// 	}
// 	return ""
// }

// // getCoreTargetVersion retrieves the core target version from environment variables or defaults to an empty string.
// func getCoreTargetVersion() string {
// 	versionFromEnv := os.Getenv("UPGRADE_TEST_CORE_TARGET_VERSION")
// 	if versionFromEnv != "" {
// 		return versionFromEnv
// 	}
// 	return ""
// }
