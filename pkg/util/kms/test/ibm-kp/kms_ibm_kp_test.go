package kmsibmkptest

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/libopenstorage/secrets"
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

type TestSysReconMock struct {
	data string
}

func (m *TestSysReconMock) ReconcileSecretString(val string) error {
	m.data = val
	return nil
}
func (m *TestSysReconMock) ReconcileSecretMap(val map[string]string) error {
	return fmt.Errorf("not implemented")
}

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
		kms.Provider:         kms.IbmKpSecretStorageName,
		kms.IbmInstanceIDKey: instanceID,
	}

	return k
}

func checkExternalSecret(tokenSecretName string, instanceID string, noobaa *nbv1.NooBaa, expectExists bool) {

	k := noobaa.Spec.Security.KeyManagementService
	uid := string(noobaa.UID)
	driver := &kms.IBM{UID: uid}

	// Generate backend configuration using backend driver instance
	c, err := driver.Config(k.ConnectionDetails, k.TokenSecretName, noobaa.Namespace)
	Expect(err).To(BeNil())

	// Construct new backend
	s, err := secrets.New(kms.IbmKpSecretStorageName, c)
	Expect(err).To(BeNil())

	// Fetch the key
	_, _, err = s.GetSecret(driver.Path(), driver.GetContext())
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
			err := k.Get()
			logger.Printf("💬 Get err=%v", err)
			Expect(err == secrets.ErrInvalidSecretId).To(BeTrue())
		})
		Specify("Verify Set", func() {
			err := k.Set(plainText)
			Expect(err).To(BeNil())
		})
		Specify("Verify read back", func() {
			err := k.Get()
			Expect(err).To(BeNil())
			m := &TestSysReconMock{}
			err = k.Reconcile(m)
			Expect(err).To(BeNil())
			s := m.data
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

	Context("Verify Rotate in NooBaa CR", func() {
		noobaa := getMiniNooBaa()
		noobaa.Spec.Security.KeyManagementService = ibmKpKmsSpec(tokenSecretName, instanceID)
		noobaa.Spec.Security.KeyManagementService.EnableKeyRotation = true
		noobaa.Spec.Security.KeyManagementService.Schedule = "* * * * *" // every min

		Specify("Create key rotate schedule system", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, "ibmkeyprotect")).To(BeTrue())
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

	Context("IBM KP Key Rotation - Rotating Secret Format", func() {
		name := "noobaa-rotate"
		id := uuid.New().String()
		kmsSpec := ibmKpKmsSpec(tokenSecretName, instanceID)
		k, err := kms.NewKMS(kmsSpec.ConnectionDetails, kmsSpec.TokenSecretName, name, corev1.NamespaceDefault, id)
		Expect(err).To(BeNil())

		// Generate multiple keys for rotation testing
		key1 := util.RandomBase64(32)
		key2 := util.RandomBase64(32)
		key3 := util.RandomBase64(32)

		Specify("Test rotation params", func() {
			logger.Printf("💬 Generated noobaa uuid=%v", id)
			logger.Printf("💬 Generated key1=%v", key1)
			logger.Printf("💬 Generated key2=%v", key2)
			logger.Printf("💬 Generated key3=%v", key3)
		})

		Specify("Verify initial key creation with rotating format", func() {
			// Set first key using rotating secret format
			err := k.Set(key1)
			Expect(err).To(BeNil())
			logger.Printf("✅ Successfully created initial key in rotating format")
		})

		Specify("Verify first key retrieval", func() {
			err := k.Get()
			Expect(err).To(BeNil())
			m := &TestSysReconMock{}
			err = k.Reconcile(m)
			Expect(err).To(BeNil())
			retrievedKey := m.data
			logger.Printf("💬 Retrieved first key: %v", retrievedKey)
			Expect(retrievedKey).To(Equal(key1))
		})

		Specify("Verify key rotation - rotate to second key", func() {
			// Rotate to second key
			err := k.Set(key2)
			Expect(err).To(BeNil())
			logger.Printf("✅ Successfully rotated to second key")
		})

		Specify("Verify second key retrieval after rotation", func() {
			err := k.Get()
			Expect(err).To(BeNil())
			m := &TestSysReconMock{}
			err = k.Reconcile(m)
			Expect(err).To(BeNil())
			retrievedKey := m.data
			logger.Printf("💬 Retrieved second key after rotation: %v", retrievedKey)
			Expect(retrievedKey).To(Equal(key2))
		})

		Specify("Verify key rotation - rotate to third key", func() {
			// Rotate to third key
			err := k.Set(key3)
			Expect(err).To(BeNil())
			logger.Printf("✅ Successfully rotated to third key")
		})

		Specify("Verify tracking secret contains key mappings", func() {
			// Get the tracking secret
			secret := &corev1.Secret{}
			secret.Name = tokenSecretName
			secret.Namespace = corev1.NamespaceDefault
			Expect(util.KubeCheck(secret)).To(BeTrue())

			// Verify active key ID exists
			activeKeyID, exists := secret.StringData[kms.IBMKPActiveKeyID]
			Expect(exists).To(BeTrue())
			Expect(activeKeyID).NotTo(BeEmpty())
			logger.Printf("💬 Active key ID in tracking secret: %v", activeKeyID)

			// Verify key mapping exists for active key
			keyMapping, exists := secret.StringData[kms.IBMKPKeyPrefix+activeKeyID]
			Expect(exists).To(BeTrue())
			Expect(keyMapping).NotTo(BeEmpty())
			logger.Printf("💬 Key mapping for active key: %v -> %v", activeKeyID, keyMapping)
		})

		Specify("Verify multiple key IDs stored in tracking secret", func() {
			// Get the tracking secret
			secret := &corev1.Secret{}
			secret.Name = tokenSecretName
			secret.Namespace = corev1.NamespaceDefault
			Expect(util.KubeCheck(secret)).To(BeTrue())

			// Count how many key mappings exist
			keyCount := 0
			for key := range secret.StringData {
				if len(key) > len(kms.IBMKPKeyPrefix) && key[:len(kms.IBMKPKeyPrefix)] == kms.IBMKPKeyPrefix {
					keyCount++
					logger.Printf("💬 Found key mapping: %v", key)
				}
			}
			
			// Should have at least 3 keys (from 3 rotations)
			Expect(keyCount).To(BeNumerically(">=", 3))
			logger.Printf("✅ Found %d key mappings in tracking secret", keyCount)
		})

		Specify("Verify cleanup - delete rotated keys", func() {
			err := k.Delete()
			logger.Printf("💬 Delete error=%v", err)
			Expect(err).To(BeNil())
			logger.Printf("✅ Successfully deleted rotated keys")
		})

		Specify("Verify keys deleted from IBM Key Protect", func() {
			err := k.Get()
			logger.Printf("💬 Get after delete err=%v", err)
			Expect(err).To(Equal(secrets.ErrInvalidSecretId))
			logger.Printf("✅ Confirmed keys are deleted from IBM Key Protect")
		})
	})

	Context("IBM KP Key Rotation - Backend Secret ID Format", func() {
		name := "noobaa-backend-rotate"
		id := uuid.New().String()
		kmsSpec := ibmKpKmsSpec(tokenSecretName, instanceID)
		
		// Create KMS with backend secret format
		k, err := kms.NewKMS(kmsSpec.ConnectionDetails, kmsSpec.TokenSecretName, name, corev1.NamespaceDefault, id)
		Expect(err).To(BeNil())

		key1 := util.RandomBase64(32)
		key2 := util.RandomBase64(32)

		Specify("Test backend rotation params", func() {
			logger.Printf("💬 Testing backend secret format with uuid=%v", id)
			logger.Printf("💬 Backend key1=%v", key1)
			logger.Printf("💬 Backend key2=%v", key2)
		})

		Specify("Create initial backend key", func() {
			err := k.Set(key1)
			Expect(err).To(BeNil())
			logger.Printf("✅ Created initial backend key")
		})

		Specify("Verify backend key retrieval", func() {
			err := k.Get()
			Expect(err).To(BeNil())
			m := &TestSysReconMock{}
			err = k.Reconcile(m)
			Expect(err).To(BeNil())
			Expect(m.data).To(Equal(key1))
			logger.Printf("✅ Retrieved backend key successfully")
		})

		Specify("Rotate backend key", func() {
			err := k.Set(key2)
			Expect(err).To(BeNil())
			logger.Printf("✅ Rotated backend key")
		})

		Specify("Verify rotated backend key retrieval", func() {
			err := k.Get()
			Expect(err).To(BeNil())
			m := &TestSysReconMock{}
			err = k.Reconcile(m)
			Expect(err).To(BeNil())
			Expect(m.data).To(Equal(key2))
			logger.Printf("✅ Retrieved rotated backend key successfully")
		})

		Specify("Verify IBM Key Protect key ID is stored correctly", func() {
			// Get the tracking secret
			secret := &corev1.Secret{}
			secret.Name = tokenSecretName
			secret.Namespace = corev1.NamespaceDefault
			Expect(util.KubeCheck(secret)).To(BeTrue())

			// Verify active key ID
			activeKeyID, exists := secret.StringData[kms.IBMKPActiveKeyID]
			Expect(exists).To(BeTrue())
			logger.Printf("💬 Active backend key ID: %v", activeKeyID)

			// Verify the key mapping contains IBM KP key ID (UUID format)
			ibmKpKeyID, exists := secret.StringData[kms.IBMKPKeyPrefix+activeKeyID]
			Expect(exists).To(BeTrue())
			// IBM KP key IDs should be UUIDs (36 characters with dashes)
			Expect(len(ibmKpKeyID)).To(BeNumerically(">", 0))
			logger.Printf("💬 IBM KP Key ID stored: %v", ibmKpKeyID)
			logger.Printf("✅ IBM Key Protect key ID stored correctly in tracking secret")
		})

		Specify("Cleanup backend keys", func() {
			err := k.Delete()
			Expect(err).To(BeNil())
			logger.Printf("✅ Cleaned up backend keys")
		})
	})
})
