package admissionunittests

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("BackingStore admission unit tests", func() {

	var (
		bs  *nbv1.BackingStore
		err error
	)

	BeforeEach(func() {
		bs = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore)
		bs.Name = "bs-name"
		bs.Namespace = "test"

	})

	Describe("Validate create operations", func() {
		Describe("General backingstore validations", func() {
			Context("Empty secret name", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "",
								Namespace: "test",
							},
						},
					}
					err = validations.ValidateBSEmptySecretName(*bs)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Failed creating the Backingstore, please provide a valid ARN or secret name"))
				})
				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "full-secret-name",
								Namespace: "test",
							},
						},
					}
					err = validations.ValidateBSEmptySecretName(*bs)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
			Context("Empty Target Bucket", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "",
						},
					}
					err = validations.ValidateBSEmptyTargetBucket(*bs)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Failed creating the Backingstore, please provide target bucket"))
				})
				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
						},
					}
					err = validations.ValidateBSEmptyTargetBucket(*bs)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
			Context("Invalid store type", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: "invalid",
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					err = validations.ValidateBackingStore(*bs)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Invalid Backingstore type, please provide a valid Backingstore type"))
				})
				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					err = validations.ValidateBackingStore(*bs)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
		Describe("Pvpool backingstore", func() {
			Context("Resource name too long", func() {
				It("Should Deny", func() {
					bs.Name = "pvpool-too-long-name-should-fail-after-exceeding-43-characters"
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
					}

					err = validations.ValidatePvpoolNameLength(*bs)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Unsupported BackingStore name length, please provide a name shorter then 43 characters"))
				})

				It("Should Allow", func() {
					bs.Name = "pvpool-not-too-long-name"
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
					}

					err = validations.ValidatePvpoolNameLength(*bs)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
			Context("Minimum volume count", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: -5,
						},
					}

					err = validations.ValidateMinVolumeCount(*bs)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Unsupported volume count, the minimum supported volume count is 1"))

				})
				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 5,
						},
					}

					err = validations.ValidateMinVolumeCount(*bs)
					Ω(err).ShouldNot(HaveOccurred())

				})
			})
			Context("Maximum volume count", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 25,
						},
					}

					err = validations.ValidateMaxVolumeCount(*bs)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Unsupported volume count, the maximum supported volume count is 20"))

				})
				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 15,
						},
					}

					err = validations.ValidateMaxVolumeCount(*bs)
					Ω(err).ShouldNot(HaveOccurred())

				})
			})
			Context("Minimum volume size", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							VolumeResources: &corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: *resource.NewScaledQuantity(int64(5), resource.Giga),
								},
							},
						},
					}

					err = validations.ValidatePvpoolMinVolSize(*bs)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Invalid volume size, minimum volume size is 16Gi"))
				})

				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							VolumeResources: &corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: *resource.NewScaledQuantity(int64(20), resource.Giga),
								},
							},
						},
					}

					err = validations.ValidatePvpoolMinVolSize(*bs)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
		Describe("S3 Compatible backingstore", func() {
			Context("Invalid signature version", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeS3Compatible,
						S3Compatible: &nbv1.S3CompatibleSpec{
							SignatureVersion: "v5",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					err = validations.ValidateSigVersion(bs.Spec.S3Compatible.SignatureVersion)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Invalid S3 compatible Backingstore signature version, please choose either v2/v4"))
				})

				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeS3Compatible,
						S3Compatible: &nbv1.S3CompatibleSpec{
							SignatureVersion: "v4",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					err = validations.ValidateSigVersion(bs.Spec.S3Compatible.SignatureVersion)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})

	Describe("Validate update operations", func() {
		var (
			updatedBS *nbv1.BackingStore
		)

		BeforeEach(func() {
			updatedBS = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore)
			updatedBS.Name = "bs-name"
			updatedBS.Namespace = "test"
		})

		Describe("Pvpool backingstore", func() {
			Context("Scale down node count", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 10,
						},
					}

					updatedBS.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 15,
						},
					}

					err = validations.ValidatePvpoolScaleDown(*bs, *updatedBS)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Scaling down the number of nodes is not currently supported"))
				})

				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 15,
						},
					}

					updatedBS.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 10,
						},
					}

					err = validations.ValidatePvpoolScaleDown(*bs, *updatedBS)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
		Describe("Cloud backingstore", func() {
			Context("Update target bucket", func() {
				It("Should Deny", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
						},
					}

					updatedBS.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-other-bucket",
						},
					}

					err = validations.ValidateTargetBSBucketChange(*bs, *updatedBS)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Changing a Backingstore target bucket is unsupported"))
				})

				It("Should Allow", func() {
					bs.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "same-target-bucket",
						},
					}

					updatedBS.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "same-target-bucket",
						},
					}

					err = validations.ValidateTargetBSBucketChange(*bs, *updatedBS)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})

