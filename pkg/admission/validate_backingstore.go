package admission

import (
	"encoding/json"
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// const configuration values for the validation checks
const (
	MinimumVolumeSize       = 16
	MaximumPvpoolNameLength = 43
	MaximumVolumeCount      = 20
)

// BackingStoreValidator is the context of a backingstore validation
type BackingStoreValidator struct {
	BackingStore *nbv1.BackingStore
	Logger       *logrus.Entry
	arRequest    *admissionv1.AdmissionReview
	arResponse   *admissionv1.AdmissionReview
}

// NewBackingStoreValidator initializes a BackingStoreValidator to be used for loading and validating a backingstore
func NewBackingStoreValidator(arRequest admissionv1.AdmissionReview) *BackingStoreValidator {
	bsv := &BackingStoreValidator{
		arRequest: &arRequest,
		Logger:    logrus.WithField("admission backingstore validation", arRequest.Request.Namespace),
	}
	return bsv
}

// ValidateBackingstore call appropriate validations based on the operation
func (bsv *BackingStoreValidator) ValidateBackingstore() admissionv1.AdmissionReview {
	bsv.arResponse = &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     bsv.arRequest.Request.UID,
			Allowed: true,
			Result: &metav1.Status{
				Message: "allowed",
			},
		},
	}

	switch bsv.arRequest.Request.Operation {
	case admissionv1.Create:
		bsv.ValidateCreate()
	case admissionv1.Update:
		bsv.ValidateUpdate()
	case admissionv1.Delete:
		bsv.ValidateDelete()
	default:
		bsv.Logger.Error("Failed to identify backindstore operation type")
	}
	return *bsv.arResponse
}

// SetValidationResult responsible of assinging the return values of a validation into the response appropriate fields
func (bsv *BackingStoreValidator) SetValidationResult(isAllowed bool, message string) {
	bsv.arResponse.Response.Allowed = isAllowed
	bsv.arResponse.Response.Result.Message = message
}

// DeserializeBS extract the backingstore object from the request
func (bsv *BackingStoreValidator) DeserializeBS(rawBS []byte) *nbv1.BackingStore {
	BS := nbv1.BackingStore{}
	if err := json.Unmarshal(rawBS, &BS); err != nil {
		bsv.Logger.Error("error deserializing backingstore")
	}
	return &BS
}

