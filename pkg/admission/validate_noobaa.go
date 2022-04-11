package admission

import (
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
	case admissionv1.Delete:
		nv.ValidateDeleteNoobaa()
	default:
		nv.Logger.Errorf("No action registered for the operation type: %v", nv.arRequest.Request.Operation)
	}

	return *nv.arResponse
}

// ValidateDeleteNoobaa runs all the validations tests for DELETE operations
func (nv *ResourceValidator) ValidateDeleteNoobaa() {
	nv.SetValidationResult(false, "Deletion of NooBaa resource is prohibited")
}