var _ = Describe("NamespaceStore admission unit tests", func() {

	var (
		ns  *nbv1.NamespaceStore
		err error
	)

	BeforeEach(func() {
		ns = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml).(*nbv1.NamespaceStore)
		ns.Name = "ns-name"
		ns.Namespace = "test"
	})

	Describe("Validate create operations", func() {
		Describe("General namespacestore validations", func() {
			Context("Empty secret name", func() {
				It("Should Deny", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "",
								Namespace: "test",
							},
						},
					}
					err = validations.ValidateNSEmptySecretName(*ns)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Failed creating the namespacestore, please provide secret name"))
				})
				It("Should Allow", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}
					err = validations.ValidateNSEmptySecretName(*ns)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
			Context("Empty Target Bucket", func() {
				It("Should Deny", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "",
						},
					}
					err = validations.ValidateNSEmptyTargetBucket(*ns)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Failed creating the namespacestore, please provide target bucket"))
				})
				It("Should Allow", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
						},
					}
					err = validations.ValidateNSEmptyTargetBucket(*ns)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
			Context("Invalid store type", func() {
				It("Should Deny", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: "invalid",
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}
					err = validations.ValidateNamespaceStore(ns)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Invalid Namespacestore type, please provide a valid Namespacestore type"))
				})
				It("Should Allow", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}
					err = validations.ValidateNamespaceStore(ns)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
		Describe("NSFS validations", func() {
			Context("Empty pvc name", func() {
				It("Should Deny", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "",
						},
					}
					err = validations.ValidateNsStoreNSFS(ns)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("PvcName must not be empty"))
				})
				It("Should Allow", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "pvc-name",
						},
					}
					err = validations.ValidateNsStoreNSFS(ns)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
			Context("Invalid SubPath", func() {
				It("Should Deny", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "pvc-name",
							SubPath: "/path",
						},
					}
					err = validations.ValidateNsStoreNSFS(ns)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("SubPath /path must be a relative path"))
				})
				It("Should Deny", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "pvc-name",
							SubPath: "../path",
						},
					}
					err = validations.ValidateNsStoreNSFS(ns)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("SubPath ../path must not contain '..'"))
				})
				It("Should Allow", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "pvc-name",
							SubPath: "valid/sub/path",
						},
					}
					err = validations.ValidateNsStoreNSFS(ns)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
			Context("Validate too long mount path", func() {
				It("Should Deny", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "pvc-name",
							SubPath: "valid/sub/path",
						},
					}
					ns.Name = "nsfs-too-long-name-should-fail-after-exceeding-63-characters"
					mountPath := "/nsfs/" + ns.Name
					err = validations.ValidateNsStoreNSFS(ns)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal(fmt.Sprintf("MountPath %v must be no more than 63 characters", mountPath)))
				})
				It("Should Allow", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "pvc-name",
							SubPath: "valid/sub/path",
						},
					}
					ns.Name = "nsfs-not-too-long-name"
					err = validations.ValidateNsStoreNSFS(ns)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})

	Describe("Validate update operations", func() {
		var (
			updatedNS *nbv1.NamespaceStore
		)

		BeforeEach(func() {
			updatedNS = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml).(*nbv1.NamespaceStore)
			updatedNS.Name = "ns-name"
			updatedNS.Namespace = "test"
		})

		Describe("Cloud namespacestore", func() {
			Context("Update target bucket", func() {
				It("Should Deny", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
						},
					}

					updatedNS.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-other-bucket",
						},
					}

					err = validations.ValidateTargetNSBucketChange(*ns, *updatedNS)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Changing a NamespaceStore target bucket is unsupported"))
				})

				It("Should Allow", func() {
					ns.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "same-target-bucket",
						},
					}

					updatedNS.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "same-target-bucket",
						},
					}

					err = validations.ValidateTargetNSBucketChange(*ns, *updatedNS)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})

