package kmskmiptest

import (
	"os"

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

func simpleKmsSpec(token, endpoint string) nbv1.KeyManagementServiceSpec {
	k := nbv1.KeyManagementServiceSpec{}
	k.TokenSecretName = token
	k.ConnectionDetails = map[string]string{
		kms.Provider:          kms.KMIPSecretStorageName,
		kms.KMIPEndpoint:      endpoint,
		kms.KMIPTLSServerName: "kmip.ciphertrustmanager.local",
	}

	return k
}

func checkExternalSecret(noobaa *nbv1.NooBaa, expectExists bool) {
	k := noobaa.Spec.Security.KeyManagementService
	// Fetch the KMIP secret
	secret := &corev1.Secret{}
	secret.Namespace = noobaa.Namespace
	secret.Name = k.TokenSecretName
	Expect(util.KubeCheck(secret)).To(BeTrue())
	_, exists := secret.StringData[kms.KMIPUniqueID]
	Expect(exists == expectExists).To(BeTrue())
}

func verifyExternalSecretExists(noobaa *nbv1.NooBaa) {
	checkExternalSecret(noobaa, true)
}

func verifyExternalSecretDeleted(noobaa *nbv1.NooBaa) {
	checkExternalSecret(noobaa, false)
}

var _ = Describe("KMS - KMIP", func() {

	apiAddress, apiAddressFound := os.LookupEnv(kms.KMIPEndpoint)
	tokenSecretName, tokenSecretNameFound := os.LookupEnv(kms.KMIPSecret)

	Context("Verify KMIP NooBaa", func() {
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = simpleKmsSpec(tokenSecretName, apiAddress)
		Specify("Verify ENV", func() {
			Expect(apiAddressFound).To(BeTrue())
			logger.Printf("💬 Found %v=%v", kms.KMIPEndpoint, apiAddress)

			Expect(tokenSecretNameFound).To(BeTrue())
			logger.Printf("💬 Found %v=%v", kms.KMIPSecret, tokenSecretName)
			logger.Printf("💬 KMS Spec %v", noobaa.Spec.Security.KeyManagementService)
		})
		Specify("Create KMIP Noobaa", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition status Init", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInit)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, kms.KMIPSecretStorageName)).To(BeTrue())
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

	Context("Invalid KMIP Configuration", func() {
		Specify("Ivalid Token Secret name", func() {
			noobaa := getMiniNooBaa()
			noobaa.Spec.Security.KeyManagementService = simpleKmsSpec(tokenSecretName, apiAddress)
			noobaa.Spec.Security.KeyManagementService.TokenSecretName = "invalid"
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInvalid)).To(BeTrue())
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
		Specify("Missing KMIP endpoint", func() {
			noobaa := getMiniNooBaa()
			noobaa.Spec.Security.KeyManagementService = simpleKmsSpec(tokenSecretName, apiAddress)
			delete(noobaa.Spec.Security.KeyManagementService.ConnectionDetails, kms.KMIPEndpoint)
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInvalid)).To(BeTrue())
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
	})
})
