package kms

import (
	"fmt"
	"strings"

	"github.com/libopenstorage/secrets"
)

////////////////////////////////////////////////////////////////////////////
/////////// KMS provides uniform access to several backend types ///////////
////////////////////////////////////////////////////////////////////////////
const (
	Provider             = "KMS_PROVIDER" // backend type configuration key
)

// SingleSecret represents a single secret
// several backend types are implemented, more types could be added
type SingleSecret interface {
	// Get secret value from KMS
	Get() (string, error)

	// Set secret value in KMS
	Set(value string) error

	// Delete secret value from KMS
	Delete() error
}

// Driver is a backend type specific driver interface for libopenstorage/secrets framework
type Driver interface {
	Path()          string
	Name()          string
	Config(connectionDetails map[string]string, tokenSecretName, namespace string)  (map[string]interface{}, error)
	GetContext()    map[string]string
	SetContext()    map[string]string
	DeleteContext() map[string]string
}

// DriverCtor is a Driver constructor function type
type DriverCtor func(
	name string,
	namespace string,
	uid string,
) (Driver)

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
) (Driver) {
	if dCtor, exists := kmsDrivers[dType]; exists {
		return dCtor(name, namespace, uid)
	}
	return nil
}

// KMS implements SingleSecret interface using backend implementation of
// secrets.Secrets interface and using backend type specific driver
type KMS struct {
	secrets.Secrets   // secrets interface
	Type   string     // backend system type, k8s, vault & ibm are supported
	driver Driver     // backend type specific driver
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

	return &KMS{s, t, driver}, nil
}

// Get secret value from KMS
func (k *KMS) Get() (string, error) {
	s, err := k.GetSecret(k.driver.Path(), k.driver.GetContext())
	if err != nil {
		// handle k8s get from non-existent secret
		if strings.Contains(err.Error(), "not found") {
			return "", secrets.ErrInvalidSecretId
		}
		return "", err
	}

	return s[k.driver.Name()].(string), nil
}

// Set secret value in KMS
func (k *KMS) Set(v string) error {
	data := map[string]interface{} {
		k.driver.Name(): v,
	}

	return k.PutSecret(k.driver.Path(), data, k.driver.SetContext())
}

// Delete secret value from KMS
func (k *KMS) Delete() error {
	return k.DeleteSecret(k.driver.Path(), k.driver.DeleteContext())
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
