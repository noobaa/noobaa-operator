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

// NewNoobaaValidator initializes a NoobaaValidator to be used for loading and validating a noobaa resource
func NewNoobaaValidator(arRequest admissionv1.AdmissionReview) *ResourceValidator {
	nv := &ResourceValidator{
		Logger:    logrus.WithField("admission noobaa validation", arRequest.Request.Namespace),
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
	return nv
}

// ValidateNoobaa call appropriate validations based on the operation
func (nv *ResourceValidator) ValidateNoobaa() admissionv1.AdmissionReview {
	switch nv.arRequest.Request.Operation {
	case admissionv1.Create:
		nv.ValidateNoobaaCreate()
	case admissionv1.Update:
		nv.ValidateNoobaaUpdate()
	case admissionv1.Delete:
		nv.ValidateDeleteNoobaa()
	default:
		nv.Logger.Errorf("No action registered for the operation type: %v", nv.arRequest.Request.Operation)
	}

	return *nv.arResponse
}

// DeserializeNB extract noobaa cr from the request
func (nv *ResourceValidator) DeserializeNB(rawNB []byte) *nbv1.NooBaa {
	NB := nbv1.NooBaa{}
	if err := json.Unmarshal(rawNB, &NB); err != nil {
		nv.Logger.Error("error deserializing noobaa")
	}
	return &NB
}

// ValidateDeleteNoobaa runs all the validations tests for DELETE operations
func (nv *ResourceValidator) ValidateDeleteNoobaa() {
	noobaaCR := nv.DeserializeNB(nv.arRequest.Request.OldObject.Raw)
	if noobaaCR == nil {
		nv.SetValidationResult(false, "failed deserializing noobaa")
		return
	}

	if err := validations.ValidateNoobaaDeletion(*noobaaCR); err != nil && util.IsValidationError(err) {
		nv.SetValidationResult(false, err.Error())
		return
	}
}

// ValidateNoobaaUpdate runs all the validations tests for UPDATE operations
func (nv *ResourceValidator) ValidateNoobaaUpdate() {
	noobaaCR := nv.DeserializeNB(nv.arRequest.Request.Object.Raw)
	if noobaaCR == nil {
		nv.SetValidationResult(false, "failed deserializing noobaa")
		return
	}

	if err := validations.ValidateNoobaaUpdate(*noobaaCR); err != nil && util.IsValidationError(err) {
		nv.SetValidationResult(false, err.Error())
		return
	}
}

// ValidateNoobaaCreate runs all the validations tests for CREATE operations
func (nv *ResourceValidator) ValidateNoobaaCreate() {
	noobaaCR := nv.DeserializeNB(nv.arRequest.Request.Object.Raw)
	if noobaaCR == nil {
		nv.SetValidationResult(false, "failed deserializing noobaa")
		return
	}

	if err := validations.ValidateNoobaaCreation(*noobaaCR); err != nil && util.IsValidationError(err) {
		nv.SetValidationResult(false, err.Error())
		return
	}
}
