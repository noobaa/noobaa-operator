package kmsrotatetest

import (
	"fmt"
	"time"

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

func getMiniNooBaa() *nbv1.NooBaa {
	options.MiniEnv = true
	options.Namespace = corev1.NamespaceDefault
	nb := system.LoadSystemDefaults()
	return nb
}

func getSchedMiniNooBaa() *nbv1.NooBaa {
	nb := getMiniNooBaa()
	nb.Spec.Security.KeyManagementService.EnableKeyRotation = true
	nb.Spec.Security.KeyManagementService.Schedule = "* * * * *" // every min
	return nb
}

var _ = Describe("KMS - K8S Key Rotate", func() {
	log := util.Logger()

	Context("Verify Upgrade", func() {
		noobaa := getMiniNooBaa()
		cipherKeyB64 := util.RandomBase64(32)
		log.Printf("💬 Generated cipher_key_b64=%v", cipherKeyB64)

		Specify("Create old format K8S root master key secret", func() {
			s := &corev1.Secret{}

			s.Name = noobaa.Name + "-root-master-key"
			s.Namespace = noobaa.Namespace
			s.StringData = map[string]string{
				"cipher_key_b64": cipherKeyB64,
			}

			Expect(util.KubeCreateFailExisting(s)).To(BeTrue())
		})
		Specify("Create default system", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, secrets.TypeK8s)).To(BeTrue())
		})
		Specify("Verify KMS condition status Sync", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSSync)).To(BeTrue())
		})
		Specify("Verify new format K8S root master key secret", func() {
			volumeSecret := &corev1.Secret{}
			volumeSecret.Name = noobaa.Name + "-root-master-key-volume"
			volumeSecret.Namespace = noobaa.Namespace
			Expect(util.KubeCheck(volumeSecret)).To(BeTrue())

			activeRootKey, activeRootKeyOk := volumeSecret.StringData[kms.ActiveRootKey]
			log.Printf("💬 Found activeRootKey=%v", activeRootKey)
			Expect(activeRootKeyOk).To(BeTrue())
			rootKeyValue, rootKeyValueOk := volumeSecret.StringData[activeRootKey]
			log.Printf("💬 Found rootKeyValue=%v", rootKeyValue)
			Expect(rootKeyValueOk).To(BeTrue())
			Expect(rootKeyValue == cipherKeyB64).To(BeTrue())
		})
		Specify("Delete NooBaa", func() {
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
	})

	Context("Verify Rotate", func() {
		noobaa := getSchedMiniNooBaa()

		Specify("Create key rotate schedule system", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, secrets.TypeK8s)).To(BeTrue())
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
	
	Context("Verify Rotate - Max key Limit", func() {
		noobaa := getSchedMiniNooBaa()
		now := time.Now()

		Specify("Create new format K8S root master key secret", func() {
			s := &corev1.Secret{}
			secret := map[string]string{}
			for i := 1; i < 60; i++ { // 6 keys are in range - rest can be removed
				secret[fmt.Sprintf("key-%v", now.AddDate(0, -i, 0).UnixNano())] = fmt.Sprintf("%v month a go", i)
			}
			secret["active_root_key"] = fmt.Sprintf("key-%v", now.UnixNano())
			secret[fmt.Sprintf("key-%v", now.UnixNano())] = "latest_key"

			s.Name = noobaa.Name + "-root-master-key-backend"
			s.Namespace = noobaa.Namespace
			s.StringData = secret


			Expect(util.KubeCreateFailExisting(s)).To(BeTrue())
		})
		Specify("Create key rotate schedule system", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, secrets.TypeK8s)).To(BeTrue())
		})
		Specify("Verify KMS condition status Key Rotate", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSKeyRotate)).To(BeTrue())
		})
		Specify("Verify the secret was updated as should", func() {
			s := &corev1.Secret{}
			s.Name = noobaa.Name + "-root-master-key-volume"
			s.Namespace = noobaa.Namespace
			Expect(util.KubeCheck(s)).To(BeTrue())
			Expect(len(s.StringData)).To(BeEquivalentTo(51)) // 50 keys min + "active_key"
		})
		Specify("Delete NooBaa", func() {
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
	})

	Context("Verify Rotate - 6 Months Limit", func() {
		noobaa := getSchedMiniNooBaa()
		now := time.Now()

		Specify("Create new format K8S root master key secret", func() {
			s := &corev1.Secret{}
			secret := map[string]string{}
			for i := 1; i < 50; i++ {
				secret[fmt.Sprintf("key-%v", now.AddDate(0, 0, -i).UnixNano())] = fmt.Sprintf("%v days a go", i)
			}
			secret[fmt.Sprintf("key-%v", now.AddDate(0, -6, 0).UnixNano())] = "6 months a go" // Should be removed
			secret[fmt.Sprintf("key-%v", now.AddDate(0, -7, 0).UnixNano())] = "7 months a go" // Should be removed
			secret["active_root_key"] = fmt.Sprintf("key-%v", now.UnixNano())
			secret[fmt.Sprintf("key-%v", now.UnixNano())] = "latest_key"

			s.Name = noobaa.Name + "-root-master-key-backend"
			s.Namespace = noobaa.Namespace
			s.StringData = secret


			Expect(util.KubeCreateFailExisting(s)).To(BeTrue())
		})
		Specify("Create key rotate schedule system", func() {
			Expect(util.KubeCreateFailExisting(noobaa)).To(BeTrue())
		})
		Specify("Verify KMS condition Type", func() {
			Expect(util.NooBaaCondition(noobaa, nbv1.ConditionTypeKMSType, secrets.TypeK8s)).To(BeTrue())
		})
		Specify("Verify KMS condition status Key Rotate", func() {
			Expect(util.NooBaaCondStatus(noobaa, nbv1.ConditionKMSKeyRotate)).To(BeTrue())
		})
		Specify("Verify the secret was updated as should", func() {
			s := &corev1.Secret{}
			s.Name = noobaa.Name + "-root-master-key-volume"
			s.Namespace = noobaa.Namespace
			Expect(util.KubeCheck(s)).To(BeTrue())
			Expect(len(s.StringData)).To(BeEquivalentTo(52)) // 50 keys in range + "active_key" + 1 new key
		})
		Specify("Delete NooBaa", func() {
			Expect(util.KubeDelete(noobaa)).To(BeTrue())
		})
	})
})
