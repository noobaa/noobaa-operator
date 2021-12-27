package admission

import (
	"encoding/json"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BucketClassValidator is the context of a bucketclass validation
type BucketClassValidator struct {
	Logger     *logrus.Entry
	arRequest  *admissionv1.AdmissionReview
	arResponse *admissionv1.AdmissionReview
}

// NewBucketClassValidator initializes a BucketClassValidator to be used for loading and validating a bucketclass
func NewBucketClassValidator(arRequest admissionv1.AdmissionReview) *BucketClassValidator {
	bcv := &BucketClassValidator{
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
func (bcv *BucketClassValidator) ValidateBucketClass() admissionv1.AdmissionReview {
	switch bcv.arRequest.Request.Operation {
	case admissionv1.Create:
		bcv.ValidateCreate()
	default:
		bcv.Logger.Error("Failed to identify bucketclass operation type")
	}
	return *bcv.arResponse
}

// SetValidationResult responsible of assinging the return values of a validation into the response appropriate fields
func (bcv *BucketClassValidator) SetValidationResult(isAllowed bool, message string) {
	bcv.arResponse.Response.Allowed = isAllowed
	bcv.arResponse.Response.Result.Message = message
}

// DeserializeBC extract the bucketclass object from the request
func (bcv *BucketClassValidator) DeserializeBC(rawBS []byte) *nbv1.BucketClass {
	BC := nbv1.BucketClass{}
	if err := json.Unmarshal(rawBS, &BC); err != nil {
		bcv.Logger.Error("error deserializing bucketclass")
	}
	return &BC
}

// ValidateCreate runs all the validations tests for CREATE operations
func (bcv *BucketClassValidator) ValidateCreate() {
	bc := bcv.DeserializeBC(bcv.arRequest.Request.Object.Raw)
	if bc == nil {
		return
	}

	if err := bucketclass.ValidateQuotaConfig(bc.Name, bc.Spec.Quota); err != nil && util.IsValidationError(err) {
		bcv.SetValidationResult(false, err.Error())
		return
	}
	if bc.Spec.NamespacePolicy != nil {
		if err := bucketclass.ValidateNSFSSingleBC(bc); err != nil && util.IsValidationError(err) {
			bcv.SetValidationResult(false, err.Error())
			return
		}
	}
	if bc.Spec.PlacementPolicy != nil {
		if err := bucketclass.ValidateTiersNumber(bc.Spec.PlacementPolicy.Tiers); err != nil {
			bcv.SetValidationResult(false, err.Error())
			return
		}
	}
}
