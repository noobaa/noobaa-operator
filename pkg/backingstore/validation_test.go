package backingstore

import (
	"fmt"
	"reflect"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/constants"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultEndPointURI = "https://127.0.0.1:6443"
)

func TestBackingStoreS3Compatible(t *testing.T) {

	// Valid backingstore
	defaultBS := getDefaultS3CompatibleBsStore()
	err := validations.ValidateBackingStore(defaultBS)
	AssertNotError(t, err, "Valid s3-compatible backingstore validation failed")

	// Signature version is empty — allowed
	defaultBS = getDefaultS3CompatibleBsStore()
	defaultBS.Spec.S3Compatible.SignatureVersion = ""
	err = validations.ValidateBackingStore(defaultBS)
	AssertNotError(t, err, "Empty signature version should be allowed")

	// Valid v2 signature version
	defaultBS = getDefaultS3CompatibleBsStore()
	defaultBS.Spec.S3Compatible.SignatureVersion = "v2"
	err = validations.ValidateBackingStore(defaultBS)
	AssertNotError(t, err, "Valid signature version v2 should be allowed")

	// Invalid signature version
	defaultBS = getDefaultS3CompatibleBsStore()
	defaultBS.Spec.S3Compatible.SignatureVersion = "v5"
	err = validations.ValidateBackingStore(defaultBS)
	AssertError(t, err, "Invalid signature version v5 should be denied")

	// Empty secret name
	defaultBS = getDefaultS3CompatibleBsStore()
	defaultBS.Spec.S3Compatible.Secret.Name = ""
	err = validations.ValidateBackingStore(defaultBS)
	AssertError(t, err, "Empty secret name should be denied")

	// Empty target bucket
	defaultBS = getDefaultS3CompatibleBsStore()
	defaultBS.Spec.S3Compatible.TargetBucket = ""
	err = validations.ValidateBackingStore(defaultBS)
	AssertError(t, err, "Empty target bucket should be denied")

	// Nil S3Compatible spec
	defaultBS = nbv1.BackingStore{
		Spec:       nbv1.BackingStoreSpec{Type: nbv1.StoreTypeS3Compatible},
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	err = validations.ValidateBackingStore(defaultBS)
	AssertError(t, err, "Nil S3Compatible spec should be denied")
}

func TestBackingStoreIBMCos(t *testing.T) {

	// Valid backingstore
	defaultBS := getDefaultIBMCosBsStore()
	err := validations.ValidateBackingStore(defaultBS)
	AssertNotError(t, err, "Valid ibm-cos backingstore validation failed")

	// Signature version is empty — allowed
	defaultBS = getDefaultIBMCosBsStore()
	defaultBS.Spec.IBMCos.SignatureVersion = ""
	err = validations.ValidateBackingStore(defaultBS)
	AssertNotError(t, err, "Empty signature version should be allowed")

	// Valid v2 signature version
	defaultBS = getDefaultIBMCosBsStore()
	defaultBS.Spec.IBMCos.SignatureVersion = "v2"
	err = validations.ValidateBackingStore(defaultBS)
	AssertNotError(t, err, "Valid signature version v2 should be allowed")

	// Invalid signature version
	defaultBS = getDefaultIBMCosBsStore()
	defaultBS.Spec.IBMCos.SignatureVersion = "v5"
	err = validations.ValidateBackingStore(defaultBS)
	AssertError(t, err, "Invalid signature version v5 should be denied")

	// Empty secret name
	defaultBS = getDefaultIBMCosBsStore()
	defaultBS.Spec.IBMCos.Secret.Name = ""
	err = validations.ValidateBackingStore(defaultBS)
	AssertError(t, err, "Empty secret name should be denied")

	// Empty target bucket
	defaultBS = getDefaultIBMCosBsStore()
	defaultBS.Spec.IBMCos.TargetBucket = ""
	err = validations.ValidateBackingStore(defaultBS)
	AssertError(t, err, "Empty target bucket should be denied")

	// Nil IBMCos spec
	defaultBS = nbv1.BackingStore{
		Spec:       nbv1.BackingStoreSpec{Type: nbv1.StoreTypeIBMCos},
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	err = validations.ValidateBackingStore(defaultBS)
	AssertError(t, err, "Nil IBMCos spec should be denied")
}

func TestValidateBSEndpointChange(t *testing.T) {

	// S3Compatible: endpoint change should be denied
	oldBS := getDefaultS3CompatibleBsStore()
	newBS := getDefaultS3CompatibleBsStore()
	newBS.Spec.S3Compatible.Endpoint = "https://different-endpoint.example.com"
	err := validations.ValidateBSEndpointChange(newBS, oldBS)
	AssertError(t, err, "S3Compatible endpoint change should be denied")

	// S3Compatible: endpoint change allowed when pause annotation is set
	oldBS = getDefaultS3CompatibleBsStore()
	newBS = getDefaultS3CompatibleBsStore()
	newBS.Annotations = map[string]string{
		constants.PauseReconcile: "true",
	}
	newBS.Spec.S3Compatible.Endpoint = "https://different-endpoint.example.com"
	err = validations.ValidateBSEndpointChange(newBS, oldBS)
	AssertEqual(t, err, nil, "S3Compatible endpoint change with pause annotation should be allowed")

	// endpoint change not allowed when pause annotation is set any value other than true
	oldBS = getDefaultS3CompatibleBsStore()
	newBS = getDefaultS3CompatibleBsStore()
	newBS.Annotations = map[string]string{
		constants.PauseReconcile: "false",
	}
	newBS.Spec.S3Compatible.Endpoint = "https://different-endpoint.example.com"
	err = validations.ValidateBSEndpointChange(newBS, oldBS)
	AssertError(t, err, "Endpoint change with pause annotation 'false' should be not allowed")

	// S3Compatible: secret-only change, same endpoint — should be allowed
	oldBS = getDefaultS3CompatibleBsStore()
	newBS = getDefaultS3CompatibleBsStore()
	newBS.Spec.S3Compatible.Secret.Name = "new-secret"
	err = validations.ValidateBSEndpointChange(newBS, oldBS)
	AssertNotError(t, err, "S3Compatible secret-only change should be allowed")

	// S3Compatible: canonical vs non-canonical equivalent endpoint — should be allowed
	oldBS = getDefaultS3CompatibleBsStore()
	oldBS.Spec.S3Compatible.Endpoint = "minio.s3-a.svc:9000"
	newBS = getDefaultS3CompatibleBsStore()
	newBS.Spec.S3Compatible.Endpoint = "https://minio.s3-a.svc:9000"
	err = validations.ValidateBSEndpointChange(newBS, oldBS)
	AssertNotError(t, err, "S3Compatible equivalent endpoints should be allowed")

	// IBMCos: endpoint change should be denied
	oldIBM := getDefaultIBMCosBsStore()
	newIBM := getDefaultIBMCosBsStore()
	newIBM.Spec.IBMCos.Endpoint = "https://different-ibm-endpoint.example.com"
	err = validations.ValidateBSEndpointChange(newIBM, oldIBM)
	AssertError(t, err, "IBMCos endpoint change should be denied")

	// IBMCos: endpoint change allowed when pause annotation is set
	oldIBM = getDefaultIBMCosBsStore()
	newIBM = getDefaultIBMCosBsStore()
	newIBM.Annotations = map[string]string{
		constants.PauseReconcile: "true",
	}
	newIBM.Spec.IBMCos.Endpoint = "https://different-ibm-endpoint.example.com"
	err = validations.ValidateBSEndpointChange(newIBM, oldIBM)
	AssertEqual(t, err, nil, "IBMCos endpoint change with pause annotation should be allowed")

	// IBMCos: secret-only change, same endpoint — should be allowed
	oldIBM = getDefaultIBMCosBsStore()
	newIBM = getDefaultIBMCosBsStore()
	newIBM.Spec.IBMCos.Secret.Name = "new-secret"
	err = validations.ValidateBSEndpointChange(newIBM, oldIBM)
	AssertNotError(t, err, "IBMCos secret-only change should be allowed")

	// S3Compatible: nil old spec should return error
	oldBS = getDefaultS3CompatibleBsStore()
	oldBS.Spec.S3Compatible = nil
	newBS = getDefaultS3CompatibleBsStore()
	err = validations.ValidateBSEndpointChange(newBS, oldBS)
	AssertError(t, err, "nil old S3Compatible spec should return error")

	// IBMCos: nil old spec should return error
	oldIBMNil := getDefaultIBMCosBsStore()
	oldIBMNil.Spec.IBMCos = nil
	newIBMNil := getDefaultIBMCosBsStore()
	err = validations.ValidateBSEndpointChange(newIBMNil, oldIBMNil)
	AssertError(t, err, "nil old IBMCos spec should return error")

	// aws-s3 type: no endpoint field — should always be allowed
	oldAWS := getDefaultAWSS3BsStore()
	newAWS := getDefaultAWSS3BsStore()
	err = validations.ValidateBSEndpointChange(newAWS, oldAWS)
	AssertNotError(t, err, "aws-s3 type has no endpoint field, should always be allowed")
}

func TestShouldMigrateAllBucketsDefaultPool(t *testing.T) {
	internal := nb.PoolInfo{Name: "backingstores", ResourceType: "INTERNAL"}
	firstCloud := nb.PoolInfo{Name: "first-cloud", ResourceType: "CLOUD"}

	if !shouldMigrateAllBucketsDefaultPool(&nb.SystemInfo{Pools: []nb.PoolInfo{internal}}, "first-cloud") {
		t.Fatal("first non-internal pool should migrate default buckets")
	}
	if shouldMigrateAllBucketsDefaultPool(&nb.SystemInfo{Pools: []nb.PoolInfo{internal, firstCloud}}, "second-cloud") {
		t.Fatal("later cloud pool should not migrate default buckets")
	}
	if shouldMigrateAllBucketsDefaultPool(nil, "first-cloud") {
		t.Fatal("missing system info should not migrate")
	}
}

func TestValidatePvpoolScaleDown(t *testing.T) {

	// Scale up: allowed
	oldBS := getDefaultPVPoolBsStore(3)
	newBS := getDefaultPVPoolBsStore(5)
	err := validations.ValidatePvpoolScaleDown(newBS, oldBS)
	AssertNotError(t, err, "Scale up should be allowed")

	// Same count: allowed
	oldBS = getDefaultPVPoolBsStore(3)
	newBS = getDefaultPVPoolBsStore(3)
	err = validations.ValidatePvpoolScaleDown(newBS, oldBS)
	AssertNotError(t, err, "Same volume count should be allowed")

	// Scale down: denied
	oldBS = getDefaultPVPoolBsStore(5)
	newBS = getDefaultPVPoolBsStore(3)
	err = validations.ValidatePvpoolScaleDown(newBS, oldBS)
	AssertError(t, err, "Scale down should be denied")
}

func TestValidateTargetBSBucketChange(t *testing.T) {

	// S3Compatible: bucket change should be denied
	oldBS := getDefaultS3CompatibleBsStore()
	newBS := getDefaultS3CompatibleBsStore()
	newBS.Spec.S3Compatible.TargetBucket = "different-bucket"
	err := validations.ValidateTargetBSBucketChange(newBS, oldBS)
	AssertError(t, err, "S3Compatible target bucket change should be denied")

	// S3Compatible: same bucket — allowed
	oldBS = getDefaultS3CompatibleBsStore()
	newBS = getDefaultS3CompatibleBsStore()
	err = validations.ValidateTargetBSBucketChange(newBS, oldBS)
	AssertNotError(t, err, "S3Compatible same target bucket should be allowed")

	// IBMCos: bucket change should be denied
	oldIBM := getDefaultIBMCosBsStore()
	newIBM := getDefaultIBMCosBsStore()
	newIBM.Spec.IBMCos.TargetBucket = "different-bucket"
	err = validations.ValidateTargetBSBucketChange(newIBM, oldIBM)
	AssertError(t, err, "IBMCos target bucket change should be denied")

	// aws-s3: bucket change should be denied
	oldAWS := getDefaultAWSS3BsStore()
	newAWS := getDefaultAWSS3BsStore()
	newAWS.Spec.AWSS3.TargetBucket = "different-bucket"
	err = validations.ValidateTargetBSBucketChange(newAWS, oldAWS)
	AssertError(t, err, "aws-s3 target bucket change should be denied")
}

// --- helpers ---

func AssertNotError(t *testing.T, err error, format string, a ...interface{}) {
	t.Helper()
	if err != nil {
		msg := fmt.Sprintf(format, a...)
		t.Errorf("%s: %s", msg, err)
	}
}

func AssertError(t *testing.T, err error, format string, a ...interface{}) {
	t.Helper()
	if err == nil {
		msg := fmt.Sprintf(format, a...)
		t.Errorf("%s", msg)
	}
}

func AssertEqual(t *testing.T, actual, expected interface{}, format string, a ...interface{}) {
	t.Helper()
	msg := fmt.Sprintf(format, a...)
	if (actual == nil || expected == nil) && actual != expected {
		t.Errorf("%s", msg)
		return
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%s", msg)
	}
}

func getDefaultS3CompatibleBsStore() nbv1.BackingStore {
	return nbv1.BackingStore{
		Spec: nbv1.BackingStoreSpec{
			Type: nbv1.StoreTypeS3Compatible,
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
		ObjectMeta: metav1.ObjectMeta{Name: "test-bs"},
	}
}

func getDefaultIBMCosBsStore() nbv1.BackingStore {
	return nbv1.BackingStore{
		Spec: nbv1.BackingStoreSpec{
			Type: nbv1.StoreTypeIBMCos,
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
		ObjectMeta: metav1.ObjectMeta{Name: "test-bs"},
	}
}

func getDefaultAWSS3BsStore() nbv1.BackingStore {
	return nbv1.BackingStore{
		Spec: nbv1.BackingStoreSpec{
			Type: nbv1.StoreTypeAWSS3,
			AWSS3: &nbv1.AWSS3Spec{
				TargetBucket: "some-target-bucket",
				Secret: corev1.SecretReference{
					Name:      "secret-name",
					Namespace: "namespace",
				},
				Region: "us-east-1",
			},
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-bs"},
	}
}

func getDefaultPVPoolBsStore(numVolumes int) nbv1.BackingStore {
	return nbv1.BackingStore{
		Spec: nbv1.BackingStoreSpec{
			Type: nbv1.StoreTypePVPool,
			PVPool: &nbv1.PVPoolSpec{
				NumVolumes: numVolumes,
			},
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-pvpool"},
	}
}
