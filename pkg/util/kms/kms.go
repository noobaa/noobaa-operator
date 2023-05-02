package kms

import (
	"fmt"

	"github.com/libopenstorage/secrets"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// //////////////////////////////////////////////////////////////////////////
// ///////// KMS provides uniform access to several backend types ///////////
// //////////////////////////////////////////////////////////////////////////
const (
	Provider = "KMS_PROVIDER" // backend type configuration key
)

// SecretStorage represents a key secret storage
// several backend types are implemented, more types could be added
type SecretStorage interface {
	// Get the secret string/map from KMS
	Get() error

	// Set active master root key secret value in KMS
	Set(value string) error

	// Delete the secret string/map from KMS
	Delete() error

	// Reconcile secret data with system reconciler
	// expose secret data to NooBaa pods
	Reconcile(r SecretReconciler) error
}

// SecretReconciler is interface exposed and implemented
// by the System reconciler
// The Version interface implementation is reponsible to call the appropriate
// method: either string or map
type SecretReconciler interface {
	ReconcileSecretString(val string) error
	ReconcileSecretMap(val map[string]string) error
}

// Driver is a backend type specific driver interface for libopenstorage/secrets framework
type Driver interface {
	Path() string
	Name() string
	Config(connectionDetails map[string]string, tokenSecretName, namespace string) (map[string]interface{}, error)
	GetContext() map[string]string
	SetContext() map[string]string
	DeleteContext() map[string]string
	Version(k *KMS) Version
}

// DriverCtor is a Driver constructor function type
type DriverCtor func(
	name string,
	namespace string,
	uid string,
) Driver

// kmsDrivers is a map of all registered drivers
var kmsDrivers = make(map[string]DriverCtor)

// RegisterDriver adds a new KMS driver
func RegisterDriver(name string, ctor DriverCtor) error {
	if _, exists := kmsDrivers[name]; exists {
		return fmt.Errorf("KMS driver %v is already registered", name)
	}
	kmsDrivers[name] = ctor
	return nil
}

// NewDriver returns a new instance of KMS driver identified by
// the supplied driver type.
func NewDriver(
	dType string,
	name string,
	namespace string,
	uid string,
) Driver {
	if dCtor, exists := kmsDrivers[dType]; exists {
		return dCtor(name, namespace, uid)
	}
	return nil
}

// KMS implements SingleSecret interface using backend implementation of
// secrets.Secrets interface and using backend type specific driver
type KMS struct {
	secrets.Secrets        // secrets interface
	Version                // KMS backend version, single secret or rotating/map
	Type            string // backend system type, k8s, vault & ibm are supported
	driver          Driver // backend type specific driver
}

// NewKMS creates a new secret KMS client
// or returns error otherwise
func NewKMS(connectionDetails map[string]string, tokenSecretName, name, namespace, uid string) (*KMS, error) {
	t := kmsType(connectionDetails)

	// Create KMS driver
	driver := NewDriver(t, name, namespace, uid)
	if driver == nil {
		return nil, fmt.Errorf("Unsupported KMS driver type %v", t)
	}

	// Generate backend configuration using backend driver instance
	c, err := driver.Config(connectionDetails, tokenSecretName, namespace)
	if err != nil {
		return nil, err
	}

	// Construct new backend
	s, err := secrets.New(t, c)
	if err != nil {
		return nil, err
	}

	// Create the instance with the appropriate version
	k := &KMS{s, nil, t, driver}
	k.Version = driver.Version(k)

	// Upgrade backend storage
	if err = k.Upgrade(); err != nil {
		return nil, err
	}

	return k, nil
}

// kmsType returns the secret backend type
func kmsType(connectionDetails map[string]string) string {
	if len(connectionDetails) > 0 {
		if provider, ok := connectionDetails[Provider]; ok {
			return provider
		}
	}

	// by default use Kubernes secrets
	return secrets.TypeK8s
}

// StatusValid returns true is the status is valid, false otherwise
func StatusValid(st corev1.ConditionStatus) bool {
	return st == nbv1.ConditionKMSSync || st == nbv1.ConditionKMSInit || st == nbv1.ConditionKMSKeyRotate
}
