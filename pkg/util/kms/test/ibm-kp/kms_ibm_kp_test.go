package kmsibmkptest

import (
	"os"

	"github.com/google/uuid"
	"github.com/libopenstorage/secrets"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/util/kms"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getMiniNooBaa() *nbv1.NooBaa {
	options.MiniEnv = true
	options.Namespace = corev1.NamespaceDefault
	nb := system.LoadSystemDefaults()
	return nb
}


func ibmKpKmsSpec(token, instanceID string) nbv1.KeyManagementServiceSpec {
	k := nbv1.KeyManagementServiceSpec{}
	k.TokenSecretName = token
	k.ConnectionDetails = map[string]string{
		kms.Provider: kms.IbmKpSecretStorageName,
		kms.IbmInstanceIDKey : instanceID,
	}

	return k
}

func checkExternalSecret(tokenSecretName string, instanceID string, noobaa *nbv1.NooBaa, expectExists bool) {

	k := noobaa.Spec.Security.KeyManagementService
	uid := string(noobaa.UID)
	driver := &kms.IBM{uid}

	// Generate backend configuration using backend driver instance
	c, err := driver.Config(k.ConnectionDetails, k.TokenSecretName, noobaa.Namespace)
	Expect(err).To(BeNil())
	
	// Construct new backend
	s, err := secrets.New(kms.IbmKpSecretStorageName, c)
	Expect(err).To(BeNil())

	// Fetch the key
	_, err = s.GetSecret(driver.Path(), driver.GetContext())
	Expect((err == nil) == expectExists).To(BeTrue())
}

func verifyExternalSecretExists(tokenSecretName string, instanceID string, noobaa *nbv1.NooBaa) {
	checkExternalSecret(tokenSecretName, instanceID, noobaa, true)
}

func verifyExternalSecretDeleted(tokenSecretName string, instanceID string, noobaa *nbv1.NooBaa) {
	checkExternalSecret(tokenSecretName, instanceID, noobaa, false)
}


var _ = Describe("KMS - IBM KP", func() {

	instanceID, instanceIDFound := os.LookupEnv(kms.IbmInstanceIDKey)
	tokenSecretName, tokenSecretNameFound := os.LookupEnv("TOKEN_SECRET_NAME")

	Context("Verify IBM KP NooBaa", func() {
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = ibmKpKmsSpec(tokenSecretName, instanceID)
		Specify("Verify ENV", func() {
			Expect(instanceIDFound).To(BeTrue())
			logger.Printf("💬 Found %v=%v", kms.IbmInstanceIDKey, instanceID)

			Expect(tokenSecretNameFound).To(BeTrue())
			logger.Printf("💬 Found TOKEN_SECRET_NAME=%v", tokenSecretName)
			logger.Printf("💬 KMS Spec %v", noobaa.Spec.Security.KeyManagementService)
		})
		Specify("Create IBM KP Noobaa", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition status Init", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInit)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, kms.IbmKpSecretStorageName)).To(BeTrue())
		})
		Specify("Verify external secrets exists", func() {
			verifyExternalSecretExists(tokenSecretName, instanceID, noobaa)
		})
		Specify("Restart NooBaa operator", func() {
			podList := &corev1.PodList{}
			podSelector, _ := labels.Parse("noobaa-operator=deployment")
			listOptions := client.ListOptions{Namespace: options.Namespace, LabelSelector: podSelector}

			Expect(util.KubeList(podList, &listOptions)).To(BeTrue())
			Expect(len(podList.Items)).To(BeEquivalentTo(1))
			Expect(util.KubeDelete(&podList.Items[0])).To(BeTrue())
		})
		Specify("Verify KMS condition status Sync", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSSync)).To(BeTrue())
		})
		Specify("Delete NooBaa", func() {
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
		Specify("Verify external secrets deletion", func() {
			verifyExternalSecretDeleted(tokenSecretName, instanceID, noobaa)
		})
	})

	Context("IBM KP Secret interface implementation", func() {
		name := "noobaa"
		id := uuid.New().String()
		kmsSpec := ibmKpKmsSpec(tokenSecretName, instanceID)
		k, err := kms.NewKMS(kmsSpec.ConnectionDetails, kmsSpec.TokenSecretName, name, corev1.NamespaceDefault, id)
		Expect(err).To(BeNil())
		plainText := util.RandomBase64(32)

		Specify("Test params", func() {
			logger.Printf("💬 Generated noobaa uuid=%v", id)
			logger.Printf("💬 Generated secret plaintext=%v", plainText)
			logger.Printf("💬 Generated kmsSpec=%v", kmsSpec)
		})
		Specify("Verify uninitialized Get", func() {
			_, err := k.Get()
			logger.Printf("💬 Get err=%v", err)
			Expect(err == secrets.ErrInvalidSecretId).To(BeTrue())
		})
		Specify("Verify Set", func() {
			err := k.Set(plainText)
			Expect(err).To(BeNil())
		})
		Specify("Verify read back", func() {
			s, err := k.Get()
			logger.Printf("💬 Read back secret s=%v error=%v", s, err)
			Expect(err).To(BeNil())
			Expect(s == plainText).To(BeTrue())
		})
		Specify("Verify delete", func() {
			err := k.Delete()
			logger.Printf("💬 Delete error=%v", err)
			Expect(err).To(BeNil())
		})
	})
})
