package kms

import (
	"fmt"
	"os"
	"syscall"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

//
// IBM KP driver for NooBaa root master key
//

// IBM is a NooBaa root master key ibmKpSecretStorage driver
type IBM struct {
	UID string  // NooBaa system UID
}

// NewIBM is IBM KP driver constructor
func NewIBM(
	name string,
	namespace string,
	uid string,
) (Driver) {
	return &IBM{uid}
}

// Config returns ibmKpK8sSecret secret config
func (i *IBM) Config(config map[string]string, tokenSecretName, namespace string) (map[string]interface{}, error) {
	// create a type correct copy of the configuration
	c := make(map[string]interface{})
	for k, v := range config {
		c[k] = v
	}

	// Cloud service IBM KP Instance ID should be passed from NooBaa CR
	instanceID, instanceIDFound  := config[IbmInstanceIDKey]
	if !instanceIDFound {
		return nil, fmt.Errorf("❌ Unable to find IBM Key Protect instance ID in CR %v", IbmInstanceIDKey)
	}
	os.Setenv(IbmInstanceIDKey, instanceID)

	// Fetch API Key from k8s secret
	_, api := syscall.Getenv(IbmServiceAPIKey)
	if !api {
		if err := i.keysFromSecret(tokenSecretName, namespace, c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// keysFromSecret reads API Key from k8s secret
func (*IBM) keysFromSecret(tokenSecretName, namespace string, c map[string]interface{}) error {
	secret := &corev1.Secret{}
	secret.Namespace = namespace
	secret.Name = tokenSecretName
	if !util.KubeCheck(secret) {
		return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
	}

	for _, key := range []string{IbmServiceAPIKey} {
		val, keyOk := secret.StringData[key]
		if !keyOk {
			return fmt.Errorf(`❌ Could not find key %v in secret %q in namespace %q`, key, secret.Name, secret.Namespace)
		}
		c[key] = val
		os.Setenv(key, val) // cache the value in environment
	}

	return nil
}

// Path returns secret id
func (i *IBM) Path() string {
	return "rootkeyb64-" + i.UID
}

// Name returns root key map key
func (i *IBM) Name() string {
	return i.Path()
}

// GetContext returns context used for secret get operation
func (*IBM) GetContext() map[string]string {
	return nil
}

// SetContext returns context used for secret set operation
func (*IBM) SetContext() map[string]string {
	return nil
}

// DeleteContext returns context used for secret delete operation
func (*IBM) DeleteContext() map[string]string {
	return nil
}

// Register IBM KP driver with KMS layer
func init() {
	if err := RegisterDriver(IbmKpSecretStorageName, NewIBM); err != nil {
		panic(err.Error())
	}
}
