package kmsibmkptest

import (
	"os"

	"github.com/google/uuid"
	"github.com/libopenstorage/secrets"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
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
	kms := nbv1.KeyManagementServiceSpec{}
	kms.TokenSecretName = token
	kms.ConnectionDetails = map[string]string{
		util.KmsProvider: util.IbmKpK8sSecretName,
		util.IbmInstanceIDKey : instanceID,
	}

	return kms
}

func checkExternalSecret(noobaa *nbv1.NooBaa, expectExists bool) {
	kms := noobaa.Spec.Security.KeyManagementService

	secret := &corev1.Secret{}
	secret.Name = kms.TokenSecretName
	secret.Namespace = noobaa.Namespace
	_, _, err := util.KubeGet(secret)
	Expect(err).To(BeNil())

	uid := string(noobaa.UID)
	driver := &util.KMSIBM{uid}
	secretID := driver.Path()
	wdekName := "wdek_" + secretID
	_, ok := secret.Data[wdekName]
	Expect(ok == expectExists).To(BeTrue())
}

func verifyExternalSecretExists(noobaa *nbv1.NooBaa) {
	checkExternalSecret(noobaa, true)
}

func verifyExternalSecretDeleted(noobaa *nbv1.NooBaa) {
	checkExternalSecret(noobaa, false)
}


var _ = Describe("KMS - IBM KP", func() {

	instanceID, instanceIDFound := os.LookupEnv(util.IbmInstanceIDKey)
	tokenSecretName, tokenSecretNameFound := os.LookupEnv("TOKEN_SECRET_NAME")

	Context("Verify IBM KP NooBaa", func() {
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = ibmKpKmsSpec(tokenSecretName, instanceID)
		Specify("Verify ENV", func() {
			Expect(instanceIDFound).To(BeTrue())
			logger.Printf("💬 Found %v=%v", util.IbmInstanceIDKey, instanceID)

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
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, util.IbmKpK8sSecretName)).To(BeTrue())
		})
		Specify("Verify external secrets exists", func() {
			verifyExternalSecretExists(noobaa)
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
			verifyExternalSecretDeleted(noobaa)
		})
	})

	Context("IBM KP Secret interface implementation", func() {
		name := "noobaa"
		id := uuid.New().String()
		kmsSpec := ibmKpKmsSpec(tokenSecretName, instanceID)
		k, err := util.NewKMS(kmsSpec.ConnectionDetails, kmsSpec.TokenSecretName, name, corev1.NamespaceDefault, id)
		Expect(err).To(BeNil())
		plainText := uuid.New().String()

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
	})
})
