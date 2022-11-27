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

// NewNamespaceStoreValidator initializes a BackingStoreValidator to be used for loading and validating a namespacestore
func NewNamespaceStoreValidator(arRequest admissionv1.AdmissionReview) *ResourceValidator {
	nsv := &ResourceValidator{

		Logger:    logrus.WithField("admission namespacestore validation", arRequest.Request.Namespace),
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
	return nsv
}

// ValidateNamespaceStore call appropriate validations based on the operation
func (nsv *ResourceValidator) ValidateNamespaceStore() admissionv1.AdmissionReview {
	switch nsv.arRequest.Request.Operation {
	case admissionv1.Create:
		nsv.ValidateCreateNS()
	case admissionv1.Update:
		nsv.ValidateUpdateNS()
	case admissionv1.Delete:
		nsv.ValidateDeleteNS()
	default:
		nsv.Logger.Error("Failed to identify namespacestore operation type")
	}
	return *nsv.arResponse
}

// DeserializeNS extract the namespacestore object from the request
func (nsv *ResourceValidator) DeserializeNS(rawNS []byte) *nbv1.NamespaceStore {
	NS := nbv1.NamespaceStore{}
	if err := json.Unmarshal(rawNS, &NS); err != nil {
		nsv.Logger.Error("error deserializing namespacestore")
	}
	return &NS
}

// ValidateCreateNS runs all the validations tests for CREATE operations
func (nsv *ResourceValidator) ValidateCreateNS() {
	ns := nsv.DeserializeNS(nsv.arRequest.Request.Object.Raw)
	if ns == nil {
		return
	}

	if err := validations.ValidateNamespaceStore(ns); err != nil && util.IsValidationError(err) {
		nsv.SetValidationResult(false, err.Error())
		return
	}
}

// ValidateUpdateNS runs all the validations tests for UPDATE operations
func (nsv *ResourceValidator) ValidateUpdateNS() {
	ns := nsv.DeserializeNS(nsv.arRequest.Request.Object.Raw)
	oldNS := nsv.DeserializeNS(nsv.arRequest.Request.OldObject.Raw)

	if ns == nil || oldNS == nil {
		return
	}

	if err := validations.ValidateNamespaceStore(ns); err != nil && util.IsValidationError(err) {
		nsv.SetValidationResult(false, err.Error())
		return
	}

	switch ns.Spec.Type {
	case nbv1.NSStoreTypeAWSS3, nbv1.NSStoreTypeS3Compatible, nbv1.NSStoreTypeIBMCos, nbv1.NSStoreTypeAzureBlob, nbv1.NSStoreTypeGoogleCloudStorage:
		if err := validations.ValidateTargetNSBucketChange(*ns, *oldNS); err != nil && util.IsValidationError(err) {
			nsv.SetValidationResult(false, err.Error())
			return
		}
	}
}

// ValidateDeleteNS runs all the validations tests for DELETE operations
func (nsv *ResourceValidator) ValidateDeleteNS() {
	ns := nsv.DeserializeNS(nsv.arRequest.Request.OldObject.Raw)
	if ns == nil {
		return
	}
	sysClient, err := system.Connect(false)
	if err != nil {
		nsv.Logger.Errorf("failed to load noobaa system connection info")
		return
	}
	systemInfo, err := sysClient.NBClient.ReadSystemAPI()
	if err != nil {
		nsv.Logger.Errorf("failed to call ReadSystemInfo API")
		return
	}

	if err := validations.ValidateNamespacestoreDeletion(*ns, systemInfo); err != nil && util.IsValidationError(err) {
		nsv.SetValidationResult(false, err.Error())
		return
	}
}
