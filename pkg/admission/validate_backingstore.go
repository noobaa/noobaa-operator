package admission

import (
	"encoding/json"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewBackingStoreValidator initializes a BackingStoreValidator to be used for loading and validating a backingstore
func NewBackingStoreValidator(arRequest admissionv1.AdmissionReview) *ResourceValidator {
	bsv := &ResourceValidator{
		Logger:    logrus.WithField("admission backingstore validation", arRequest.Request.Namespace),
		arRequest: &arRequest,
		arResponse: &admissionv1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
			Response: &admissionv1.AdmissionResponse{
				UID:     arRequest.Request.UID,
				Allowed: true,
				Result: &metav1.Status{
					Message: "allowed",
				},
			},
		},
	}
	return bsv
}

// ValidateBackingstore call appropriate validations based on the operation
func (bsv *ResourceValidator) ValidateBackingstore() admissionv1.AdmissionReview {
	switch bsv.arRequest.Request.Operation {
	case admissionv1.Create:
		bsv.ValidateCreateBS()
	case admissionv1.Update:
		bsv.ValidateUpdateBS()
	case admissionv1.Delete:
		bsv.ValidateDeleteBS()
	default:
		bsv.Logger.Error("Failed to identify backindstore operation type")
	}
	return *bsv.arResponse
}

// DeserializeBS extract the backingstore object from the request
func (bsv *ResourceValidator) DeserializeBS(rawBS []byte) *nbv1.BackingStore {
	BS := nbv1.BackingStore{}
	if err := json.Unmarshal(rawBS, &BS); err != nil {
		bsv.Logger.Error("error deserializing backingstore")
	}
	return &BS
}

// ValidateCreateBS runs all the validations tests for CREATE operations
func (bsv *ResourceValidator) ValidateCreateBS() {
	bs := bsv.DeserializeBS(bsv.arRequest.Request.Object.Raw)
	if bs == nil {
		return
	}

	if err := validations.ValidateBackingStore(*bs); err != nil && util.IsValidationError(err) {
		bsv.SetValidationResult(false, err.Error())
		return
	}
}

// ValidateUpdateBS runs all the validations tests for UPDATE operations
func (bsv *ResourceValidator) ValidateUpdateBS() {
	bs := bsv.DeserializeBS(bsv.arRequest.Request.Object.Raw)
	oldBS := bsv.DeserializeBS(bsv.arRequest.Request.OldObject.Raw)

	if bs == nil || oldBS == nil {
		return
	}

	if err := validations.ValidateBackingStore(*bs); err != nil && util.IsValidationError(err) {
		bsv.SetValidationResult(false, err.Error())
		return
	}

	switch bs.Spec.Type {
	case nbv1.StoreTypePVPool:
		if err := validations.ValidatePvpoolScaleDown(*bs, *oldBS); err != nil && util.IsValidationError(err) {
			bsv.SetValidationResult(false, err.Error())
			return
		}
	case nbv1.StoreTypeAWSS3, nbv1.StoreTypeIBMCos, nbv1.StoreTypeAzureBlob, nbv1.StoreTypeGoogleCloudStorage:
		if err := validations.ValidateTargetBSBucketChange(*bs, *oldBS); err != nil && util.IsValidationError(err) {
			bsv.SetValidationResult(false, err.Error())
			return
		}
	}
}

// ValidateDeleteBS runs all the validations tests for DELETE operations
func (bsv *ResourceValidator) ValidateDeleteBS() {
	bs := bsv.DeserializeBS(bsv.arRequest.Request.OldObject.Raw)
	if bs == nil {
		return
	}

	sysClient, err := system.Connect(false)
	if err != nil {
		return
	}
	systemInfo, err := sysClient.NBClient.ReadSystemAPI()
	if err != nil {
		return
	}

	if err := validations.ValidateBackingstoreDeletion(*bs, systemInfo); err != nil && util.IsValidationError(err) {
		bsv.SetValidationResult(false, err.Error())
		return
	}
}
