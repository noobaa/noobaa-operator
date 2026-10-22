package namespacestore

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/constants"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

func TestNamespaceStoreAzureBlob(t *testing.T) {
	// Valid namespacestore with secret (Azure blob requires secret; no STS path for namespace store)
	defaultNs := getDefaultAzureBlobNsStore()
	err := validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Valid Azure blob namespacestore validation failed")

	// AzureBlob spec is nil
	defaultNs = nbv1.NamespaceStore{
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeAzureBlob,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-azure"},
	}
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "AzureBlob spec nil should be denied")

	// Empty secret name (namespace store Azure requires secret)
	defaultNs = getDefaultAzureBlobNsStore()
	defaultNs.Spec.AzureBlob.Secret.Name = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Empty secret name for Azure blob namespacestore should be denied")

	// Empty target blob container
	defaultNs = getDefaultAzureBlobNsStore()
	defaultNs.Spec.AzureBlob.TargetBlobContainer = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Empty target blob container should be denied")
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

func TestNamespaceStoreGoogleCloudStorage(t *testing.T) {

	//Valid namespacestore
	defaultNs := getDefaultGoogleCloudStorageNsStore()
	err := validations.ValidateNamespaceStore(&defaultNs)
	AssertNotError(t, err, "Valid Google Cloud Storage namespacestore validation failed")

	// GoogleCloudStorage spec is nil
	defaultNs = nbv1.NamespaceStore{
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeGoogleCloudStorage,
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-gcp"},
	}
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "GoogleCloudStorage spec nil should be denied")

	// Empty secret name
	defaultNs = getDefaultGoogleCloudStorageNsStore()
	defaultNs.Spec.GoogleCloudStorage.Secret.Name = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Empty secret name for Google Cloud Storage namespacestore should be denied")

	// Empty target bucket
	defaultNs = getDefaultGoogleCloudStorageNsStore()
	defaultNs.Spec.GoogleCloudStorage.TargetBucket = ""
	err = validations.ValidateNamespaceStore(&defaultNs)
	AssertError(t, err, "Empty target bucket should be denied")
}

func TestValidateNSEndpointChange(t *testing.T) {
	// S3Compatible: endpoint change should be denied
	oldNS := getDefaultS3CompatibleNsStore()
	newNS := getDefaultS3CompatibleNsStore()
	newNS.Spec.S3Compatible.Endpoint = "https://different-endpoint.example.com"
	err := validations.ValidateNSEndpointChange(newNS, oldNS)
	AssertError(t, err, "S3Compatible endpoint change should be denied")

	// endpoint change not allowed when pause annotation is set any value other than true
	oldNS = getDefaultS3CompatibleNsStore()
	newNS = getDefaultS3CompatibleNsStore()
	newNS.Annotations = map[string]string{
		constants.PauseReconcile: "false",
	}
	newNS.Spec.S3Compatible.Endpoint = "https://different-endpoint.example.com"
	err = validations.ValidateNSEndpointChange(newNS, oldNS)
	AssertError(t, err, "Endpoint change with pause annotation 'false' should be not allowed")

	// S3Compatible: endpoint change should be allowed on pause annotation
	oldNS = getDefaultS3CompatibleNsStore()
	newNS = getDefaultS3CompatibleNsStore()
	newNS.Annotations = map[string]string{
		constants.PauseReconcile: "true",
	}
	newNS.Spec.S3Compatible.Endpoint = "https://different-endpoint.example.com"
	err = validations.ValidateNSEndpointChange(newNS, oldNS)
	AssertNotError(t, err, "S3Compatible endpoint change with pause annotation should be allowed")

	// S3Compatible: only secret changed, endpoint same — should be allowed
	oldNS = getDefaultS3CompatibleNsStore()
	newNS = getDefaultS3CompatibleNsStore()
	newNS.Spec.S3Compatible.Secret.Name = "new-secret"
	err = validations.ValidateNSEndpointChange(newNS, oldNS)
	AssertNotError(t, err, "S3Compatible secret-only change should be allowed")

	// S3Compatible: canonical vs non-canonical endpoint with secret rotation — should be allowed
	oldNS = getDefaultS3CompatibleNsStore()
	oldNS.Spec.S3Compatible.Endpoint = "minio.s3-a.svc:9000"
	newNS = getDefaultS3CompatibleNsStore()
	newNS.Spec.S3Compatible.Endpoint = "https://minio.s3-a.svc:9000"
	newNS.Spec.S3Compatible.Secret.Name = "new-secret"
	err = validations.ValidateNSEndpointChange(newNS, oldNS)
	AssertNotError(t, err, "S3Compatible secret rotation with equivalent endpoints should be allowed")

	// S3Compatible: nil old spec should return error
	oldNS = getDefaultS3CompatibleNsStore()
	oldNS.Spec.S3Compatible = nil
	newNS = getDefaultS3CompatibleNsStore()
	err = validations.ValidateNSEndpointChange(newNS, oldNS)
	AssertError(t, err, "nil old S3Compatible spec should return error")

	// IBMCos: nil old spec should return error
	oldIBMNil := getDefaultIBMCosNsStore()
	oldIBMNil.Spec.IBMCos = nil
	newIBMNil := getDefaultIBMCosNsStore()
	err = validations.ValidateNSEndpointChange(newIBMNil, oldIBMNil)
	AssertError(t, err, "nil old IBMCos spec should return error")

	// IBMCos: endpoint change should be denied
	oldIBM := getDefaultIBMCosNsStore()
	newIBM := getDefaultIBMCosNsStore()
	newIBM.Spec.IBMCos.Endpoint = "https://different-ibm-endpoint.example.com"
	err = validations.ValidateNSEndpointChange(newIBM, oldIBM)
	AssertError(t, err, "IBMCos endpoint change should be denied")

	// IBMCos: endpoint change should be allowed on pause annotation
	oldIBM = getDefaultIBMCosNsStore()
	newIBM = getDefaultIBMCosNsStore()
	newIBM.Annotations = map[string]string{
		constants.PauseReconcile: "true",
	}
	newIBM.Spec.IBMCos.Endpoint = "https://different-ibm-endpoint.example.com"
	err = validations.ValidateNSEndpointChange(newIBM, oldIBM)
	AssertNotError(t, err, "IBMCos endpoint change with pause annotation should be allowed")

	// IBMCos: only secret changed, endpoint same — should be allowed
	oldIBM = getDefaultIBMCosNsStore()
	newIBM = getDefaultIBMCosNsStore()
	newIBM.Spec.IBMCos.Secret.Name = "new-secret"
	err = validations.ValidateNSEndpointChange(newIBM, oldIBM)
	AssertNotError(t, err, "IBMCos secret-only change should be allowed")
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

func useKubeClient(t *testing.T, objs ...client.Object) {
	t.Helper()
	cl := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objs...).Build()
	prev := util.SetKubeClientForTest(cl)
	t.Cleanup(func() { util.SetKubeClientForTest(prev) })
}