// ValidateCreate runs all the validations tests for CREATE operations
func (bsv *BackingStoreValidator) ValidateCreate() {
	bsv.BackingStore = bsv.DeserializeBS(bsv.arRequest.Request.Object.Raw)
	if ok, message := bsv.ValidateBackingStoreType(); !ok {
		bsv.SetValidationResult(ok, message)
		return
	}

	if ok, message := bsv.ValidateBSEmptySecretName(); !ok {
		bsv.SetValidationResult(ok, message)
		return
	}

	switch bsv.BackingStore.Spec.Type {
	case nbv1.StoreTypePVPool:
		if ok, message := bsv.ValidatePvpoolNameLength(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := bsv.ValidateMinVolumeCount(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := bsv.ValidateMaxVolumeCount(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := bsv.ValidatePvpoolMinVolSize(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
	case nbv1.StoreTypeS3Compatible:
		if ok, message := bsv.ValidateS3CompatibleSigVersion(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
	}
}

// ValidateUpdate runs all the validations tests for UPDATE operations
func (bsv *BackingStoreValidator) ValidateUpdate() {
	bsv.BackingStore = bsv.DeserializeBS(bsv.arRequest.Request.Object.Raw)
	oldBS := bsv.DeserializeBS(bsv.arRequest.Request.OldObject.Raw)

	if ok, message := bsv.ValidateBackingStoreType(); !ok {
		bsv.SetValidationResult(ok, message)
		return
	}

	if ok, message := bsv.ValidateBSEmptySecretName(); !ok {
		bsv.SetValidationResult(ok, message)
		return
	}

	switch bsv.BackingStore.Spec.Type {
	case nbv1.StoreTypePVPool:
		if ok, message := bsv.ValidatePvpoolScaleDown(*oldBS); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := bsv.ValidatePvpoolNameLength(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := bsv.ValidateMinVolumeCount(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := bsv.ValidateMaxVolumeCount(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := bsv.ValidatePvpoolMinVolSize(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
	case nbv1.StoreTypeAWSS3, nbv1.StoreTypeIBMCos, nbv1.StoreTypeAzureBlob, nbv1.StoreTypeGoogleCloudStorage:
		if ok, message := bsv.ValidateTargetBucketChange(*oldBS); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
	case nbv1.StoreTypeS3Compatible:
		if ok, message := bsv.ValidateS3CompatibleSigVersion(); !ok {
			bsv.SetValidationResult(ok, message)
			return
		}
	}
}

// ValidateDelete runs all the validations tests for DELETE operations
func (bsv *BackingStoreValidator) ValidateDelete() {
	bsv.BackingStore = bsv.DeserializeBS(bsv.arRequest.Request.OldObject.Raw)

	if ok, message := bsv.ValidateBackingstoreDeletion(); !ok {
		bsv.SetValidationResult(ok, message)
		return
	}
}

// ValidateBSEmptySecretName validates a secret name is provided for cloud backingstores
func (bsv *BackingStoreValidator) ValidateBSEmptySecretName() (bool, string) {
	switch bsv.BackingStore.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		if len(bsv.BackingStore.Spec.AWSS3.Secret.Name) == 0 {
			return false, "Failed creating the Backingstore, please provide secret name"
		}
	case nbv1.StoreTypeS3Compatible:
		if len(bsv.BackingStore.Spec.S3Compatible.Secret.Name) == 0 {
			return false, "Failed creating the Backingstore, please provide secret name"
		}
	case nbv1.StoreTypeIBMCos:
		if len(bsv.BackingStore.Spec.IBMCos.Secret.Name) == 0 {
			return false, "Failed creating the Backingstore, please provide secret name"
		}
	case nbv1.StoreTypeAzureBlob:
		if len(bsv.BackingStore.Spec.AzureBlob.Secret.Name) == 0 {
			return false, "Failed creating the Backingstore, please provide secret name"
		}
	case nbv1.StoreTypeGoogleCloudStorage:
		if len(bsv.BackingStore.Spec.GoogleCloudStorage.Secret.Name) == 0 {
			return false, "Failed creating the Backingstore, please provide secret name"
		}
	case nbv1.StoreTypePVPool:
		break
	default:
		return false, "Failed to identify Backingstore type"
	}
	return true, "allowed"
}

// ValidateBackingStoreType validates a supported backingstore type
func (bsv *BackingStoreValidator) ValidateBackingStoreType() (bool, string) {
	switch bsv.BackingStore.Spec.Type {
	case nbv1.StoreTypeAWSS3, nbv1.StoreTypeS3Compatible, nbv1.StoreTypeIBMCos, nbv1.StoreTypeAzureBlob, nbv1.StoreTypeGoogleCloudStorage, nbv1.StoreTypePVPool:
		return true, "allowed"
	default:
		return false, "Invalid Backingstore type, please provide a valid Backingstore type"
	}
}

// ValidatePvpoolNameLength validates the name of pvpool backingstore is under 43 characters
func (bsv *BackingStoreValidator) ValidatePvpoolNameLength() (bool, string) {
	if len(bsv.BackingStore.Name) > MaximumPvpoolNameLength {
		return false, "Unsupported BackingStore name length, please provide a name shorter then 43 characters"
	}
	return true, "allowed"
}

// ValidatePvpoolMinVolSize validates pvpool volume size is above 16GB
func (bsv *BackingStoreValidator) ValidatePvpoolMinVolSize() (bool, string) {
	min := *resource.NewScaledQuantity(int64(MinimumVolumeSize), resource.Giga)
	if bsv.BackingStore.Spec.PVPool.VolumeResources.Requests.Storage().Cmp(min) == -1 {
		return false, "Invalid volume size, minimum volume size is 16Gi"
	}
	return true, "allowed"
}

// ValidateS3CompatibleSigVersion validates S3 Compatible backingstore provided with a supported signature viersion
func (bsv *BackingStoreValidator) ValidateS3CompatibleSigVersion() (bool, string) {
	switch bsv.BackingStore.Spec.S3Compatible.SignatureVersion {
	case nbv1.S3SignatureVersionV2, nbv1.S3SignatureVersionV4:
		return true, "allowed"
	default:
		return false, "Invalid S3 compatible Backingstore signature version, please choose either v2/v4"
	}
}

// ValidateMaxVolumeCount validates pvpool backingstore volume count is under 20
func (bsv *BackingStoreValidator) ValidateMaxVolumeCount() (bool, string) {
	if bsv.BackingStore.Spec.PVPool.NumVolumes > MaximumVolumeCount {
		return false, "Unsupported volume count, the maximum supported volume count is 20"
	}
	return true, "allowed"
}

// ValidateMinVolumeCount validates pvpool backingstore volume count is above 0
func (bsv *BackingStoreValidator) ValidateMinVolumeCount() (bool, string) {
	if bsv.BackingStore.Spec.PVPool.NumVolumes <= 0 {
		return false, "Unsupported volume count, the minimum supported volume count is 1"
	}
	return true, "allowed"
}

// ValidatePvpoolScaleDown validates an operation of scaling down node in pvpool backingstore
func (bsv *BackingStoreValidator) ValidatePvpoolScaleDown(oldBs nbv1.BackingStore) (bool, string) {
	if oldBs.Spec.PVPool.NumVolumes > bsv.BackingStore.Spec.PVPool.NumVolumes {
		return false, "Scaling down the number of nodes is not currently supported"
	}
	return true, "allowed"
}

// ValidateTargetBucketChange validates the user is not trying to update the backingstore target bucket
func (bsv *BackingStoreValidator) ValidateTargetBucketChange(oldBs nbv1.BackingStore) (bool, string) {
	switch bsv.BackingStore.Spec.Type {
	case nbv1.StoreTypeAWSS3:
		if oldBs.Spec.AWSS3.TargetBucket != bsv.BackingStore.Spec.AWSS3.TargetBucket {
			return false, "Changing a Backingstore target bucket is unsupported"
		}
	case nbv1.StoreTypeS3Compatible:
		if oldBs.Spec.S3Compatible.TargetBucket != bsv.BackingStore.Spec.S3Compatible.TargetBucket {
			return false, "Changing a Backingstore target bucket is unsupported"
		}
	case nbv1.StoreTypeIBMCos:
		if oldBs.Spec.IBMCos.TargetBucket != bsv.BackingStore.Spec.IBMCos.TargetBucket {
			return false, "Changing a Backingstore target bucket is unsupported"
		}
	case nbv1.StoreTypeAzureBlob:
		if oldBs.Spec.AzureBlob.TargetBlobContainer != bsv.BackingStore.Spec.AzureBlob.TargetBlobContainer {
			return false, "Changing a Backingstore target bucket is unsupported"
		}
	case nbv1.StoreTypeGoogleCloudStorage:
		if oldBs.Spec.GoogleCloudStorage.TargetBucket != bsv.BackingStore.Spec.GoogleCloudStorage.TargetBucket {
			return false, "Changing a Backingstore target bucket is unsupported"
		}
	default:
		return false, "Failed to identify Backingstore type"
	}
	return true, "allowed"
}

// ValidateBackingstoreDeletion validates the deleted backingstore not containing data buckets
func (bsv *BackingStoreValidator) ValidateBackingstoreDeletion() (bool, string) {
	sysClient, err := system.Connect(false)
	if err != nil {
		bsv.Logger.Error("Failed to load noobaa system connection info")
		return true, "allowed"
	}
	systemInfo, err := sysClient.NBClient.ReadSystemAPI()
	if err != nil {
		bsv.Logger.Error("Failed to call ReadSystemInfo API")
		return true, "allowed"
	}

	for _, pool := range systemInfo.Pools {
		if pool.Name == bsv.BackingStore.Name {
			if pool.Undeletable == "IS_BACKINGSTORE" || pool.Undeletable == "BEING_DELETED" {
				return true, "allowed"
			}
			return false, fmt.Sprintf("Cannot complete because pool %q in %q state", pool.Name, pool.Undeletable)
		}
	}

	return true, "allowed"
}
