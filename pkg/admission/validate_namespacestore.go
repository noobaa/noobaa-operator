package admission

import (
	"encoding/json"
	"fmt"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// const configuration values for the validation checks
const (
	MaximumMountPathLength = 63
)

// NamespaceStoreValidator is the context of a namespacestore validation
type NamespaceStoreValidator struct {
	NamespaceStore *nbv1.NamespaceStore
	Logger         *logrus.Entry
	arRequest      *admissionv1.AdmissionReview
	arResponse     *admissionv1.AdmissionReview
}

// NewNamespaceStoreValidator initializes a BackingStoreValidator to be used for loading and validating a namespacestore
func NewNamespaceStoreValidator(arRequest admissionv1.AdmissionReview) *NamespaceStoreValidator {
	nsv := &NamespaceStoreValidator{
		arRequest: &arRequest,
		Logger:    logrus.WithField("admission namespacestore validation", arRequest.Request.Namespace),
	}
	return nsv
}

// ValidateNamespaceStore call appropriate validations based on the operation
func (nsv *NamespaceStoreValidator) ValidateNamespaceStore() admissionv1.AdmissionReview {
	nsv.arResponse = &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     nsv.arRequest.Request.UID,
			Allowed: true,
			Result: &metav1.Status{
				Message: "allowed",
			},
		},
	}

	switch nsv.arRequest.Request.Operation {
	case admissionv1.Create:
		nsv.ValidateCreate()
	case admissionv1.Update:
		nsv.ValidateUpdate()
	case admissionv1.Delete:
		nsv.ValidateDelete()
	default:
		nsv.Logger.Error("Failed to identify namespacestore operation type")
	}
	return *nsv.arResponse
}

// SetValidationResult responsible of assinging the return values of a validation into the response appropriate fields
func (nsv *NamespaceStoreValidator) SetValidationResult(isAllowed bool, message string) {
	nsv.arResponse.Response.Allowed = isAllowed
	nsv.arResponse.Response.Result.Message = message
}

// DeserializeNS extract the namespacestore object from the request
func (nsv *NamespaceStoreValidator) DeserializeNS(rawNS []byte) *nbv1.NamespaceStore {
	NS := nbv1.NamespaceStore{}
	if err := json.Unmarshal(rawNS, &NS); err != nil {
		nsv.Logger.Error("error deserializing namespacestore")
	}
	return &NS
}

