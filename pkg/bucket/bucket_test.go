package bucket

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("mergeQuotaForUpdate", func() {
	Context("when only max-objects is set and a size quota already exists", func() {
		It("should preserve the size limit and apply the object limit", func() {
			existing := &nb.QuotaConfig{
				Size: &nb.SizeQuotaConfig{Value: 5, Unit: "GIB"},
			}
			mergedQuota, err := mergeQuotaForUpdate(existing, "b", "", "2")
			Expect(err).NotTo(HaveOccurred())
			Expect(mergedQuota).To(Equal(nb.QuotaConfig{
				Size:     &nb.SizeQuotaConfig{Value: 5, Unit: "GIB"},
				Quantity: &nb.QuantityQuotaConfig{Value: 2},
			}))
		})
	})

	Context("when only max-objects is set and both limits already exist", func() {
		It("should preserve size and replace the object limit", func() {
			existing := &nb.QuotaConfig{
				Size:     &nb.SizeQuotaConfig{Value: 1.5, Unit: "GIB"},
				Quantity: &nb.QuantityQuotaConfig{Value: 10},
			}
			mergedQuota, err := mergeQuotaForUpdate(existing, "b", "", "3")
			Expect(err).NotTo(HaveOccurred())
			Expect(mergedQuota.Size).NotTo(BeNil())
			Expect(mergedQuota.Size.Value).To(Equal(1.5))
			Expect(mergedQuota.Size.Unit).To(Equal("GIB"))
			Expect(mergedQuota.Quantity).NotTo(BeNil())
			Expect(mergedQuota.Quantity.Value).To(Equal(3))
		})
	})

	Context("when max-size is explicitly cleared with zero", func() {
		It("should remove the size limit and keep the object limit", func() {
			existing := &nb.QuotaConfig{
				Size:     &nb.SizeQuotaConfig{Value: 5, Unit: "GIB"},
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
