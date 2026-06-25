package bucket

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CmdUpdate", func() {
	It("should register the deep-archive-resource flag", func() {
		cmd := CmdUpdate()
		f := cmd.Flags().Lookup("deep-archive-resource")
		Expect(f).NotTo(BeNil())
		Expect(f.DefValue).To(Equal(""))
	})
})

var _ = Describe("buildArchivePolicyForUpdate", func() {
	It("should return a valid archive policy config", func() {
		policy, err := buildArchivePolicyForUpdate("archive-ns")
		Expect(err).NotTo(HaveOccurred())
		Expect(policy).NotTo(BeNil())
		Expect(policy.DeepArchiveResource).NotTo(BeNil())
		Expect(policy.DeepArchiveResource.Resource).To(Equal("archive-ns"))
	})
})

var _ = Describe("IsArchiveNamespaceStore for deep archive update", func() {
	It("should return true when archive is set", func() {
		ns := &nbv1.NamespaceStore{
			Spec: nbv1.NamespaceStoreSpec{
				Type:    nbv1.NSStoreTypeS3Compatible,
				Archive: true,
			},
		}
		Expect(util.IsArchiveNamespaceStore(ns)).To(BeTrue())
	})

	It("should return false when archive is not set on s3-compatible store", func() {
		ns := &nbv1.NamespaceStore{
			Spec: nbv1.NamespaceStoreSpec{
				Type:    nbv1.NSStoreTypeS3Compatible,
				Archive: false,
			},
		}
		Expect(util.IsArchiveNamespaceStore(ns)).To(BeFalse())
	})

	It("should return false when archive is not set on aws-s3 store", func() {
		ns := &nbv1.NamespaceStore{
			Spec: nbv1.NamespaceStoreSpec{
				Type:    nbv1.NSStoreTypeAWSS3,
				Archive: false,
			},
		}
		Expect(util.IsArchiveNamespaceStore(ns)).To(BeFalse())
	})

	It("should return false for a nil NamespaceStore", func() {
		Expect(util.IsArchiveNamespaceStore(nil)).To(BeFalse())
	})
})

var _ = Describe("mergeQuotaForUpdate", func() {
	Context("when only max-objects is set and a size quota already exists", func() {
		It("should preserve the size limit and apply the object limit", func() {
			existing := &nb.QuotaConfig{
				// Unit uses NooBaa short suffixes (K,M,G,...) from GetBytesAndUnits / QuotaSizeToBytes, not Kubernetes GiB-style strings.
				Size: &nb.SizeQuotaConfig{Value: 5, Unit: "G"},
			}
			mergedQuota, err := mergeQuotaForUpdate(existing, "b", "", "2")
			Expect(err).NotTo(HaveOccurred())
			Expect(mergedQuota).To(Equal(nb.QuotaConfig{
				Size:     &nb.SizeQuotaConfig{Value: 5, Unit: "G"},
				Quantity: &nb.QuantityQuotaConfig{Value: 2},
			}))
		})
	})

	Context("when only max-objects is set and both limits already exist", func() {
		It("should preserve size and replace the object limit", func() {
			existing := &nb.QuotaConfig{
				Size:     &nb.SizeQuotaConfig{Value: 1.5, Unit: "G"},
				Quantity: &nb.QuantityQuotaConfig{Value: 10},
			}
			mergedQuota, err := mergeQuotaForUpdate(existing, "b", "", "3")
			Expect(err).NotTo(HaveOccurred())
			Expect(mergedQuota.Size).NotTo(BeNil())
			Expect(mergedQuota.Size.Value).To(Equal(1.5))
			Expect(mergedQuota.Size.Unit).To(Equal("G"))
			Expect(mergedQuota.Quantity).NotTo(BeNil())
			Expect(mergedQuota.Quantity.Value).To(Equal(3))
		})
	})

	Context("when max-size is explicitly cleared with zero", func() {
		It("should remove the size limit and keep the object limit", func() {
			existing := &nb.QuotaConfig{
				Size:     &nb.SizeQuotaConfig{Value: 5, Unit: "G"},
				Quantity: &nb.QuantityQuotaConfig{Value: 2},
			}
			mergedQuota, err := mergeQuotaForUpdate(existing, "b", "0", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(mergedQuota.Size).To(BeNil())
			Expect(mergedQuota.Quantity).NotTo(BeNil())
			Expect(mergedQuota.Quantity.Value).To(Equal(2))
		})
	})
})

var _ = Describe("ValidateQuotaAgainstBucketUsage", func() {
	It("returns a validation error when maxObjects quota is below current usage", func() {
		bucket := &nb.BucketInfo{Name: "b", BucketType: "REGULAR"}
		bucket.NumObjects = &struct {
			Value      int64 `json:"value"`
			LastUpdate int64 `json:"last_update"`
		}{Value: 10}
		quota := &nb.QuotaConfig{Quantity: &nb.QuantityQuotaConfig{Value: 5}}
		err := nb.ValidateQuotaAgainstBucketUsage(bucket, quota)
		Expect(err).To(HaveOccurred())
		Expect(util.IsValidationError(err)).To(BeTrue())
	})

	It("returns a validation error when maxSize quota is below current data usage", func() {
		used := nb.BigInt{N: 10000}
		bucket := &nb.BucketInfo{Name: "b", BucketType: "REGULAR"}
		bucket.NumObjects = &struct {
			Value      int64 `json:"value"`
			LastUpdate int64 `json:"last_update"`
		}{Value: 0}
		bucket.DataCapacity = &struct {
			Size                      *nb.BigInt `json:"size,omitempty"`
			SizeReduced               *nb.BigInt `json:"size_reduced,omitempty"`
			Free                      *nb.BigInt `json:"free,omitempty"`
			AvailableSizeToUpload     *nb.BigInt `json:"available_size_for_upload,omitempty"`
			AvailableQuantityToUpload *nb.BigInt `json:"available_quantity_for_upload,omitempty"`
			LastUpdate                int64      `json:"last_update"`
		}{
			Size: &used,
		}
		// 1 KiB cap (QuotaSizeToBytes) vs 10000 bytes used
		quota := &nb.QuotaConfig{Size: &nb.SizeQuotaConfig{Value: 1, Unit: "K"}}
		err := nb.ValidateQuotaAgainstBucketUsage(bucket, quota)
		Expect(err).To(HaveOccurred())
		Expect(util.IsValidationError(err)).To(BeTrue())
	})
})
