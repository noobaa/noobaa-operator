package admission

import (
	"encoding/json"
	"reflect"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewNoobaaAccountValidator initializes a NoobaaAccountValidator to be used for loading and validating a noobaaaccount
func NewNoobaaAccountValidator(arRequest admissionv1.AdmissionReview) *ResourceValidator {
	nav := &ResourceValidator{

		Logger:    logrus.WithField("admission noobaaaccount validation", arRequest.Request.Namespace),
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
	return nav
}

// ValidateNoobaAaccount call appropriate validations based on the operation
func (nav *ResourceValidator) ValidateNoobaAaccount() admissionv1.AdmissionReview {
	switch nav.arRequest.Request.Operation {
	case admissionv1.Create:
		nav.ValidateCreateNA()
	case admissionv1.Update:
		nav.ValidateUpdateNA()
	default:
		nav.Logger.Error("Failed to identify noobaaaccount operation type")
	}
	return *nav.arResponse
}

// DeserializeNA extract the noobaaaccount object from the request
func (nav *ResourceValidator) DeserializeNA(rawNA []byte) *nbv1.NooBaaAccount {
	NA := nbv1.NooBaaAccount{}
	if err := json.Unmarshal(rawNA, &NA); err != nil {
		nav.Logger.Error("error deserializing noobaaaccount")
	}
	return &NA
}

// ValidateCreateNA runs all the validations tests for CREATE operations
func (nav *ResourceValidator) ValidateCreateNA() {
	na := nav.DeserializeNA(nav.arRequest.Request.Object.Raw)
	if na == nil {
		return
	}

	if err := validations.ValidateNSFSConfig(*na); err != nil && util.IsValidationError(err) {
		nav.SetValidationResult(false, err.Error())
		return
	}
}

// ValidateUpdateNA runs all the validations tests for UPDATE operations
func (nav *ResourceValidator) ValidateUpdateNA() {
	na := nav.DeserializeNA(nav.arRequest.Request.Object.Raw)
	oldNA := nav.DeserializeNA(nav.arRequest.Request.OldObject.Raw)

	if na == nil || oldNA == nil {
		return
	}

	if err := validations.ValidateRemoveNSFSConfig(*na, *oldNA); err != nil && util.IsValidationError(err) {
		nav.SetValidationResult(false, err.Error())
		return
	}

	if !reflect.DeepEqual(oldNA.Spec.NsfsAccountConfig, na.Spec.NsfsAccountConfig) {
		if err := validations.ValidateNSFSConfig(*na); err != nil && util.IsValidationError(err) {
			nav.SetValidationResult(false, err.Error())
			return
		}
	}	
}
