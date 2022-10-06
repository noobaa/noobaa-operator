package admissionintegtests

import (
	"context"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/backingstore"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucket"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/namespacestore"
	"github.com/noobaa/noobaa-operator/v5/pkg/operator"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace = "test"
)

var _ = Describe("Admission validate setup", func() {
	Context("Validate admission resources", func() {
		It("Should Exist", func() {
			options.Namespace = namespace
			c := operator.LoadOperatorConf(&cobra.Command{})
			operator.LoadAdmissionConf(c)

			Expect(util.KubeCheck(c.WebhookConfiguration)).To(BeTrue())
			Expect(util.KubeCheck(c.WebhookSecret)).To(BeTrue())
			Expect(util.KubeCheck(c.WebhookService)).To(BeTrue())
		})
	})
})

var _ = Describe("Admission server integration tests", func() {

	var (
		testBackingstore   *nbv1.BackingStore
		testNamespacestore *nbv1.NamespaceStore
		testBucketclass    *nbv1.BucketClass
		result             bool
		err                error
	)

	Setup()

	Describe("Create operations", func() {
		BeforeEach(func() {
			testBackingstore = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore)
			testBackingstore.Name = "bs-name"
			testBackingstore.Namespace = namespace
			testNamespacestore = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml).(*nbv1.NamespaceStore)
			testNamespacestore.Name = "ns-name"
			testNamespacestore.Namespace = namespace
			testBucketclass = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml).(*nbv1.BucketClass)
			testBucketclass.Name = "bc-name"
			testBucketclass.Namespace = namespace
		})
		Context("Empty secret name", func() {
			It("Should Deny", func() {
				testBackingstore.Spec = nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAWSS3,
					AWSS3: &nbv1.AWSS3Spec{
						TargetBucket: "some-target-bucket",
						Secret: corev1.SecretReference{
							Name:      "",
							Namespace: namespace,
						},
					},
				}

				result, err = KubeCreate(testBackingstore)
				Expect(result).To(BeFalse())
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("admission webhook \"admissionwebhook.noobaa.io\" denied the request: Failed creating the Backingstore, please provide a valid ARN or secret name"))

				testNamespacestore.Spec = nbv1.NamespaceStoreSpec{
					Type: nbv1.NSStoreTypeAWSS3,
					AWSS3: &nbv1.AWSS3Spec{
						TargetBucket: "some-target-bucket",
						Secret: corev1.SecretReference{
							Name:      "",
							Namespace: namespace,
						},
					},
				}

				result, err = KubeCreate(testNamespacestore)
				Expect(result).To(BeFalse())
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("admission webhook \"admissionwebhook.noobaa.io\" denied the request: Failed creating the NamespaceStore, please provide a valid ARN or secret name"))
			})
		})
		Context("Non Empty AWS STS ARN", func() {
			arnString := "some-aws-arn"
			It("Should Allow", func() {
				stsBackingStore := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore)
				stsBackingStore.Name = "sts-bs-name"
				stsBackingStore.Namespace = namespace

				stsBackingStore.Spec = nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAWSS3,
					AWSS3: &nbv1.AWSS3Spec{
						TargetBucket:  "some-target-bucket",
						AWSSTSRoleARN: &arnString,
					},
				}

				result, err = KubeCreate(stsBackingStore)
				Expect(result).To(BeTrue())
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("Invalid store type", func() {
			It("Should Deny", func() {
				testBackingstore.Spec = nbv1.BackingStoreSpec{
					Type: "invalid",
					AWSS3: &nbv1.AWSS3Spec{
						TargetBucket: "some-target-bucket",
						Secret: corev1.SecretReference{
							Name:      "secret-name",
							Namespace: namespace,
						},
					},
				}

				result, err = KubeCreate(testBackingstore)
				Expect(result).To(BeFalse())
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("admission webhook \"admissionwebhook.noobaa.io\" denied the request: Invalid Backingstore type, please provide a valid Backingstore type"))

				testNamespacestore.Spec = nbv1.NamespaceStoreSpec{
					Type: "invalid",
					AWSS3: &nbv1.AWSS3Spec{
						TargetBucket: "some-target-bucket",
						Secret: corev1.SecretReference{
							Name:      "secret-name",
							Namespace: namespace,
						},
					},
				}

				result, err = KubeCreate(testNamespacestore)
				Expect(result).To(BeFalse())
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("admission webhook \"admissionwebhook.noobaa.io\" denied the request: Invalid Namespacestore type, please provide a valid Namespacestore type"))
			})
		})
		Context("Pass create validations", func() {
			It("Should Allow", func() {
				testBackingstore.Spec = nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypePVPool,
					PVPool: &nbv1.PVPoolSpec{
						NumVolumes: 2,
						VolumeResources: &corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: *resource.NewScaledQuantity(int64(20), resource.Giga),
							},
						},
					},
				}

				result, err = KubeCreate(testBackingstore)
				Expect(result).To(BeTrue())
				Ω(err).ShouldNot(HaveOccurred())
				Expect(backingstore.WaitReady(testBackingstore)).To(BeTrue())

				testNamespacestore.Spec = nbv1.NamespaceStoreSpec{
					Type: nbv1.NSStoreTypeS3Compatible,
					S3Compatible: &nbv1.S3CompatibleSpec{
						TargetBucket: "first.bucket",
						Endpoint:     "s3." + namespace + ".svc.cluster.local:443",
						Secret: corev1.SecretReference{
							Name:      "noobaa-admin",
							Namespace: namespace,
						},
					},
				}

				result, err = KubeCreate(testNamespacestore)
				Expect(result).To(BeTrue())
				Ω(err).ShouldNot(HaveOccurred())
				Expect(namespacestore.WaitReady(testNamespacestore)).To(BeTrue())

				testBucketclass.Spec.NamespacePolicy = &nbv1.NamespacePolicy{
					Type: nbv1.NSBucketClassTypeCache,
					Cache: &nbv1.CacheNamespacePolicy{
						HubResource: "ns-name",
						Caching: &nbv1.CacheSpec{
							TTL:    60,
							Prefix: "test",
						},
					},
				}

				result, err = KubeCreate(testBucketclass)
				Expect(result).To(BeTrue())
				Ω(err).ShouldNot(HaveOccurred())
				Expect(bucketclass.WaitReady(testBucketclass)).To(BeTrue())
			})
		})
	})

	Describe("Update operations", func() {
		Context("Scale Down Number of Nodes", func() {
			It("Should Deny", func() {
				bsList := &nbv1.BackingStoreList{
					TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
				}
				if !util.KubeList(bsList, &client.ListOptions{Namespace: options.Namespace}) {
					return
				}
				bs := &bsList.Items[0]

				bs.Spec.PVPool.NumVolumes = 1

				result, err = KubeUpdate(bs)
				Expect(result).To(BeFalse())
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("admission webhook \"admissionwebhook.noobaa.io\" denied the request: Scaling down the number of nodes is not currently supported"))
			})
		})
		Context("Updated target bucket", func() {
			It("Should Deny", func() {
				nsList := &nbv1.NamespaceStoreList{
					TypeMeta: metav1.TypeMeta{Kind: "NamespaceStoreList"},
				}
				if !util.KubeList(nsList, &client.ListOptions{Namespace: options.Namespace}) {
					return
				}
				ns := &nsList.Items[0]

				ns.Spec.S3Compatible.TargetBucket = "some-other-bucket"

				result, err = KubeUpdate(ns)
				Expect(result).To(BeFalse())
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("admission webhook \"admissionwebhook.noobaa.io\" denied the request: Changing a NamespaceStore target bucket is unsupported"))
			})
		})
	})

	Describe("Delete operations", func() {
		It("Should Deny", func() {
			defaultBs := &nbv1.BackingStore{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "noobaa-default-backing-store",
					Namespace: namespace,
				},
			}

			// try to delete noobaa-default-backing-store and failing because it has data buckets
			result, err = KubeDelete(defaultBs)
			Expect(result).To(BeFalse())
			Ω(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("admission webhook \"admissionwebhook.noobaa.io\" denied the request: cannot complete because pool \"noobaa-default-backing-store\" in \"IN_USE\" state"))
		})
		It("Should Allow", func() {
			// delete "bs-name" backingstore
			result, err = KubeDelete(testBackingstore)
			Expect(result).To(BeTrue())
			Ω(err).ShouldNot(HaveOccurred())

			// delete "ns-name" namespacestore
			result, err = KubeDelete(testNamespacestore)
			Expect(result).To(BeTrue())
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

func Setup() {
	options.Namespace = namespace
	bcList := &nbv1.BucketClassList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClassList"},
	}
	if !util.KubeList(bcList, &client.ListOptions{Namespace: options.Namespace}) {
		return
	}
	bc := &bcList.Items[0]

	bucketclass.WaitReady(bc)
	bucket.RunCreate(&cobra.Command{}, []string{"test.bucket"})
}

func KubeCreate(obj client.Object) (bool, error) {
	client := util.KubeClient()
	err := client.Create(context.TODO(), obj)
	if err == nil {
		return true, err
	}
	return false, err
}

func KubeUpdate(obj client.Object) (bool, error) {
	client := util.KubeClient()
	err := client.Update(context.TODO(), obj)
	if err == nil {
		return true, err
	}
	return false, err
}

func KubeDelete(obj client.Object) (bool, error) {
	client := util.KubeClient()
	err := client.Delete(context.TODO(), obj)
	statusErr, ok := err.(*errors.StatusError)
	if ok {
		return false, statusErr
	}
	return true, err
}