// ValidateCreate runs all the validations tests for CREATE operations
func (nsv *NamespaceStoreValidator) ValidateCreate() {
	nsv.NamespaceStore = nsv.DeserializeNS(nsv.arRequest.Request.Object.Raw)

	if ok, message := nsv.ValidateNamespaceStoreType(); !ok {
		nsv.SetValidationResult(ok, message)
		return
	}

	if ok, message := nsv.ValidateNSEmptySecretName(); !ok {
		nsv.SetValidationResult(ok, message)
		return
	}

	switch nsv.NamespaceStore.Spec.Type {
	case nbv1.NSStoreTypeNSFS:
		if ok, message := nsv.ValidateMountPath(); !ok {
			nsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := nsv.ValidateEmptyPvcName(); !ok {
			nsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := nsv.ValidateSubPath(); !ok {
			nsv.SetValidationResult(ok, message)
			return
		}
	}
}

// ValidateUpdate runs all the validations tests for UPDATE operations
func (nsv *NamespaceStoreValidator) ValidateUpdate() {
	nsv.NamespaceStore = nsv.DeserializeNS(nsv.arRequest.Request.Object.Raw)
	oldNS := nsv.DeserializeNS(nsv.arRequest.Request.OldObject.Raw)

	if ok, message := nsv.ValidateNamespaceStoreType(); !ok {
		nsv.SetValidationResult(ok, message)
		return
	}

	if ok, message := nsv.ValidateNSEmptySecretName(); !ok {
		nsv.SetValidationResult(ok, message)
		return
	}

	switch nsv.NamespaceStore.Spec.Type {
	case nbv1.NSStoreTypeNSFS:
		if ok, message := nsv.ValidateMountPath(); !ok {
			nsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := nsv.ValidateEmptyPvcName(); !ok {
			nsv.SetValidationResult(ok, message)
			return
		}
		if ok, message := nsv.ValidateSubPath(); !ok {
			nsv.SetValidationResult(ok, message)
			return
		}
	case nbv1.NSStoreTypeAWSS3, nbv1.NSStoreTypeS3Compatible, nbv1.NSStoreTypeIBMCos, nbv1.NSStoreTypeAzureBlob:
		if ok, message := nsv.ValidateTargetBucketChange(*oldNS); !ok {
			nsv.SetValidationResult(ok, message)
			return
		}
	}
}

// ValidateDelete runs all the validations tests for DELETE operations
func (nsv *NamespaceStoreValidator) ValidateDelete() {
	nsv.NamespaceStore = nsv.DeserializeNS(nsv.arRequest.Request.OldObject.Raw)

	if ok, message := nsv.ValidateNamespacestoreDeletion(); !ok {
		nsv.SetValidationResult(ok, message)
		return
	}
}

// ValidateNSEmptySecretName validates a secret name is provided for cloud namespacestore
func (nsv *NamespaceStoreValidator) ValidateNSEmptySecretName() (bool, string) {
	switch nsv.NamespaceStore.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		if len(nsv.NamespaceStore.Spec.AWSS3.Secret.Name) == 0 {
			return false, "Failed creating the namespacestore, please provide secret name"
		}
	case nbv1.NSStoreTypeS3Compatible:
		if len(nsv.NamespaceStore.Spec.S3Compatible.Secret.Name) == 0 {
			return false, "Failed creating the namespacestore, please provide secret name"
		}
	case nbv1.NSStoreTypeIBMCos:
		if len(nsv.NamespaceStore.Spec.IBMCos.Secret.Name) == 0 {
			return false, "Failed creating the namespacestore, please provide secret name"
		}
	case nbv1.NSStoreTypeAzureBlob:
		if len(nsv.NamespaceStore.Spec.AzureBlob.Secret.Name) == 0 {
			return false, "Failed creating the namespacestore, please provide secret name"
		}
	case nbv1.NSStoreTypeNSFS:
		break
	default:
		return false, "Failed to identify namespacestore type"
	}
	return true, "allowed"
}

// ValidateNamespaceStoreType validates a supported namespacestore type
func (nsv *NamespaceStoreValidator) ValidateNamespaceStoreType() (bool, string) {
	switch nsv.NamespaceStore.Spec.Type {
	case nbv1.NSStoreTypeAWSS3, nbv1.NSStoreTypeS3Compatible, nbv1.NSStoreTypeIBMCos, nbv1.NSStoreTypeAzureBlob, nbv1.NSStoreTypeNSFS:
		return true, "allowed"
	default:
		return false, "Invalid namespacestore type, please provide a valid one"
	}
}

// ValidateEmptyPvcName validates NSFS pvc name provided
func (nsv *NamespaceStoreValidator) ValidateEmptyPvcName() (bool, string) {
	if nsv.NamespaceStore.Spec.NSFS.PvcName == "" {
		return false, "Failed to create NSFS, please provide pvc name"
	}
	return true, "allowed"
}

// ValidateSubPath validates NSFS SubPath is relative path and not containing '..' character
func (nsv *NamespaceStoreValidator) ValidateSubPath() (bool, string) {
	path := nsv.NamespaceStore.Spec.NSFS.SubPath
	if len(path) > 0 && path[0] == '/' {
		return false, "Failed to create NSFS, SubPath must be a relative path"
	}
	if strings.Contains(path, "..") {
		return false, "Failed to create NSFS, SubPath must not contain '..'"
	}
	return true, "allowed"
}

// ValidateMountPath validates NSFS mount path, including '/nsfs/', is no more than 63 characters
func (nsv *NamespaceStoreValidator) ValidateMountPath() (bool, string) {
	mountPath := "/nsfs/" + nsv.NamespaceStore.Name
	if len(mountPath) > MaximumMountPathLength {
		return false, "Failed to create NSFS, MountPath must be no more than 63 characters"
	}
	return true, "allowed"
}

// ValidateTargetBucketChange validates the user is not trying to update the namespacestore target bucket
func (nsv *NamespaceStoreValidator) ValidateTargetBucketChange(oldNs nbv1.NamespaceStore) (bool, string) {
	switch nsv.NamespaceStore.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		if oldNs.Spec.AWSS3.TargetBucket != nsv.NamespaceStore.Spec.AWSS3.TargetBucket {
			return false, "Changing a NamespaceStore target bucket is unsupported"
		}
	case nbv1.NSStoreTypeS3Compatible:
		if oldNs.Spec.S3Compatible.TargetBucket != nsv.NamespaceStore.Spec.S3Compatible.TargetBucket {
			return false, "Changing a NamespaceStore target bucket is unsupported"
		}
	case nbv1.NSStoreTypeIBMCos:
		if oldNs.Spec.IBMCos.TargetBucket != nsv.NamespaceStore.Spec.IBMCos.TargetBucket {
			return false, "Changing a NamespaceStore target bucket is unsupported"
		}
	case nbv1.NSStoreTypeAzureBlob:
		if oldNs.Spec.AzureBlob.TargetBlobContainer != nsv.NamespaceStore.Spec.AzureBlob.TargetBlobContainer {
			return false, "Changing a NamespaceStore target bucket is unsupported"
		}
	default:
		return false, "Failed to identify NamespaceStore type"
	}
	return true, "allowed"
}

// ValidateNamespacestoreDeletion validates the deleted namespacestore not containing data buckets
func (nsv *NamespaceStoreValidator) ValidateNamespacestoreDeletion() (bool, string) {
	sysClient, err := system.Connect(false)
	if err != nil {
		nsv.Logger.Error("Failed to load noobaa system connection info")
		return true, "allowed"
	}
	systemInfo, err := sysClient.NBClient.ReadSystemAPI()
	if err != nil {
		nsv.Logger.Error("Failed to call ReadSystemInfo API")
		return true, "allowed"
	}

	for _, nsr := range systemInfo.NamespaceResources {
		if nsr.Name == nsv.NamespaceStore.Name {
			if nsr.Undeletable == "IN_USE" {
				return false, fmt.Sprintf("Cannot complete because nsr %q in %q state", nsr.Name, nsr.Undeletable)
			}
			return true, "allowed"
		}
	}

	return true, "allowed"
}
