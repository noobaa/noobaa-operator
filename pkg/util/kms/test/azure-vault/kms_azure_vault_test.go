package kmsazurevaulttest

import (
	"os"

	"github.com/libopenstorage/secrets/azure"
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

func azureKMSSpec(azureVaultURL string) nbv1.KeyManagementServiceSpec {
	k := nbv1.KeyManagementServiceSpec{}
	k.ConnectionDetails = map[string]string{
		kms.AzureVaultURL:             azureVaultURL,
		kms.Provider:                  azure.Name,
		kms.AzureVaultClientID:        "e08b2886-f826-4746-8673-47044661d1a1",
		kms.AzureClientCertSecretName: "azure-ocs-ffwc9o1j",
		kms.AzureVaultTenantID:        "9cf78105-e3e9-4321-b88d-b001b66c762b",
	}

	return k
}

var _ = Describe("KMS - Azure Vault", func() {
	Context("Verify Vault ServiceAccount Kubernetes Auth", func() {
		noobaa := getMiniNooBaa()
		azureVaultURL, azureVaultURLFound := os.LookupEnv("AZURE_VAULT_URL")
		noobaa.Spec.Security.KeyManagementService = azureKMSSpec(azureVaultURL)

		Specify("Verify Azure vault URL", func() {
			Expect(azureVaultURLFound).To(BeTrue())
		})

		Specify("Create KMS Noobaa", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Restart NooBaa operator", func() {
			podList := &corev1.PodList{}
			podSelector, _ := labels.Parse("noobaa-operator=deployment")
			listOptions := client.ListOptions{Namespace: options.Namespace, LabelSelector: podSelector}

			Expect(util.KubeList(podList, &listOptions)).To(BeTrue())
			Expect(len(podList.Items)).To(BeEquivalentTo(1))
			Expect(util.KubeDelete(&podList.Items[0])).To(BeTrue())
		})
		// TODO: As of now azure key vault is a cloud service and to test
		// this case, an account needs to be created at azure side.
		// Create Azure key vault and provide the parameters
		// Below condition always be corev1.ConditionStatus = "Invalid"
		// utill we provide the actual azure key vault credentials
		// Change Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSSync)).To(BeFalse())
		// to Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSSync)).To(BeTrue())
		// once we have azure valut in place
		Specify("Verify KMS condition status Sync", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSSync)).To(BeFalse())
		})
		Specify("Delete NooBaa", func() {
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
	})
	Context("Verify Azure vault fail", func() {
		noobaa := getMiniNooBaa()
		azureVaultURL, azureVaultURLFound := os.LookupEnv("AZURE_VAULT_URL")
		k := azureKMSSpec(azureVaultURL)
		noobaa.Spec.Security.KeyManagementService = k

		Specify("Verify Azure vault URL", func() {
			Expect(azureVaultURLFound).To(BeTrue())
		})

		Specify("Create KMS Noobaa", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition status Invalid", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInvalid)).To(BeTrue())
		})
		Specify("Delete NooBaa", func() {
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
	})

})
