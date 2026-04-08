package bucket

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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
