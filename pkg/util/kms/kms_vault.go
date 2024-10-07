package kms

import (
	"fmt"
	"os"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/vault"
	corev1 "k8s.io/api/core/v1"
)

// Vault authentication config options
const (
	VaultAddr       = "VAULT_ADDR"
	VaultCaCert     = "VAULT_CACERT"
	VaultClientCert = "VAULT_CLIENT_CERT"
	VaultClientKey  = "VAULT_CLIENT_KEY"
	VaultSkipVerify = "VAULT_SKIP_VERIFY"
	VaultToken      = "VAULT_TOKEN"
	RootSecretPath  = "NOOBAA_ROOT_SECRET_PATH"
)

// Vault is a vault driver
type Vault struct {
	UID  string // NooBaa system UID
	name string // NooBaa system name
	ns   string // NooBaa system namespace
}

// NewVault is vault driver constructor
func NewVault(
	name string,
	namespace string,
	uid string,
) Driver {
	return &Vault{uid, name, namespace}
}

//
// Vault specific driver for root master key
//

// Config returns this driver secret config
func (*Vault) Config(config map[string]string, tokenSecretName, namespace string) (map[string]interface{}, error) {
	// create a type correct copy of the configuration
	c := make(map[string]interface{})
	for k, v := range config {
		c[k] = v
	}

	// create TLS files out of secrets
	// replace in vaultConfig secret names with local temp files paths
	err := tlsConfig(c, namespace)
	if err != nil {
		return nil, fmt.Errorf(`❌ Could not init vault tls config %q in namespace %q`, config, namespace)
	}

	// fetch vault token from the secret
	if config[vault.AuthMethod] != vault.AuthMethodKubernetes {
		secret := &corev1.Secret{}
		secret.Namespace = namespace
		secret.Name = tokenSecretName
		if !util.KubeCheck(secret) {
			return nil, fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		token := secret.StringData["token"]
		c[VaultToken] = token
	}

	return c, nil
}

// Name returns root key map key
func (v *Vault) Name() string {
	return "rootkeyb64-" + v.UID
}

// Path return vault's kv secret id
func (v *Vault) Path() string {
	return RootSecretPath + "/rootkeyb64-" + v.UID
}

// GetContext returns context used for secret get operation
func (v *Vault) GetContext() map[string]string {
	return nil
}

// SetContext returns context used for secret set operation
func (v *Vault) SetContext() map[string]string {
	return nil
}

// DeleteContext returns context used for secret delete operation
func (v *Vault) DeleteContext() map[string]string {
	// see https://github.com/libopenstorage/secrets/commit/dde442ea20ec9d59c71cea5ee0f21eeffd17ed19
	return map[string]string{
		secrets.DestroySecret: "true",
	}
}

//
// Config utils
//

// tlsConfig create temp files with TLS keys and certs
func tlsConfig(config map[string]interface{}, namespace string) error {
	secret := &corev1.Secret{}
	secret.Namespace = namespace

	if caCertSecretName, ok := config[VaultCaCert]; ok {
		secret.Name = caCertSecretName.(string)
		if !util.KubeCheckOptional(secret) {
			return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		caFileAddr, err := writeCrtsToFile(secret.Name, namespace, secret.Data["cert"], VaultCaCert)
		if err != nil {
			return fmt.Errorf("can not write crt %v to file %v", VaultCaCert, err)
		}
		config[VaultCaCert] = caFileAddr
	}

	if clientCertSecretName, ok := config[VaultClientCert]; ok {
		secret.Name = clientCertSecretName.(string)
		if !util.KubeCheckOptional(secret) {
			return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		clientCertFileAddr, err := writeCrtsToFile(secret.Name, namespace, secret.Data["cert"], VaultClientCert)
		if err != nil {
			return fmt.Errorf("can not write crt %v to file %v", VaultClientCert, err)
		}
		config[VaultClientCert] = clientCertFileAddr

	}
	if clientKeySecretName, ok := config[VaultClientKey]; ok {
		secret.Name = clientKeySecretName.(string)
		if !util.KubeCheckOptional(secret) {
			return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		clientKeyFileAddr, err := writeCrtsToFile(secret.Name, namespace, secret.Data["key"], VaultClientKey)
		if err != nil {
			return fmt.Errorf("can not write crt %v to file %v", VaultClientKey, err)
		}
		config[VaultClientKey] = clientKeyFileAddr
	}
	return nil
}

func writeCrtsToFile(secretName string, namespace string, secretValue []byte, envVarName string) (string, error) {
	// check here first the env variable
	if envVar, found := os.LookupEnv(envVarName); found && envVar != "" {
		return envVar, nil
	}

	// Generate a temp file
	file, err := os.CreateTemp("", "")
	if err != nil {
		return "", fmt.Errorf("failed to generate temp file for k8s secret %q content, %v", secretName, err)
	}

	// close the temp file when out of scope
	defer file.Close()

	// Write into a file
	err = os.WriteFile(file.Name(), secretValue, 0444)
	if err != nil {
		return "", fmt.Errorf("failed to write k8s secret %q content to a file %v", secretName, err)
	}

	// update the env var with the path
	envVarValue := file.Name()
	envVarKey := envVarName

	err = os.Setenv(envVarKey, envVarValue)
	if err != nil {
		return "", fmt.Errorf("can not set env var %v %v", envVarKey, envVarValue)
	}
	return envVarValue, nil
}

// Version returns the current driver KMS version
// either single string or map, i.e. rotating key
func (k *Vault) Version(kms *KMS) Version {
	return &VersionRotatingSecret{VersionBase{kms, nil}, k.name, k.ns}
}

// Register Vault driver with KMS layer
func init() {
	if err := RegisterDriver(vault.Name, NewVault); err != nil {
		panic(err.Error())
	}
}
