package namespacestore

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

const (
	defaultEndPointURI = "https://127.0.0.1:6443"
)

// ValidateNamespaceStore validates namespacestore configuration
func ValidateNamespaceStore(nsStore *nbv1.NamespaceStore) error {
	switch nsStore.Spec.Type {

	case nbv1.NSStoreTypeNSFS:
		return validateNsStoreNSFS(nsStore)

	case nbv1.NSStoreTypeAWSS3:
		return nil

	case nbv1.NSStoreTypeS3Compatible:
		return validateNsStoreS3Compatible(nsStore)

	case nbv1.NSStoreTypeIBMCos:
		return validateNsStoreIBMCos(nsStore)

	case nbv1.NSStoreTypeAzureBlob:
		return nil

	default:
		return util.NewPersistentError("ConfigurationError",
			fmt.Sprintf("Invalid namespace store type %q", nsStore.Spec.Type))
	}
}

// validateNamespaceStoreNSFS validates namespacestore nsfs type configuration
func validateNsStoreNSFS(nsStore *nbv1.NamespaceStore) error {
	nsfs := nsStore.Spec.NSFS

	if nsfs == nil {
		return nil
	}

	//pvcName validation
	if nsfs.PvcName == "" {
		return util.NewPersistentError("InvalidPvcName", "PvcName must not be empty")
	}

	//SubPath validation
	if nsfs.SubPath != "" {
		path := nsfs.SubPath
		if len(path) > 0 && path[0] == '/' {
			return util.NewPersistentError("InvalidSubPath",
				fmt.Sprintf("SubPath %s must be a relative path", path))
		}
		parts := strings.Split(path, "/")
		for _, item := range parts {
			if item == ".." {
				return util.NewPersistentError("InvalidSubPath",
					fmt.Sprintf("SubPath %s must not contain '..'", path))
			}
		}
	}

	//Check the mountPath
	mountPath := "/nsfs/" + nsStore.Name
	if len(mountPath) > 63 {
		return util.NewPersistentError("InvalidMountPath",
		fmt.Sprintf(" MountPath %v must be no more than 63 characters", mountPath))
	}

	return nil
}

// validateNsStoreS3Compatible validates namespacestore S3Compatible type configuration
func validateNsStoreS3Compatible(nsStore *nbv1.NamespaceStore) error {
	s3Compatible := nsStore.Spec.S3Compatible

	if s3Compatible == nil {
		return nil
	}

	err := validateSignatureVersion(s3Compatible.SignatureVersion, nsStore.Name)
	if err != nil {
		return err
	}

	err = validateEndPoint(&s3Compatible.Endpoint)
	if err != nil {
		return err
	}

	return nil
}

// validateNsStoreIBMCos validates namespacestore IBMCos type configuration
func validateNsStoreIBMCos(nsStore *nbv1.NamespaceStore) error {
	IBMCos := nsStore.Spec.IBMCos

	if IBMCos == nil {
		return nil
	}

	err := validateSignatureVersion(IBMCos.SignatureVersion, nsStore.Name)
	if err != nil {
		return err
	}

	err = validateEndPoint(&IBMCos.Endpoint)
	if err != nil {
		return err
	}

	return nil
}

//SignatureVersion validation, must be empty or v2 or v4
func validateSignatureVersion(signature nbv1.S3SignatureVersion, nsStoreName string) error {
	if signature != "" &&
		signature != nbv1.S3SignatureVersionV2 &&
		signature != nbv1.S3SignatureVersionV4 {
		return util.NewPersistentError("InvalidSignatureVersion",
			fmt.Sprintf("Invalid s3 signature version %q for namespace store %q",
				signature, nsStoreName))
	}
	return nil
}

//Endpoint validation and sets default
func validateEndPoint(endPointPointer *string) error {
	endPoint := *endPointPointer

	if endPoint == "" {
		endPoint = defaultEndPointURI
	}

	match, err := regexp.MatchString(`^\w+://`, endPoint)
	if err != nil {
		return util.NewPersistentError("InvalidEndpoint",
			fmt.Sprintf("Invalid endpoint url %q: %v", endPoint, err))
	}
	if !match {
		endPoint = "https://" + endPoint
	}
	u, err := url.Parse(endPoint)
	if err != nil {
		return util.NewPersistentError("InvalidEndpoint",
			fmt.Sprintf("Invalid endpoint url %q: %v", endPoint, err))
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	*endPointPointer = u.String()

	return nil
}
