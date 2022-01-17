package kms

import (
	"github.com/libopenstorage/secrets/k8s"
)

// K8S is a Kubernetes driver
type K8S struct {
	name string  // NooBaa system name
	ns   string  // NooBaa system namespace
}

// NewK8S is Kubernetes secret driver constructor
func NewK8S(
	name string,
	namespace string,
	uid string,
) (Driver) {
	return &K8S{name, namespace}
}

//
// Kubernetes secret driver for root master key
//

// Path returns the k8s secret name
func (k *K8S) Path() string {
	return k.name + "-root-master-key"
}

// Name returns root key map key
func (*K8S) Name() string {
	return "cipher_key_b64"
}

func k8sContext(ns string) map[string]string {
	return map[string]string{
		k8s.SecretNamespace : ns,
	}
}

// GetContext returns context used for secret get operation
func (k *K8S) GetContext() map[string]string {
	return k8sContext(k.ns)
}

// SetContext returns context used for secret set operation
func (k *K8S) SetContext() map[string]string {
	return k8sContext(k.ns)
}

// DeleteContext returns context used for secret delete operation
func (k *K8S) DeleteContext() map[string]string {
	return k8sContext(k.ns)
}

// Config returns this driver secret config
func (k *K8S) Config(map[string]string, string, string) (map[string]interface{}, error) {
	return nil, nil
}

// Register Kubernetes secret driver with KMS layer
func init() {
	if err := RegisterDriver(k8s.Name, NewK8S); err != nil {
		panic(err.Error())
	}
}
