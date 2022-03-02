package namespacestore

import (
	"fmt"
	"reflect"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// const configuration values for the validation checks
const (
	defaultEndPointURI     = "https://127.0.0.1:6443"
	MaximumMountPathLength = 63
)

func TestNamespaceStoreNSFS(t *testing.T) {

	//Valid namespacestore
	defaultNs := getDefaultNSFSNsStore()
	err := validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Valid namespacestore validation is failed")

	//Pvcname is empty
	defaultNs = getDefaultNSFSNsStore()
	defaultNs.Spec.NSFS.PvcName = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Validation empty pvcName is failed")

	//SubPath is not relative
	defaultNs = getDefaultNSFSNsStore()
	defaultNs.Spec.NSFS.SubPath = "/"
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Validation relative subPath %s is failed", defaultNs.Spec.NSFS.SubPath)

	//SubPath contains '..'
	defaultNs = getDefaultNSFSNsStore()
	defaultNs.Spec.NSFS.SubPath = "test/../test2"
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Validation relative subPath %s is failed", defaultNs.Spec.NSFS.SubPath)

}

func TestNamespaceStoreS3Compatible(t *testing.T) {

	//Valid namespacestore
	defaultNs := getDefaultS3CompatibleNsStore()
	err := validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Valid namespacestore validation is failed")

	//Signature version is empty
	defaultNs = getDefaultS3CompatibleNsStore()
	defaultNs.Spec.S3Compatible.SignatureVersion = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Empty sugnature version validation is failed")

	//Valid v2 signature version
	defaultNs = getDefaultS3CompatibleNsStore()
	defaultNs.Spec.S3Compatible.SignatureVersion = "v2"
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Valid sugnature version %s validation is failed", defaultNs.Spec.S3Compatible.SignatureVersion)

	//Ivalid signature version
	defaultNs = getDefaultS3CompatibleNsStore()
	defaultNs.Spec.S3Compatible.SignatureVersion = "v5"
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Invalid sugnature version %s validation is failed", defaultNs.Spec.S3Compatible.SignatureVersion)

	//Empty endPoint
	defaultNs = getDefaultS3CompatibleNsStore()
	defaultNs.Spec.S3Compatible.Endpoint = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Empty endPoint validation is failed")
	AssertEqual(t, defaultEndPointURI, defaultNs.Spec.S3Compatible.Endpoint,
		"EndPoint has no the default value, %s : %s", defaultNs.Spec.S3Compatible.Endpoint, defaultEndPointURI)

	//Invalid endPoint
	defaultNs = getDefaultS3CompatibleNsStore()
	defaultNs.Spec.S3Compatible.Endpoint = "hostname:port"
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Invalid endPoint %s validation is failed", defaultNs.Spec.S3Compatible.Endpoint)

}

func TestNamespaceStoreIBMCos(t *testing.T) {

	//Valid namespacestore
	defaultNs := getDefaultIBMCosNsStore()
	err := validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Valid namespacestore validation is failed")

	//Signature version is empty
	defaultNs = getDefaultIBMCosNsStore()
	defaultNs.Spec.IBMCos.SignatureVersion = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Empty sugnature version validation is failed")

	//Valid v2 signature version
	defaultNs = getDefaultIBMCosNsStore()
	defaultNs.Spec.IBMCos.SignatureVersion = "v2"
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Valid sugnature version %s validation is failed", defaultNs.Spec.IBMCos.SignatureVersion)

	//Ivalid signature version
	defaultNs = getDefaultIBMCosNsStore()
	defaultNs.Spec.IBMCos.SignatureVersion = "v5"
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Invalid sugnature version %s validation is failed", defaultNs.Spec.IBMCos.SignatureVersion)

	//Empty endPoint
	defaultNs = getDefaultIBMCosNsStore()
	defaultNs.Spec.IBMCos.Endpoint = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Empty endPoint validation is failed")
	AssertEqual(t, defaultEndPointURI, defaultNs.Spec.IBMCos.Endpoint,
		"EndPoint has no the default value, %s : %s", defaultNs.Spec.IBMCos.Endpoint, defaultEndPointURI)

	//Invalid endPoint
	defaultNs = getDefaultIBMCosNsStore()
	defaultNs.Spec.IBMCos.Endpoint = "hostname:port"
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Invalid endPoint %s validation is failed", defaultNs.Spec.IBMCos.Endpoint)

}

func AssertNotError(t *testing.T, err error, format string, a ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, a...)
		t.Errorf("%s: %s", msg, err)
	}
}

func AssertError(t *testing.T, err error, format string, a ...interface{}) {
	if err == nil {
		msg := fmt.Sprintf(format, a...)
		t.Errorf("%s", msg)
	}
}

func AssertEqual(t *testing.T, actual, expected interface{}, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)

	if (actual == nil || expected == nil) && actual != expected {
		t.Errorf("%s", msg)
		return
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%s", msg)
	}
}

func getDefaultIBMCosNsStore() nbv1.NamespaceStore {
	return nbv1.NamespaceStore{
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeIBMCos,
			IBMCos: &nbv1.IBMCosSpec{
				SignatureVersion: nbv1.S3SignatureVersionV4,
				Endpoint:         defaultEndPointURI,
				Secret: corev1.SecretReference{
					Name:      "secret-name",
					Namespace: "namespace",
				},
				TargetBucket: "some-target-bucket",
			},
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test1"},
	}
}

func getDefaultS3CompatibleNsStore() nbv1.NamespaceStore {
	return nbv1.NamespaceStore{
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeS3Compatible,
			S3Compatible: &nbv1.S3CompatibleSpec{
				SignatureVersion: nbv1.S3SignatureVersionV4,
				Endpoint:         defaultEndPointURI,
				Secret: corev1.SecretReference{
					Name:      "secret-name",
					Namespace: "namespace",
				},
				TargetBucket: "some-target-bucket",
			},
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test1"},
	}
}

func getDefaultNSFSNsStore() nbv1.NamespaceStore {
	return nbv1.NamespaceStore{
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeNSFS,
			NSFS: &nbv1.NSFSSpec{
				PvcName: "pv-pool",
				SubPath: "subpath/",
			},
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test1"},
	}
}
