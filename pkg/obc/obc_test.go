package obc

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("OBC referenced BucketClass", func() {
	Context("When a bucketclass exists in the same namespace that of OBC", func() {
		It("should return the BucketClass wth namespace to be the same as that of OBC", func() {
			obcNS := "obc-ns"
			systemNS := "test"

			obc := &nbv1.ObjectBucketClaim{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: obcNS,
				},
				Spec: nbv1.ObjectBucketClaimSpec{
					AdditionalConfig: map[string]string{
						"bucketclass": "bc1",
					},
				},
			}

			ns, ok := getBucketClass(
				obc,
				nil,
				systemNS,
				func(o client.Object) bool {
					return o.GetNamespace() == obcNS
				},
			)

			Expect(ok).To(BeTrue())
			Expect(ns.GetNamespace()).To(Equal(obcNS))
		})
	})

	Context("When a bucketclass exists in the system namespace", func() {
		It("should return the BucketClass's namespace to be the same as that of system", func() {
			obcNS := "obc-ns"
			systemNS := "test"

			obc := &nbv1.ObjectBucketClaim{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: obcNS,
				},
				Spec: nbv1.ObjectBucketClaimSpec{
					AdditionalConfig: map[string]string{
						"bucketclass": "bc1",
					},
				},
			}

			bc, ok := getBucketClass(
				obc,
				nil,
				systemNS,
				func(o client.Object) bool {
					return o.GetNamespace() == systemNS
				},
			)

			Expect(ok).To(BeTrue())
			Expect(bc.GetNamespace()).To(Equal(systemNS))
		})
	})

	Context("When a bucketclass exists in the system namespace as well as OBC namespace", func() {
		It("should return the BucketClass's namespace to be the same as that of OBC", func() {
			obcNS := "obc-ns"
			systemNS := "test"

			obc := &nbv1.ObjectBucketClaim{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: obcNS,
				},
				Spec: nbv1.ObjectBucketClaimSpec{
					AdditionalConfig: map[string]string{
						"bucketclass": "bc1",
					},
				},
			}

			bc, ok := getBucketClass(
				obc,
				nil,
				systemNS,
				func(o client.Object) bool {
					return o.GetNamespace() == systemNS || o.GetNamespace() == obcNS
				},
			)

			Expect(ok).To(BeTrue())
			Expect(bc.GetNamespace()).To(Equal(obcNS))
		})
	})

	Context("When a bucketclass does not exists in system namespace or OBC namespace", func() {
		It("should return the BucketClass's namespace to be the same as that of system namespace with \"ok\" set to false", func() {
			obcNS := "obc-ns"
			systemNS := "test"

			obc := &nbv1.ObjectBucketClaim{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: obcNS,
				},
				Spec: nbv1.ObjectBucketClaimSpec{
					AdditionalConfig: map[string]string{
						"bucketclass": "bc1",
					},
				},
			}

			bc, ok := getBucketClass(
				obc,
				nil,
				systemNS,
				func(o client.Object) bool {
					return false
				},
			)

			Expect(ok).To(BeFalse())
			Expect(bc.GetNamespace()).To(Equal(systemNS))
		})
	})

	Context("When a bucketclass is not specified in the OBC", func() {
		It("should return empty string for namespace and \"exists\" set to false", func() {
			obcNS := "obc-ns"
			systemNS := "test"

			obc := &nbv1.ObjectBucketClaim{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: obcNS,
				},
				Spec: nbv1.ObjectBucketClaimSpec{
					AdditionalConfig: map[string]string{
						"bucketclass": "",
					},
				},
			}

			bc, ok := getBucketClass(
				obc,
				nil,
				systemNS,
				func(o client.Object) bool {
					return false
				},
			)

			Expect(ok).To(BeFalse())
			Expect(bc.GetNamespace()).To(Equal(""))
		})
	})
})
