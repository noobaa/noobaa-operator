package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/k8s"
	"github.com/libopenstorage/secrets/vault"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
)

///////////////////////////////////
/////////// KMS ///////////
///////////////////////////////////
const (
	KmsProvider             = "KMS_PROVIDER"
)

// KMSSecret represents a single secret
// k8s and vault secrets are implemented, more types could be added
type KMSSecret interface {
	// Get secret value from KMS
	Get() (string, error)

	// Set secret value in KMS
	Set(value string) error

	// Delete secret value from KMS
	Delete() error
}

// KMSDriver is a type sepcific driver for libopenstorage/secrets framework
type KMSDriver interface {
	Path()          string
	Name()          string
	Config(connectionDetails map[string]string, tokenSecretName, namespace string)  (map[string]interface{}, error)
	GetContext()    map[string]string
	SetContext()    map[string]string
	DeleteContext() map[string]string
}

// KMS implements KMSSecret interface
type KMS struct {
	secrets.Secrets   // secrets interface
	Type   string     // system type, k8s & vault are supported
	driver KMSDriver  // type specific driver
}

// NewKMS creates a new secret KMS client
// or returns error otherwise
func NewKMS(connectionDetails map[string]string, tokenSecretName, name, namespace, uid string) (*KMS, error) {
	t := kmsType(connectionDetails)

	var driver KMSDriver
	switch t {
	case k8s.Name:
		driver = &KMSK8S{name, namespace}
	case vault.Name:
		driver = &KMSVault{uid}
	default:
		return nil, fmt.Errorf("Unsupported KMS type %v", t)
	}

	c, err := driver.Config(connectionDetails, tokenSecretName, namespace)
	if err != nil {
		return nil, err
	}

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
		if provider, ok := connectionDetails[KmsProvider]; ok {
			return provider
		}
	}

	// by default use Kubernes secrets
	return secrets.TypeK8s
}

//
// Test shared utilities
//

// NooBaaCondStatus waits for requested NooBaa CR KSM condition status
// returns false if timeout
func NooBaaCondStatus(noobaa* nbv1.NooBaa, s corev1.ConditionStatus) bool {
	return NooBaaCondition(noobaa, nbv1.ConditionTypeKMSStatus, s)
}

// NooBaaCondition waits for requested NooBaa CR KSM condition type & status
// returns false if timeout
func NooBaaCondition(noobaa* nbv1.NooBaa, t conditionsv1.ConditionType, s corev1.ConditionStatus) bool {
	found := false

	timeout := 120 // seconds
	for i := 0; i < timeout; i++ {
		_, _, err := KubeGet(noobaa)
		Panic(err)

		if noobaaStatus(noobaa, t, s) {
			found = true
			break
		}
		time.Sleep(time.Second)
	}

	return found
}