func nsfsStore(name, pvcName string, phase nbv1.NamespaceStorePhase, created time.Time) *nbv1.NamespaceStore {
	return &nbv1.NamespaceStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "openshift-storage",
			CreationTimestamp: metav1.NewTime(created),
		},
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeNSFS,
			NSFS: &nbv1.NSFSSpec{PvcName: pvcName},
		},
		Status: nbv1.NamespaceStoreStatus{Phase: phase},
	}
}

func checkNsfsPvcReconciler(store *nbv1.NamespaceStore) *Reconciler {
	return &Reconciler{
		NamespaceStore: store,
		Logger:         logrus.NewEntry(logrus.New()),
	}
}

func TestCheckNsfsPvcRejectsReadyStoreWhenPvcIsGone(t *testing.T) {
	store := nsfsStore("nsfs-ns-store-test", "nsfs-test-pvc", nbv1.NamespaceStorePhaseReady, time.Now().Add(-time.Hour))
	useKubeClient(t, store)

	err := checkNsfsPvcReconciler(store).checkNsfsPvc()
	perr, ok := err.(*util.PersistentError)
	if !ok {
		t.Fatalf("expected MissingPvc persistent error, got %v", err)
	}
	if perr.Reason != "MissingPvc" {
		t.Fatalf("reason = %q", perr.Reason)
	}
}

func TestCheckNsfsPvcWaitsForYoungStore(t *testing.T) {
	// Age-only grace matches MissingSecret, including after SetPhase(Verifying)
	// has already cleared a Ready phase in memory.
	store := nsfsStore("nsfs-new", "nsfs-test-pvc", nbv1.NamespaceStorePhaseReady, time.Now())
	useKubeClient(t, store)

	err := checkNsfsPvcReconciler(store).checkNsfsPvc()
	if err == nil || util.IsPersistentError(err) {
		t.Fatalf("expected a temporary error, got %v", err)
	}
}

func TestCheckNsfsPvcAllowsExistingPvc(t *testing.T) {
	store := nsfsStore("nsfs-ok", "nsfs-test-pvc", nbv1.NamespaceStorePhaseReady, time.Now().Add(-time.Hour))
	pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "nsfs-test-pvc", Namespace: store.Namespace}}
	useKubeClient(t, store, pvc)

	if err := checkNsfsPvcReconciler(store).checkNsfsPvc(); err != nil {
		t.Fatal(err)
	}
}

func TestMapPvcToNamespaceStores(t *testing.T) {
	match := nsfsStore("nsfs-match", "shared-pvc", nbv1.NamespaceStorePhaseReady, time.Now())
	other := nsfsStore("nsfs-other", "other-pvc", nbv1.NamespaceStorePhaseReady, time.Now())
	useKubeClient(t, match, other)

	reqs := MapPvcToNamespaceStores(types.NamespacedName{Name: "shared-pvc", Namespace: match.Namespace})
	if len(reqs) != 1 || reqs[0].Name != match.Name {
		t.Fatalf("requests = %+v", reqs)
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

func getDefaultAzureBlobNsStore() nbv1.NamespaceStore {
	return nbv1.NamespaceStore{
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeAzureBlob,
			AzureBlob: &nbv1.AzureBlobSpec{
				TargetBlobContainer: "azure-container",
				Secret: corev1.SecretReference{
					Name:      "azure-secret",
					Namespace: "namespace",
				},
			},
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-azure"},
	}
}

func getDefaultGoogleCloudStorageNsStore() nbv1.NamespaceStore {
	return nbv1.NamespaceStore{
		Spec: nbv1.NamespaceStoreSpec{
			Type: nbv1.NSStoreTypeGoogleCloudStorage,
			GoogleCloudStorage: &nbv1.GoogleCloudStorageSpec{
				TargetBucket: "gcp-target-bucket",
				Secret: corev1.SecretReference{
					Name:      "gcp-secret",
					Namespace: "namespace",
				},
			},
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-gcp"},
	}
}
