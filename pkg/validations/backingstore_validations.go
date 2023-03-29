package validations

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"k8s.io/apimachinery/pkg/api/resource"
)

// const configuration values for the validation checks
const (
	MinimumVolumeSize       int64 = 16 * 1024 * 1024 * 1024 // 16Gi
	MaximumPvpoolNameLength       = 43
	MaximumVolumeCount            = 20
)

// ValidateBackingStore validates create validations on resource Backinstore
func ValidateBackingStore(bs nbv1.BackingStore) error {
	if err := ValidateBSEmptySecretName(bs); err != nil {
		return err
	}
	if err := ValidateBSEmptyTargetBucket(bs); err != nil {
		return err
	}
	switch bs.Spec.Type {
	case nbv1.StoreTypePVPool:
		if err := ValidatePvpoolNameLength(bs); err != nil {
			return err
		}
		if err := ValidatePvpoolMinVolSize(bs); err != nil {
			return err
		}
		if err := ValidateMaxVolumeCount(bs); err != nil {
			return err
		}
		if err := ValidateMinVolumeCount(bs); err != nil {
			return err
		}
	case nbv1.StoreTypeS3Compatible:
		return ValidateSigVersion(bs.Spec.S3Compatible.SignatureVersion)
	case nbv1.StoreTypeIBMCos:
		return ValidateSigVersion(bs.Spec.IBMCos.SignatureVersion)
	case nbv1.StoreTypeAWSS3, nbv1.StoreTypeAzureBlob, nbv1.StoreTypeGoogleCloudStorage:
		return nil
	default:
		return util.ValidationError{
			Msg: "Invalid Backingstore type, please provide a valid Backingstore type",
		}
	}
	return nil
}

// ValidateBSEmptySecretName validates a secret name is provided for cloud backingstores
func ValidateBSEmptySecretName(bs nbv1.BackingStore) error {
	switch bs.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		if len(bs.Spec.AWSS3.Secret.Name) == 0 {
			if err := ValidateAWSSTSARN(bs); err != nil {
				return err
			}
		}
	case nbv1.StoreTypeS3Compatible:
		if len(bs.Spec.S3Compatible.Secret.Name) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide secret name",
			}
		}
	case nbv1.StoreTypeIBMCos:
		if len(bs.Spec.IBMCos.Secret.Name) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide secret name",
			}
		}
	case nbv1.StoreTypeAzureBlob:
		if len(bs.Spec.AzureBlob.Secret.Name) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide secret name",
			}
		}
	case nbv1.StoreTypeGoogleCloudStorage:
		if len(bs.Spec.GoogleCloudStorage.Secret.Name) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide secret name",
			}
		}
	case nbv1.StoreTypePVPool:
		break
	default:
		return util.ValidationError{
			Msg: "Invalid Backingstore type, please provide a valid Backingstore type",
		}
	}
	return nil
}

// ValidateBSEmptyTargetBucket validates a target bucket name is provided for cloud backingstores
func ValidateBSEmptyTargetBucket(bs nbv1.BackingStore) error {
	switch bs.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		if len(bs.Spec.AWSS3.TargetBucket) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide target bucket",
			}
		}
	case nbv1.StoreTypeS3Compatible:
		if len(bs.Spec.S3Compatible.TargetBucket) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide target bucket",
			}
		}
	case nbv1.StoreTypeIBMCos:
		if len(bs.Spec.IBMCos.TargetBucket) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide target bucket",
			}
		}
	case nbv1.StoreTypeAzureBlob:
		if len(bs.Spec.AzureBlob.TargetBlobContainer) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide target bucket",
			}
		}
	case nbv1.StoreTypeGoogleCloudStorage:
		if len(bs.Spec.GoogleCloudStorage.TargetBucket) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide target bucket",
			}
		}
	case nbv1.StoreTypePVPool:
		break
	default:
		return util.ValidationError{
			Msg: "Invalid Backingstore type, please provide a valid Backingstore type",
		}
	}
	return nil
}

// ValidateAWSSTSARN validates the existence of the AWS STS ARN string
func ValidateAWSSTSARN(bs nbv1.BackingStore) error {
	if bs.Spec.AWSS3 != nil {
		if bs.Spec.AWSS3.AWSSTSRoleARN == nil {
			return util.ValidationError{
				Msg: "Failed creating the Backingstore, please provide a valid ARN or secret name",
			}
		}
	}
	return nil
}

// ValidatePvpoolNameLength validates the name of pvpool backingstore is under 43 characters
func ValidatePvpoolNameLength(bs nbv1.BackingStore) error {
	if len(bs.Name) > MaximumPvpoolNameLength {
		return util.ValidationError{
			Msg: "Unsupported BackingStore name length, please provide a name shorter then 43 characters",
		}
	}
	return nil
}

