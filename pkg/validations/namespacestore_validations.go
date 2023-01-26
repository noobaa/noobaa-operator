package validations

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

// const configuration values for the validation checks
const (
	defaultEndPointURI     = "https://127.0.0.1:6443"
	MaximumMountPathLength = 63
)

// ValidateNamespaceStore validates namespacestore configuration
func ValidateNamespaceStore(nsStore *nbv1.NamespaceStore) error {
	if err := ValidateNSEmptySecretName(*nsStore); err != nil {
		return err
	}
	if err := ValidateNSEmptyTargetBucket(*nsStore); err != nil {
		return err
	}
	switch nsStore.Spec.Type {

	case nbv1.NSStoreTypeNSFS:
		return ValidateNsStoreNSFS(nsStore)

	case nbv1.NSStoreTypeAWSS3:
		return nil

	case nbv1.NSStoreTypeS3Compatible:
		return ValidateNsStoreS3Compatible(nsStore)

	case nbv1.NSStoreTypeIBMCos:
		return ValidateNsStoreIBMCos(nsStore)

	case nbv1.NSStoreTypeAzureBlob:
		return nil

	case nbv1.NSStoreTypeGoogleCloudStorage:
		return nil

	default:
		return util.ValidationError{
			Msg: "Invalid Namespacestore type, please provide a valid Namespacestore type",
		}
	}
}

// ValidateNsStoreNSFS validates namespacestore nsfs type configuration
func ValidateNsStoreNSFS(nsStore *nbv1.NamespaceStore) error {
	nsfs := nsStore.Spec.NSFS

	if nsfs == nil {
		return nil
	}

	//pvcName validation
	if nsfs.PvcName == "" {
		return util.ValidationError{
			Msg: "PvcName must not be empty",
		}
	}

	//SubPath validation
	if nsfs.SubPath != "" {
		path := nsfs.SubPath
		if len(path) > 0 && path[0] == '/' {
			return util.ValidationError{
				Msg: fmt.Sprintf("SubPath %s must be a relative path", path),
			}
		}
		parts := strings.Split(path, "/")
		for _, item := range parts {
			if item == ".." {
				return util.ValidationError{
					Msg: fmt.Sprintf("SubPath %s must not contain '..'", path),
				}
			}
		}
	}

	//Check the mountPath
	mountPath := "/nsfs/" + nsStore.Name
	if len(mountPath) > MaximumMountPathLength {
		return util.ValidationError{
			Msg: fmt.Sprintf("MountPath %v must be no more than 63 characters", mountPath),
		}
	}

	return nil
}

// ValidateNsStoreS3Compatible validates namespacestore S3Compatible type configuration
func ValidateNsStoreS3Compatible(nsStore *nbv1.NamespaceStore) error {
	s3Compatible := nsStore.Spec.S3Compatible

	if s3Compatible == nil {
		return nil
	}

	err := ValidateSignatureVersion(s3Compatible.SignatureVersion, nsStore.Name)
	if err != nil {
		return err
	}

	err = ValidateEndPoint(&s3Compatible.Endpoint)
	if err != nil {
		return err
	}

	return nil
}

// ValidateNsStoreIBMCos validates namespacestore IBMCos type configuration
func ValidateNsStoreIBMCos(nsStore *nbv1.NamespaceStore) error {
	IBMCos := nsStore.Spec.IBMCos

	if IBMCos == nil {
		return nil
	}

	err := ValidateSignatureVersion(IBMCos.SignatureVersion, nsStore.Name)
	if err != nil {
		return err
	}

	err = ValidateEndPoint(&IBMCos.Endpoint)
	if err != nil {
		return err
	}

	return nil
}

//ValidateSignatureVersion validation, must be empty or v2 or v4
func ValidateSignatureVersion(signature nbv1.S3SignatureVersion, nsStoreName string) error {
	if signature != "" &&
		signature != nbv1.S3SignatureVersionV2 &&
		signature != nbv1.S3SignatureVersionV4 {
		return util.ValidationError{
			Msg: fmt.Sprintf("Invalid s3 signature version %q for namespace store %q", signature, nsStoreName),
		}
	}
	return nil
}

//ValidateEndPoint Endpoint validation and sets default
func ValidateEndPoint(endPointPointer *string) error {
	endPoint := *endPointPointer

	if endPoint == "" {
		endPoint = defaultEndPointURI
	}

	match, err := regexp.MatchString(`^\w+://`, endPoint)
	if err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("Invalid endpoint url %q: %v", endPoint, err),
		}
	}
	if !match {
		endPoint = "https://" + endPoint
	}
	u, err := url.Parse(endPoint)
	if err != nil {
		return util.ValidationError{
			Msg: fmt.Sprintf("Invalid endpoint url %q: %v", endPoint, err),
		}
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	*endPointPointer = u.String()

	return nil
}

