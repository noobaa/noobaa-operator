package admission

import (
	"encoding/json"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewBucketClassValidator initializes a BucketClassValidator to be used for loading and validating a bucketclass
func NewBucketClassValidator(arRequest admissionv1.AdmissionReview) *ResourceValidator {
	bcv := &ResourceValidator{
		Logger:    logrus.WithField("admission bucketclass validation", arRequest.Request.Namespace),
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
	return bcv
}

// ValidateBucketClass call appropriate validations based on the operation
func (bcv *ResourceValidator) ValidateBucketClass() admissionv1.AdmissionReview {
	switch bcv.arRequest.Request.Operation {
	case admissionv1.Create:
		bcv.ValidateCreateBC()
	case admissionv1.Update:
		bcv.ValidateUpdateBC()
	default:
		bcv.Logger.Error("Failed to identify bucketclass operation type")
	}
	return *bcv.arResponse
}

// DeserializeBC extract the bucketclass object from the request
func (bcv *ResourceValidator) DeserializeBC(rawBS []byte) *nbv1.BucketClass {
	BC := nbv1.BucketClass{}
	if err := json.Unmarshal(rawBS, &BC); err != nil {
		bcv.Logger.Error("error deserializing bucketclass")
	}
	return &BC
}

// ValidateCreateBC runs all the validations tests for CREATE operations
func (bcv *ResourceValidator) ValidateCreateBC() {
	bc := bcv.DeserializeBC(bcv.arRequest.Request.Object.Raw)
	if bc == nil {
		return
	}

	if err := validations.ValidateQuotaConfig(bc.Name, bc.Spec.Quota); err != nil && util.IsValidationError(err) {
		bcv.SetValidationResult(false, err.Error())
		return
	}
	if bc.Spec.NamespacePolicy != nil {
		if err := validations.ValidateNSFSSingleBC(bc); err != nil && util.IsValidationError(err) {
			bcv.SetValidationResult(false, err.Error())
			return
		}
	}
	if bc.Spec.PlacementPolicy != nil {
		if err := validations.ValidateTiersNumber(bc.Spec.PlacementPolicy.Tiers); err != nil {
			bcv.SetValidationResult(false, err.Error())
			return
		}
	}
}

// ValidateUpdateBC runs all the validations tests for UPDATE operations
func (bcv *ResourceValidator) ValidateUpdateBC() {
	oldBC := bcv.DeserializeBC(bcv.arRequest.Request.OldObject.Raw)
	newBC := bcv.DeserializeBC(bcv.arRequest.Request.Object.Raw)

	if err := validations.ValidateImmutLabelChange(newBC, oldBC, map[string]struct{}{"noobaa-operator": {}}); err != nil {
		bcv.SetValidationResult(false, err.Error())
		return
	}
}
