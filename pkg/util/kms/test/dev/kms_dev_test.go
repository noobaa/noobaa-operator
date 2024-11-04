package kmsdevtest

import (
	"os"
	"os/exec"

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


func simpleKmsSpec(token, apiAddress string) nbv1.KeyManagementServiceSpec {
	k := nbv1.KeyManagementServiceSpec{}
	k.TokenSecretName = token
	k.ConnectionDetails = map[string]string{
		kms.VaultAddr : apiAddress,
		vault.VaultBackendPathKey : "noobaa/",
		kms.Provider : vault.Name,
	}

	return k
}

func checkExternalSecret(noobaa *nbv1.NooBaa, expectedNil bool) {
	k := noobaa.Spec.Security.KeyManagementService
	uid := string(noobaa.UID)
	driver := kms.NewVault(noobaa.Name, noobaa.Namespace, uid)
	secretPath := driver.Path()
	if v, ok := (driver.Version(nil)).(*kms.VersionRotatingSecret); ok {
		secretPath = v.BackendSecretName()
	}
	path := k.ConnectionDetails[vault.VaultBackendPathKey] + secretPath
	cmd := exec.Command("kubectl", "exec", "vault-0", "--", "vault", "kv", "get", path)
	logger.Printf("Running command: path %v args %v ", cmd.Path, cmd.Args)
	err := cmd.Run()
	actualResult := (err == nil)
	Expect(actualResult == expectedNil).To(BeTrue())
}

func verifyExternalSecretExists(noobaa *nbv1.NooBaa) {
	checkExternalSecret(noobaa, true)
}

func verifyExternalSecretDeleted(noobaa *nbv1.NooBaa) {
	checkExternalSecret(noobaa, false)
}


var _ = Describe("KMS - K8S, Dev Vault", func() {

	Context("Verify K8S KMS NooBaa", func() {
		noobaa := getMiniNooBaa()
		Specify("Create default system", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition status Init", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInit)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, secrets.TypeK8s)).To(BeTrue())
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

	apiAddress, apiAddressFound := os.LookupEnv("API_ADDRESS")
	tokenSecretName, tokenSecretNameFound := os.LookupEnv("TOKEN_SECRET_NAME")

	Context("Verify Vault NooBaa", func() {
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = simpleKmsSpec(tokenSecretName, apiAddress)
		Specify("Verify ENV", func() {
			Expect(apiAddressFound).To(BeTrue())
			logger.Printf("💬 Found API_ADDRESS=%v", apiAddress)

			Expect(tokenSecretNameFound).To(BeTrue())
			logger.Printf("💬 Found TOKEN_SECRET_NAME=%v", tokenSecretName)
			logger.Printf("💬 KMS Spec %v", noobaa.Spec.Security.KeyManagementService)
		})
		Specify("Create Vault Noobaa", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition status Init", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInit)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, secrets.TypeVault)).To(BeTrue())
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

	Context("Verify Vault v2", func() {
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = simpleKmsSpec(tokenSecretName, apiAddress)
		// v1 and v2 backends are defined in install-dev-kms-noobaa.sh
		noobaa.Spec.Security.KeyManagementService.ConnectionDetails[vault.VaultBackendPathKey] = "noobaav2/"
		Specify("Create Vault v2 Noobaa", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition status Init", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInit)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, secrets.TypeVault)).To(BeTrue())
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

	Context("Invalid Vault Configuration", func() {
		Specify("Ivalid Token Secret name", func() {
			noobaa := getMiniNooBaa()
			noobaa.Spec.Security.KeyManagementService = simpleKmsSpec(tokenSecretName, apiAddress)
			noobaa.Spec.Security.KeyManagementService.TokenSecretName = "invalid"
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInvalid)).To(BeTrue())
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
		Specify("Ivalid KMS provider", func() {
			noobaa := getMiniNooBaa()
			noobaa.Spec.Security.KeyManagementService = simpleKmsSpec(tokenSecretName, apiAddress)
			noobaa.Spec.Security.KeyManagementService.ConnectionDetails[kms.Provider] = "invalid"
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSInvalid)).To(BeTrue())
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
	})
})
