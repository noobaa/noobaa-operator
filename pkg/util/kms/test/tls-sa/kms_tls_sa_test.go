package kmstlstestsa

import (
	"os"

	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/vault"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/util/kms"
	. "github.com/onsi/ginkgo/v2"
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

func tlsSAKMSSpec(apiAddress string) nbv1.KeyManagementServiceSpec {
	k := nbv1.KeyManagementServiceSpec{}
	k.ConnectionDetails = map[string]string{
		kms.VaultAddr : apiAddress,
		vault.VaultBackendPathKey : "noobaa/",
		kms.Provider: vault.Name,
		vault.AuthMethod: vault.AuthMethodKubernetes,
		kms.VaultCaCert: "vault-ca-cert",
		kms.VaultClientCert: "vault-client-cert",
		kms.VaultClientKey: "vault-client-key",
		kms.VaultSkipVerify: "true",
		vault.AuthKubernetesRole : "noobaa",
	}

	return k
}

var _ = Describe("KMS - TLS Vault SA", func() {
	apiAddress, apiAddressFound := os.LookupEnv("API_ADDRESS")
	Context("Verify Vault ServiceAccount Kubernetes Auth", func() {
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = tlsSAKMSSpec(apiAddress)

		Specify("Verify API Address", func() {
			Expect(apiAddressFound).To(BeTrue())
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
	Context("Verify Vault TLS fail", func() {
		noobaa := getMiniNooBaa()
		k := tlsSAKMSSpec(apiAddress)
		delete (k.ConnectionDetails, kms.VaultSkipVerify)
		noobaa.Spec.Security.KeyManagementService = k
		Specify("Verify API Address", func() {
			Expect(apiAddressFound).To(BeTrue())
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

	Context("Verify Rotate", func() {
		apiAddress, apiAddressFound := os.LookupEnv("API_ADDRESS")
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = tlsSAKMSSpec(apiAddress)
		noobaa.Spec.Security.KeyManagementService.EnableKeyRotation = true
		noobaa.Spec.Security.KeyManagementService.Schedule = "* * * * *" // every min

		Specify("Verify API Address", func() {
			Expect(apiAddressFound).To(BeTrue())
		})
		Specify("Create key rotate schedule system", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, secrets.TypeVault)).To(BeTrue())
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
		Specify("Verify KMS condition status Key Rotate", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSKeyRotate)).To(BeTrue())
		})
		Specify("Delete NooBaa", func() {
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
	})

})
