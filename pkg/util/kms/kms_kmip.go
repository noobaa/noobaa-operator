package kms

import (
	"fmt"
	"strconv"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	corev1 "k8s.io/api/core/v1"
)

// KMIP client config options
const (
	KMIPEndpoint      = "KMIP_ENDPOINT"
	KMIPSecret        = "KMIP_CERTS_SECRET"
	KMIPUniqueID      = "UniqueIdentifier"
	KMIPTLSServerName = "TLS_SERVER_NAME"
	KMIPReadTimeOut   = "READ_TIMEOUT"
	KMIPWriteTimeOut  = "WRITE_TIMEOUT"
	KMPSecret         = "secret"
	KMIPCACERT        = "CA_CERT"
	KMIPCLIENTCERT    = "CLIENT_CERT"
	KMIPCLIENTKEY     = "CLIENT_KEY"
)

// KMIP is a kmip driver
type KMIP struct {
}

// NewKMIP is KMIP driver constructor
func NewKMIP(
	name string,
	namespace string,
	uid string,
) Driver {
	return &KMIP{}
}

//
// KMIP specific driver for root master key
//

// Config returns this driver secret config
func (*KMIP) Config(config map[string]string, tokenSecretName, namespace string) (map[string]interface{}, error) {
	// create a type correct copy of the configuration
	c := make(map[string]interface{})

	// Fetch the KMIP secret
	secret := &corev1.Secret{}
	secret.Namespace = namespace
	secret.Name = tokenSecretName
	if !util.KubeCheck(secret) {
		return nil, fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
	}

	c[KMPSecret] = secret
	c[KMIPCACERT] = secret.StringData[KMIPCACERT]
	c[KMIPCLIENTCERT] = secret.StringData[KMIPCLIENTCERT]
	c[KMIPCLIENTKEY] = secret.StringData[KMIPCLIENTKEY]

	if endpoint, exists := config[KMIPEndpoint]; exists {
		c[KMIPEndpoint] = endpoint
	} else {
		return nil, fmt.Errorf("❌ Could not find in config KMP Endpoint %v", KMIPEndpoint)
	}
	if tlsServerName, exists := config[KMIPTLSServerName]; exists {
		c[KMIPTLSServerName] = tlsServerName
	}
	if readTimeoutString, exists := config[KMIPReadTimeOut]; exists {
		if readTimeoutInt, err := strconv.Atoi(readTimeoutString); err == nil {
			c[KMIPReadTimeOut] = readTimeoutInt
		}
	}
	if writeTimeoutString, exists := config[KMIPWriteTimeOut]; exists {
		if writeTimeoutInt, err := strconv.Atoi(writeTimeoutString); err == nil {
			c[KMIPWriteTimeOut] = writeTimeoutInt
		}
	}

	return c, nil
}

// Name returns root key map key
func (k *KMIP) Name() string {
	return "KMIPSecret"
}

// Path return  kv secret id
func (k *KMIP) Path() string {
	return k.Name()
}

// GetContext returns context used for secret get operation
func (k *KMIP) GetContext() map[string]string {
	return nil
}

// SetContext returns context used for secret set operation
func (k *KMIP) SetContext() map[string]string {
	return nil
}

// DeleteContext returns context used for secret delete operation
func (k *KMIP) DeleteContext() map[string]string {
	return nil
}

// Register KMIP driver with KMS layer
func init() {
	if err := RegisterDriver(KMIPSecretStorageName, NewKMIP); err != nil {
		panic(err.Error())
	}
}
