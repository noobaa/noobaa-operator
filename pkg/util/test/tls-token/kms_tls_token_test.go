package kmstlstesttoken

import (
	"os"

	"github.com/libopenstorage/secrets/vault"
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

func tlsTokenKMSSpec(token, apiAddress string) nbv1.KeyManagementServiceSpec {
	kms := nbv1.KeyManagementServiceSpec{}
	kms.TokenSecretName = token
	kms.ConnectionDetails = map[string]string{
		util.VaultAddr : apiAddress,
		vault.VaultBackendPathKey : "noobaa/",
		util.KmsProvider: vault.Name,
		util.VaultCaCert: "vault-ca-cert",
		util.VaultClientCert: "vault-client-cert",
		util.VaultClientKey: "vault-client-key",
		util.VaultSkipVerify: "true",
	}

	return kms
}

var _ = Describe("KMS - TLS Vault Token", func() {
	apiAddress, apiAddressFound := os.LookupEnv("API_ADDRESS")
	tokenSecretName, tokenSecretNameFound := os.LookupEnv("TOKEN_SECRET_NAME")

	Context("Verify Vault Token Auth", func() {
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = tlsTokenKMSSpec(tokenSecretName, apiAddress)

		Specify("Verify API Address", func() {
			Expect(apiAddressFound).To(BeTrue())
		})
		Specify("Verify Token secret", func() {
			Expect(tokenSecretNameFound).To(BeTrue())
			logger.Printf("💬 Found TOKEN_SECRET_NAME=%v", tokenSecretName)
			logger.Printf("💬 KMS Spec %v", noobaa.Spec.Security.KeyManagementService)
		})
		Specify("Create KMS Noobaa", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition status Init", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInit)).To(BeTrue())
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
	})
})
