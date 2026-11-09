package bucketclass

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// systemNS is the namespace which is used in the unit tests to describe
	// noobaa system namespace
	systemNS = "test"
)

var _ = Describe("Verify Bucketclass provisioner actions", func() {
	Context("When bucketclass is in the same namespace as NooBaa system", func() {
		It("should allow object for the provisioner", func() {
			obj := &nbv1.BucketClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: systemNS,
				},
			}

			Expect(isObjectForProvisioner(obj, systemNS)).To(BeTrue())
		})
	})

	Context("When bucketclass is not in the same namespace as NooBaa system", func() {
		It("should disallow object for the provisioner", func() {
			obj := &nbv1.BucketClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "random",
				},
			}

			Expect(isObjectForProvisioner(obj, systemNS)).To(BeFalse())
		})
	})

	Context("When bucketclass is not in the same namespace as NooBaa system: Valid provisioner label", func() {
		It("should allow object for the provisioner", func() {
			obj := &nbv1.BucketClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "random",
					Labels: map[string]string{
						"noobaa-operator": systemNS,
					},
				},
			}

			Expect(isObjectForProvisioner(obj, systemNS)).To(BeTrue())
		})
	})

	Context("When bucketclass is not in the same namespace as NooBaa system: Invalid provisioner label", func() {
		It("should disallow object for the provisioner", func() {
			obj := &nbv1.BucketClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "random",
					Labels: map[string]string{
						"noobaa-operator": "xyz",
					},
				},
			}

			Expect(isObjectForProvisioner(obj, systemNS)).To(BeFalse())
		})
	})
})

var _ = Describe("BucketClass backing store readiness", func() {
	It("should return a retryable error when a backingstore name is empty so UpdateBucketClass is not called", func() {
		err := validations.ValidatePlacementPolicy(&nbv1.PlacementPolicy{
			Tiers: []nbv1.Tier{{BackingStores: []string{""}}},
		}, systemNS)
		Expect(err).To(HaveOccurred())
		Expect(util.IsPersistentError(err)).To(BeFalse())
	})
})
