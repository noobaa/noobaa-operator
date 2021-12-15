package util

import (
	"fmt"
	"os"
	"syscall"

	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/ibm"

	corev1 "k8s.io/api/core/v1"
)

const (
	// IbmInstanceIDKey is the Key Protect Service's Instance ID
	IbmInstanceIDKey   = ibm.IbmInstanceIdKey
)

//
// IBM KP K8S secret driver for NooBaa root master key
//

// KMSIBM is a NooBaa root master key ibmKpK8sSecret driver
type KMSIBM struct {
	UID string  // NooBaa system UID
}

// Config returns ibmKpK8sSecret secret config
func (k *KMSIBM) Config(config map[string]string, tokenSecretName, namespace string) (map[string]interface{}, error) {
	// create a type correct copy of the configuration
	c := make(map[string]interface{})
	for k, v := range config {
		c[k] = v
	}

	// Pass the token k8s secret to the ibmKpK8sSecret implementation
	c[secretName]      = tokenSecretName
	c[secretNamespace] = namespace

	// Cloud service IBM KP Instance ID should be passed from NooBaa CR
	instanceID, instanceIDFound  := config[IbmInstanceIDKey]
	if !instanceIDFound {
		return nil, fmt.Errorf("❌ Unable to find IBM Key Protect instance ID in CR %v", IbmInstanceIDKey)
	}
	os.Setenv(IbmInstanceIDKey, instanceID)

	// Fetch API Key and Customer Root Key from k8s secret
	_, api := syscall.Getenv(ibm.IbmServiceApiKey)
	_, crk := syscall.Getenv(ibm.IbmCustomerRootKey)
	if !api || !crk {
		if err := k.keysFromSecret(tokenSecretName, namespace, c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// keysFromSecret reads API Key and Customer Root Key from k8s secret
func (k *KMSIBM) keysFromSecret(tokenSecretName, namespace string, c map[string]interface{}) error {
	secret := &corev1.Secret{}
	secret.Namespace = namespace
	secret.Name = tokenSecretName
	if !KubeCheck(secret) {
		return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
	}

	for _, key := range []string{ibm.IbmServiceApiKey, ibm.IbmCustomerRootKey} {
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
func (k *KMSIBM) Path() string {
	return "rootkeyb64-" + k.UID
}

// Name returns root key map key
func (k *KMSIBM) Name() string {
	return k.Path()
}

// custom  key context is used for plaintext keys
var customCtx = map[string]string{ secrets.CustomSecretData: "true"}

// GetContext returns context used for secret get operation
func (k *KMSIBM) GetContext() map[string]string {
	return customCtx
}

// SetContext returns context used for secret set operation
func (k *KMSIBM) SetContext() map[string]string {
	return customCtx
}

// DeleteContext returns context used for secret delete operation
func (k *KMSIBM) DeleteContext() map[string]string {
	return customCtx
}