// ValidateNSEmptySecretName validates a secret name is provided for cloud namespacestore
func ValidateNSEmptySecretName(ns nbv1.NamespaceStore) error {
	switch ns.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		if len(ns.Spec.AWSS3.Secret.Name) == 0 {
			if err := ValidateNSEmptyAWSARN(ns); err != nil {
				return err
			}
		}
	case nbv1.NSStoreTypeS3Compatible:
		if len(ns.Spec.S3Compatible.Secret.Name) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide secret name",
			}
		}
	case nbv1.NSStoreTypeIBMCos:
		if len(ns.Spec.IBMCos.Secret.Name) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide secret name",
			}
		}
	case nbv1.NSStoreTypeAzureBlob:
		if len(ns.Spec.AzureBlob.Secret.Name) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide secret name",
			}
		}
	case nbv1.NSStoreTypeGoogleCloudStorage:
		if len(ns.Spec.GoogleCloudStorage.Secret.Name) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide secret name",
			}
		}
	case nbv1.NSStoreTypeNSFS:
		break
	default:
		return util.ValidationError{
			Msg: "Invalid Namespacestore type, please provide a valid Namespacestore type",
		}
	}
	return nil
}

// ValidateNSEmptyTargetBucket validates a target bucket name is provided for cloud namespacestore
func ValidateNSEmptyTargetBucket(ns nbv1.NamespaceStore) error {
	switch ns.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		if len(ns.Spec.AWSS3.TargetBucket) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide target bucket",
			}
		}
	case nbv1.NSStoreTypeS3Compatible:
		if len(ns.Spec.S3Compatible.TargetBucket) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide target bucket",
			}
		}
	case nbv1.NSStoreTypeIBMCos:
		if len(ns.Spec.IBMCos.TargetBucket) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide target bucket",
			}
		}
	case nbv1.NSStoreTypeAzureBlob:
		if len(ns.Spec.AzureBlob.TargetBlobContainer) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide target bucket",
			}
		}
	case nbv1.NSStoreTypeGoogleCloudStorage:
		if len(ns.Spec.GoogleCloudStorage.TargetBucket) == 0 {
			return util.ValidationError{
				Msg: "Failed creating the namespacestore, please provide target bucket",
			}
		}
	case nbv1.NSStoreTypeNSFS:
		break
	default:
		return util.ValidationError{
			Msg: "Invalid Namespacestore type, please provide a valid Namespacestore type",
		}
	}
	return nil
}

// ValidateTargetNSBucketChange validates the user is not trying to update the namespacestore target bucket
func ValidateTargetNSBucketChange(ns nbv1.NamespaceStore, oldNs nbv1.NamespaceStore) error {
	switch ns.Spec.Type {
	case nbv1.NSStoreTypeAWSS3:
		if oldNs.Spec.AWSS3.TargetBucket != ns.Spec.AWSS3.TargetBucket {
			return util.ValidationError{
				Msg: "Changing a NamespaceStore target bucket is unsupported",
			}
		}
	case nbv1.NSStoreTypeS3Compatible:
		if oldNs.Spec.S3Compatible.TargetBucket != ns.Spec.S3Compatible.TargetBucket {
			return util.ValidationError{
				Msg: "Changing a NamespaceStore target bucket is unsupported",
			}
		}
	case nbv1.NSStoreTypeIBMCos:
		if oldNs.Spec.IBMCos.TargetBucket != ns.Spec.IBMCos.TargetBucket {
			return util.ValidationError{
				Msg: "Changing a NamespaceStore target bucket is unsupported",
			}
		}
	case nbv1.NSStoreTypeAzureBlob:
		if oldNs.Spec.AzureBlob.TargetBlobContainer != ns.Spec.AzureBlob.TargetBlobContainer {
			return util.ValidationError{
				Msg: "Changing a NamespaceStore target bucket is unsupported",
			}
		}
	case nbv1.NSStoreTypeGoogleCloudStorage:
		if oldNs.Spec.GoogleCloudStorage.TargetBucket != ns.Spec.GoogleCloudStorage.TargetBucket {
			return util.ValidationError{
				Msg: "Changing a NamespaceStore target bucket is unsupported",
			}
		}
	default:
		return util.ValidationError{
			Msg: "Failed to identify NamespaceStore type",
		}
	}
	return nil
}

// ValidateNSEmptyAWSARN validates if ARN is present in the NamespaceStore Spec
func ValidateNSEmptyAWSARN(ns nbv1.NamespaceStore) error {
	if ns.Spec.AWSS3 != nil { 
		if ns.Spec.AWSS3.AWSSTSRoleARN == nil {
				return util.ValidationError{
					Msg: "Failed creating the NamespaceStore, please provide a valid ARN or secret name",
			}
		}
	}
	return nil
}

// ValidateNamespacestoreDeletion validates the deleted namespacestore not containing data buckets
func ValidateNamespacestoreDeletion(ns nbv1.NamespaceStore, systemInfo nb.SystemInfo) error {
	for _, nsr := range systemInfo.NamespaceResources {
		if nsr.Name == ns.Name {
			if nsr.Undeletable == "IN_USE" {
				return util.ValidationError{
					Msg: fmt.Sprintf("cannot complete because nsr %q in %q state", nsr.Name, nsr.Undeletable),
				}
			}
			return nil
		}
	}

	return nil
}