// ValidatePvpoolMinVolSize validates pvpool volume size is above 16GB
func ValidatePvpoolMinVolSize(bs nbv1.BackingStore) error {
	if bs.Spec.PVPool.VolumeResources == nil {
		return nil
	}
	min := *resource.NewQuantity(int64(MinimumVolumeSize), resource.BinarySI)
	if bs.Spec.PVPool.VolumeResources.Requests.Storage().Cmp(min) == -1 {
		return util.ValidationError{
			Msg: "Invalid volume size, minimum volume size is 16Gi",
		}
	}
	return nil
}

// ValidateSigVersion validates backingstore provided with a supported signature version
func ValidateSigVersion(sigver nbv1.S3SignatureVersion) error {
	switch sigver {
	case nbv1.S3SignatureVersionV2, nbv1.S3SignatureVersionV4, "":
		return nil
	default:
		return util.ValidationError{
			Msg: "Invalid S3 compatible Backingstore signature version, please choose either v2/v4",
		}
	}
}

// ValidateMaxVolumeCount validates pvpool backingstore volume count is under 20
func ValidateMaxVolumeCount(bs nbv1.BackingStore) error {
	if bs.Spec.PVPool.NumVolumes > MaximumVolumeCount {
		return util.ValidationError{
			Msg: "Unsupported volume count, the maximum supported volume count is 20",
		}
	}
	return nil
}

// ValidateMinVolumeCount validates pvpool backingstore volume count is above 0
func ValidateMinVolumeCount(bs nbv1.BackingStore) error {
	if bs.Spec.PVPool.NumVolumes <= 0 {
		return util.ValidationError{
			Msg: "Unsupported volume count, the minimum supported volume count is 1",
		}
	}
	return nil
}

// ValidatePvpoolScaleDown validates an operation of scaling down node in pvpool backingstore
func ValidatePvpoolScaleDown(bs nbv1.BackingStore, oldBs nbv1.BackingStore) error {
	if oldBs.Spec.PVPool.NumVolumes > bs.Spec.PVPool.NumVolumes {
		return util.ValidationError{
			Msg: "Scaling down the number of nodes is not currently supported",
		}
	}
	return nil
}

// ValidateTargetBSBucketChange validates the user is not trying to update the backingstore target bucket
func ValidateTargetBSBucketChange(bs nbv1.BackingStore, oldBs nbv1.BackingStore) error {
	switch bs.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		if oldBs.Spec.AWSS3.TargetBucket != bs.Spec.AWSS3.TargetBucket {
			return util.ValidationError{
				Msg: "Changing a Backingstore target bucket is unsupported",
			}
		}
	case nbv1.StoreTypeS3Compatible:
		if oldBs.Spec.S3Compatible.TargetBucket != bs.Spec.S3Compatible.TargetBucket {
			return util.ValidationError{
				Msg: "Changing a Backingstore target bucket is unsupported",
			}
		}
	case nbv1.StoreTypeIBMCos:
		if oldBs.Spec.IBMCos.TargetBucket != bs.Spec.IBMCos.TargetBucket {
			return util.ValidationError{
				Msg: "Changing a Backingstore target bucket is unsupported",
			}
		}
	case nbv1.StoreTypeAzureBlob:
		if oldBs.Spec.AzureBlob.TargetBlobContainer != bs.Spec.AzureBlob.TargetBlobContainer {
			return util.ValidationError{
				Msg: "Changing a Backingstore target bucket is unsupported",
			}
		}
	case nbv1.StoreTypeGoogleCloudStorage:
		if oldBs.Spec.GoogleCloudStorage.TargetBucket != bs.Spec.GoogleCloudStorage.TargetBucket {
			return util.ValidationError{
				Msg: "Changing a Backingstore target bucket is unsupported",
			}
		}
	default:
		return util.ValidationError{
			Msg: "Failed to identify Backingstore type",
		}
	}
	return nil
}

// ValidateBackingstoreDeletion validates the deleted backingstore not containing data buckets
func ValidateBackingstoreDeletion(bs nbv1.BackingStore, systemInfo nb.SystemInfo) error {
	for _, pool := range systemInfo.Pools {
		if pool.Name == bs.Name {
			if pool.Undeletable == "IS_BACKINGSTORE" || pool.Undeletable == "BEING_DELETED" {
				return nil
			} else if pool.Undeletable == "CONNECTED_BUCKET_DELETING" {
				return util.ValidationError{
					Msg: fmt.Sprintf("cannot complete because objects in Backingstore %q are still being deleted, Please try later", pool.Name),
				}
			}
			return util.ValidationError{
				Msg: fmt.Sprintf("cannot complete because pool %q in %q state", pool.Name, pool.Undeletable),
			}
		}
	}

	return nil
}