var _ = Describe("BucketClass admission unit tests", func() {
	var (
		bc  *nbv1.BucketClass
		err error
	)

	BeforeEach(func() {
		bc = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml).(*nbv1.BucketClass)
		bc.Name = "bc-name"
		bc.Namespace = "test"
	})

	Describe("Validate create operations", func() {
		Context("Unsupported tiers number", func() {
			It("Should Deny", func() {
				bc.Spec.PlacementPolicy = &nbv1.PlacementPolicy{
					Tiers: []nbv1.Tier{{
						Placement:     "",
						BackingStores: []string{"bs-name"},
					}, {
						Placement:     "",
						BackingStores: []string{"bs-name"},
					}, {
						Placement:     "",
						BackingStores: []string{"bs-name"},
					}},
				}
				err = validations.ValidateTiersNumber(bc.Spec.PlacementPolicy.Tiers)
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("unsupported number of tiers, bucketclass supports only 1 or 2 tiers"))
			})
			It("Should Allow", func() {
				bc.Spec.PlacementPolicy = &nbv1.PlacementPolicy{
					Tiers: []nbv1.Tier{{
						Placement:     "",
						BackingStores: []string{"bs-name"},
					}, {
						Placement:     "",
						BackingStores: []string{"bs-name"},
					}},
				}
				err = validations.ValidateTiersNumber(bc.Spec.PlacementPolicy.Tiers)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("Validate quota", func() {
			It("Should Deny", func() {
				bc.Spec.Quota = &nbv1.Quota{
					MaxSize:    "2Gi",
					MaxObjects: "-1",
				}
				err = validations.ValidateQuotaConfig(bc.Name, bc.Spec.Quota)
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("ob \"bc-name\" validation error: invalid maxObjects value. O or any positive number "))
			})
			It("Should Deny", func() {
				bc.Spec.Quota = &nbv1.Quota{
					MaxSize:    "-1Gi",
					MaxObjects: "10",
				}
				err = validations.ValidateQuotaConfig(bc.Name, bc.Spec.Quota)
				Ω(err).Should(HaveOccurred())
				Expect(err.Error()).To(Equal("ob \"bc-name\" validation error: invalid obcMaxSizeValue value: min 1Gi, max 1023Pi, 0 to remove quota"))
			})
			It("Should Allow", func() {
				bc.Spec.Quota = &nbv1.Quota{
					MaxSize:    "20Gi",
					MaxObjects: "10",
				}
				err = validations.ValidateQuotaConfig(bc.Name, bc.Spec.Quota)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
	})
})

var _ = Describe("NooBaaAccount admission unit tests", func() {

	var (
		na  *nbv1.NooBaaAccount
		err error
	)

	BeforeEach(func() {
		na = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml).(*nbv1.NooBaaAccount)
		na.Name = "na-name"
		na.Namespace = "test"
	})

	Describe("Validate create operations", func() {
		Describe("Noobaaaccount NSFS create validations", func() {
			Context("UID and GID are a whole positive number", func() {
				It("Should Deny Negative UID", func() {
					na.Spec = nbv1.NooBaaAccountSpec{
						NsfsAccountConfig: &nbv1.AccountNsfsConfig{
							UID:            -3,
							GID:            2,
							NewBucketsPath: "/",
							NsfsOnly:       true,
						},
					}
					err = validations.ValidateNSFSConfig(*na)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("UID must be a whole positive number"))
				})
				It("Should Deny Negative GID", func() {
					na.Spec = nbv1.NooBaaAccountSpec{
						NsfsAccountConfig: &nbv1.AccountNsfsConfig{
							UID:            3,
							GID:            -2,
							NewBucketsPath: "/",
							NsfsOnly:       true,
						},
					}
					err = validations.ValidateNSFSConfig(*na)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("GID must be a whole positive number"))
				})
				It("Should Allow", func() {
					na.Spec = nbv1.NooBaaAccountSpec{
						NsfsAccountConfig: &nbv1.AccountNsfsConfig{
							UID:            3,
							GID:            2,
							NewBucketsPath: "/",
							NsfsOnly:       true,
						},
					}
					err = validations.ValidateNSFSConfig(*na)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})

	Describe("Validate update operations", func() {
		var (
			updatedNA *nbv1.NooBaaAccount
		)

		BeforeEach(func() {
			updatedNA = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml).(*nbv1.NooBaaAccount)
			updatedNA.Name = "na-name"
			updatedNA.Namespace = "test"
		})

		Describe("Noobaaaccount NSFS update validations", func() {
			Context("Remove NSFSAccountConfig from NooBaaAccountSpec", func() {
				It("Should Deny", func() {
					na.Spec = nbv1.NooBaaAccountSpec{
						AllowBucketCreate: true,
						NsfsAccountConfig: &nbv1.AccountNsfsConfig{
							UID:            3,
							GID:            2,
							NewBucketsPath: "/",
							NsfsOnly:       true,
						},
					}

					updatedNA.Spec = nbv1.NooBaaAccountSpec{}

					err = validations.ValidateRemoveNSFSConfig(*updatedNA, *na)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Removing the NsfsAccountConfig is unsupported"))
				})
			})
			Context("Update NSFSAccountConfig In NooBaaAccountSpec", func() {
				It("Should Allow", func() {
					na.Spec = nbv1.NooBaaAccountSpec{
						NsfsAccountConfig: &nbv1.AccountNsfsConfig{
							UID:            3,
							GID:            2,
							NewBucketsPath: "/",
							NsfsOnly:       true,
						},
					}

					updatedNA.Spec = nbv1.NooBaaAccountSpec{
						NsfsAccountConfig: &nbv1.AccountNsfsConfig{
							UID:            30,
							GID:            20,
							NewBucketsPath: "/new/",
							NsfsOnly:       false,
						},
					}

					err = validations.ValidateRemoveNSFSConfig(*na, *updatedNA)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})

var _ = Describe("Noobaa admission unit tests", func() {

	var (
		nb  *nbv1.NooBaa
		err error
	)

	BeforeEach(func() {
		nb = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa)
		nb.Name = "noobaa"
		nb.Namespace = "test"

	})

	Describe("Validate delete operations", func() {
		Describe("General noobaa validations", func() {
			Context("cleanup policy not set", func() {
				It("Should Deny", func() {
					err = validations.ValidateNoobaaDeletion(*nb)
					Ω(err).Should(HaveOccurred())
					Expect(err.Error()).To(Equal("Noobaa cleanup policy is not set, blocking Noobaa deletion"))
				})
				It("Should Allow", func() {
					nb.Spec = nbv1.NooBaaSpec{
						CleanupPolicy: nbv1.CleanupPolicySpec{
							Confirmation:        "confirmed",
							AllowNoobaaDeletion: true,
						},
					}
					err = validations.ValidateNoobaaDeletion(*nb)
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})
