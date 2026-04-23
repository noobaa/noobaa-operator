package obc

import (
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
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

var _ = Describe("getExternalDNSDetails", func() {
	noobaaWithExternalDNS := func(s3URLs, vectorsURLs []string) *nbv1.NooBaa {
		return &nbv1.NooBaa{
			Status: nbv1.NooBaaStatus{
				Services: &nbv1.ServicesStatus{
					ServiceS3:      nbv1.ServiceStatus{ExternalDNS: s3URLs},
					ServiceVectors: nbv1.ServiceStatus{ExternalDNS: vectorsURLs},
				},
			},
		}
	}

	It("returns host and port from the first S3 ExternalDNS entry", func() {
		host, port, err := getExternalDNSDetails(
			noobaaWithExternalDNS([]string{"https://s3.route.example.com:443"}, nil),
			externalDNSServiceS3,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(host).To(Equal("s3.route.example.com"))
		Expect(port).To(Equal(443))
	})

	It("returns host and port from the first Vectors ExternalDNS entry", func() {
		host, port, err := getExternalDNSDetails(
			noobaaWithExternalDNS(nil, []string{"https://vectors.apps.example.com:8443"}),
			externalDNSServiceVectors,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(host).To(Equal("vectors.apps.example.com"))
		Expect(port).To(Equal(8443))
	})

	It("uses the first URL when S3 ExternalDNS lists multiple entries", func() {
		host, port, err := getExternalDNSDetails(
			noobaaWithExternalDNS([]string{"https://first.route.test:443", "https://second.route.test:443"}, nil),
			externalDNSServiceS3,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(host).To(Equal("first.route.test"))
		Expect(port).To(Equal(443))
	})

	It("returns an error when the selected service has no ExternalDNS", func() {
		_, _, err := getExternalDNSDetails(
			noobaaWithExternalDNS(nil, []string{"https://vectors-only:443"}),
			externalDNSServiceS3,
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no external"))
		Expect(err.Error()).To(ContainSubstring("s3"))
	})

	It("returns an error for an unknown externalDNSService value", func() {
		_, _, err := getExternalDNSDetails(
			noobaaWithExternalDNS([]string{"https://x:443"}, nil),
			externalDNSService("other"),
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unknown external DNS service"))
	})

	It("returns a parse error when the primary URL is invalid", func() {
		_, _, err := getExternalDNSDetails(
			noobaaWithExternalDNS([]string{"not-a-valid-request-uri"}, nil),
			externalDNSServiceS3,
		)
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when the URL omits a numeric port", func() {
		_, _, err := getExternalDNSDetails(
			noobaaWithExternalDNS([]string{"https://s3.example.com"}, nil),
			externalDNSServiceS3,
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to parse external DNS"))
	})

	It("getExternalDNSDetails returns an error when NooBaa Status.Services is nil", func() {
		for _, nb := range []*nbv1.NooBaa{
			{},
			{Status: nbv1.NooBaaStatus{Services: nil}},
		} {
			_, _, err := getExternalDNSDetails(nb, externalDNSServiceS3)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no services found"))
		}
	})
})
