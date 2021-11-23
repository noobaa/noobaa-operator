package admissionunittests

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/admission"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("BackingStore admission unit tests", func() {

	var (
		bsv          *admission.BackingStoreValidator
		backingstore *nbv1.BackingStore
		result       bool
		message      string
	)

	BeforeEach(func() {
		backingstore = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore)
		backingstore.Name = "bs-name"
		backingstore.Namespace = "test"
		bsv = &admission.BackingStoreValidator{
			BackingStore: backingstore,
		}
		result = false
		message = ""
	})

	Describe("Validate create operations", func() {
		Describe("General backingstore validations", func() {
			Context("Empty secret name", func() {
				It("Should Deny", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "",
								Namespace: "test",
							},
						},
					}
					result, message = bsv.ValidateBSEmptySecretName()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Failed creating the Backingstore, please provide secret name"))
				})
				It("Should Allow", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "full-secret-name",
								Namespace: "test",
							},
						},
					}
					result, message = bsv.ValidateBSEmptySecretName()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
			Context("Invalid store type", func() {
				It("Should Deny", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: "invalid",
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					result, message = bsv.ValidateBackingStoreType()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Invalid Backingstore type, please provide a valid Backingstore type"))
				})
				It("Should Allow", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					result, message = bsv.ValidateBackingStoreType()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
		})
		Describe("Pvpool backingstore", func() {
			Context("Resource name too long", func() {
				It("Should Deny", func() {
					bsv.BackingStore.Name = "pvpool-too-long-name-should-fail-after-exceeding-43-characters"
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
					}

					result, message = bsv.ValidatePvpoolNameLength()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Unsupported BackingStore name length, please provide a name shorter then 43 characters"))
				})

				It("Should Allow", func() {
					bsv.BackingStore.Name = "pvpool-not-too-long-name"
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
					}

					result, message = bsv.ValidatePvpoolNameLength()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
			Context("Minimum volume count", func() {
				It("Should Deny", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: -5,
						},
					}

					result, message = bsv.ValidateMinVolumeCount()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Unsupported volume count, the minimum supported volume count is 1"))

				})
				It("Should Allow", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 5,
						},
					}

					result, message = bsv.ValidateMinVolumeCount()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))

				})
			})
			Context("Maximum volume count", func() {
				It("Should Deny", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 25,
						},
					}

					result, message = bsv.ValidateMaxVolumeCount()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Unsupported volume count, the maximum supported volume count is 20"))

				})
				It("Should Allow", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							NumVolumes: 15,
						},
					}

					result, message = bsv.ValidateMaxVolumeCount()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))

				})
			})
			Context("Minimum volume size", func() {
				It("Should Deny", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							VolumeResources: &corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: *resource.NewScaledQuantity(int64(5), resource.Giga),
								},
							},
						},
					}

					result, message = bsv.ValidatePvpoolMinVolSize()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Invalid volume size, minimum volume size is 16Gi"))
				})

				It("Should Allow", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypePVPool,
						PVPool: &nbv1.PVPoolSpec{
							VolumeResources: &corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: *resource.NewScaledQuantity(int64(20), resource.Giga),
								},
							},
						},
					}

					result, message = bsv.ValidatePvpoolMinVolSize()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
		})
		Describe("S3 Compatible backingstore", func() {
			Context("Invalid signature version", func() {
				It("Should Deny", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeS3Compatible,
						S3Compatible: &nbv1.S3CompatibleSpec{
							SignatureVersion: "v5",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					result, message = bsv.ValidateS3CompatibleSigVersion()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Invalid S3 compatible Backingstore signature version, please choose either v2/v4"))
				})

				It("Should Allow", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
						Type: nbv1.StoreTypeS3Compatible,
						S3Compatible: &nbv1.S3CompatibleSpec{
							SignatureVersion: "v4",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					result, message = bsv.ValidateS3CompatibleSigVersion()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
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
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
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

					result, message = bsv.ValidatePvpoolScaleDown(*updatedBS)
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Scaling down the number of nodes is not currently supported"))
				})

				It("Should Allow", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
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

					result, message = bsv.ValidatePvpoolScaleDown(*updatedBS)
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
		})
		Describe("Cloud backingstore", func() {
			Context("Update target bucket", func() {
				It("Should Deny", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
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

					result, message = bsv.ValidateTargetBucketChange(*updatedBS)
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Changing a Backingstore target bucket is unsupported"))
				})

				It("Should Allow", func() {
					bsv.BackingStore.Spec = nbv1.BackingStoreSpec{
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

					result, message = bsv.ValidateTargetBucketChange(*updatedBS)
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
		})
	})
})

var _ = Describe("NamespaceStore admission unit tests", func() {

	var (
		nsv            *admission.NamespaceStoreValidator
		namespacestore *nbv1.NamespaceStore
		result         bool
		message        string
	)

	BeforeEach(func() {
		namespacestore = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_namespacestore_cr_yaml).(*nbv1.NamespaceStore)
		namespacestore.Name = "ns-name"
		namespacestore.Namespace = "test"
		nsv = &admission.NamespaceStoreValidator{
			NamespaceStore: namespacestore,
		}
		result = false
		message = ""
	})

	Describe("Validate create operations", func() {
		Describe("General namespacestore validations", func() {
			Context("Empty secret name", func() {
				It("Should Deny", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "",
								Namespace: "test",
							},
						},
					}
					result, message = nsv.ValidateNSEmptySecretName()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Failed creating the namespacestore, please provide secret name"))
				})
				It("Should Allow", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}
					result, message = nsv.ValidateNSEmptySecretName()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
			Context("Invalid store type", func() {
				It("Should Deny", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: "invalid",
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					result, message = nsv.ValidateNamespaceStoreType()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Invalid namespacestore type, please provide a valid one"))
				})
				It("Should Allow", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeAWSS3,
						AWSS3: &nbv1.AWSS3Spec{
							TargetBucket: "some-target-bucket",
							Secret: corev1.SecretReference{
								Name:      "secret-name",
								Namespace: "test",
							},
						},
					}

					result, message = nsv.ValidateNamespaceStoreType()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
		})
		Describe("NSFS validations", func() {
			Context("Empty pvc name", func() {
				It("Should Deny", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "",
						},
					}
					result, message = nsv.ValidateEmptyPvcName()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Failed to create NSFS, please provide pvc name"))
				})
				It("Should Allow", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							PvcName: "pvc-name",
						},
					}
					result, message = nsv.ValidateEmptyPvcName()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
			Context("Invalid SubPath", func() {
				It("Should Deny", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							SubPath: "/path",
						},
					}
					result, message = nsv.ValidateSubPath()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Failed to create NSFS, SubPath must be a relative path"))
				})
				It("Should Deny", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							SubPath: "../path",
						},
					}
					result, message = nsv.ValidateSubPath()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Failed to create NSFS, SubPath must not contain '..'"))
				})
				It("Should Allow", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
						NSFS: &nbv1.NSFSSpec{
							SubPath: "valid/sub/path",
						},
					}
					result, message = nsv.ValidateSubPath()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
			Context("Validate too long mount path", func() {
				It("Should Deny", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
					}
					nsv.NamespaceStore.Name = "nsfs-too-long-name-should-fail-after-exceeding-63-characters"
					result, message = nsv.ValidateMountPath()
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Failed to create NSFS, MountPath must be no more than 63 characters"))
				})
				It("Should Allow", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
						Type: nbv1.NSStoreTypeNSFS,
					}
					nsv.NamespaceStore.Name = "nsfs-not-too-long-name"
					result, message = nsv.ValidateMountPath()
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
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
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
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

					result, message = nsv.ValidateTargetBucketChange(*updatedNS)
					Expect(result).To(BeFalse())
					Expect(message).To(Equal("Changing a NamespaceStore target bucket is unsupported"))
				})

				It("Should Allow", func() {
					nsv.NamespaceStore.Spec = nbv1.NamespaceStoreSpec{
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

					result, message = nsv.ValidateTargetBucketChange(*updatedNS)
					Expect(result).To(BeTrue())
					Expect(message).To(Equal("allowed"))
				})
			})
		})
	})
})
