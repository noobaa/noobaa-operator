package kms

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/azure"
	corev1 "k8s.io/api/core/v1"
)

// Azure authentication config options
const (
	AzureVaultURL       = "AZURE_VAULT_URL"
	AzureVaultClientID  = "AZURE_CLIENT_ID"
	AzureVaultTenantID  = "AZURE_TENANT_ID"
	ServiceName         = "KMS_SERVICE_NAME"
	AzureClientCertPath = "AZURE_CERT_SECRET_NAME"
)

// AzureVault is a azure kms driver
type AzureVault struct {
	UID string // NooBaa system UID
}

// NewAzureVault is azure driver constructor
func NewAzureVault(
	name string,
	namespace string,
	uid string,
) Driver {
	return &AzureVault{uid}
}

//
// Azure vault specific driver for root master key
//

// Config returns this driver secret config
func (*AzureVault) Config(config map[string]string, tokenSecretName, namespace string) (map[string]interface{}, error) {
	// create a type correct copy of the configuration
	c := make(map[string]interface{})
	for k, v := range config {
		c[k] = v
	}

	// replace in azure vault Config secret names with local temp files paths
	err := createCertTempFile(c, namespace)
	if err != nil {
		return nil, fmt.Errorf(`❌ Could not init azure vault config %q in namespace %q`, config, namespace)
	}

	return c, nil
}

// Name returns root key map key
func (v *AzureVault) Name() string {
	return "rootkeyb64-" + v.UID
}

// Path return vault's kv secret id
func (v *AzureVault) Path() string {
	return RootSecretPath + "/rootkeyb64-" + v.UID
}

// GetContext returns context used for secret get operation
func (v *AzureVault) GetContext() map[string]string {
	return nil
}

// SetContext returns context used for secret set operation
func (v *AzureVault) SetContext() map[string]string {
	return nil
}

// DeleteContext returns context used for secret delete operation
func (v *AzureVault) DeleteContext() map[string]string {
	// see https://github.com/libopenstorage/secrets/commit/dde442ea20ec9d59c71cea5ee0f21eeffd17ed19
	return map[string]string{
		secrets.DestroySecret: "true",
	}
}

//
// Config utils
//

// create temp files with private key and certificate
func createCertTempFile(config map[string]interface{}, namespace string) error {
	secret := &corev1.Secret{}
	secret.Namespace = namespace

	if clientCertSecretName, ok := config[AzureClientCertPath]; ok {
		secret.Name = clientCertSecretName.(string)
		if !util.KubeCheckOptional(secret) {
			return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		clientCertFileAddr, err := writeCrtsToFile(secret.Name, namespace, secret.Data["cert"], AzureClientCertPath)
		if err != nil {
			return fmt.Errorf("can not write crt %v to file %v", AzureClientCertPath, err)
		}
		config[AzureClientCertPath] = clientCertFileAddr
	}

	return nil
}

// Version returns the current driver KMS version
// either single string or map, i.e. rotating key
func (*AzureVault) Version(kms *KMS) Version {
	return &VersionSingleSecret{kms, nil}
}

// Register Azure driver with KMS layer
func init() {
	if err := RegisterDriver(azure.Name, NewAzureVault); err != nil {
		panic(err.Error())
	}
}
