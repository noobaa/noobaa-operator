package util

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/vault"
	corev1 "k8s.io/api/core/v1"
)

// Vault authentication config options
const (
	VaultAddr               = "VAULT_ADDR"
	VaultCaCert             = "VAULT_CACERT"
	VaultClientCert         = "VAULT_CLIENT_CERT"
	VaultClientKey          = "VAULT_CLIENT_KEY"
	VaultSkipVerify         = "VAULT_SKIP_VERIFY"
	VaultToken              = "VAULT_TOKEN"
	RootSecretPath          = "NOOBAA_ROOT_SECRET_PATH"
)

// KMSVault is a vault driver
type KMSVault struct {
	UID string  // NooBaa system UID
}

//
// Vault specific driver for root master key
//

// Config returns this driver secret config
func (*KMSVault) Config(config map[string]string, tokenSecretName, namespace string) (map[string]interface{}, error) {
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
		if !KubeCheck(secret) {
			return nil, fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		token := secret.StringData["token"]
		c[VaultToken] = token
	}

	log.Infof("KMS vault config: %v", c)
	return c, nil
}

// Name returns root key map key
func (k *KMSVault) Name() string {
	return "rootkeyb64-" + k.UID
}

// Path return vault's kv secret id
func (k *KMSVault) Path() string {
	return RootSecretPath + "/rootkeyb64-" + k.UID
}

// GetContext returns context used for secret get operation
func (k *KMSVault) GetContext() map[string]string {
	return nil
}

// SetContext returns context used for secret set operation
func (k *KMSVault) SetContext() map[string]string {
	return nil
}

// DeleteContext returns context used for secret delete operation
func (k *KMSVault) DeleteContext() map[string]string {
	// see https://github.com/libopenstorage/secrets/commit/dde442ea20ec9d59c71cea5ee0f21eeffd17ed19
	return map[string]string{
		secrets.DestroySecret: "true",
	}
}

//
// Config utils
//

// tlsConfig create temp files with TLS keys and certs
func tlsConfig(config map[string]interface{}, namespace string) (error) {
	secret := &corev1.Secret{}
	secret.Namespace = namespace

	if caCertSecretName, ok := config[VaultCaCert]; ok {
		secret.Name = caCertSecretName.(string)
		if !KubeCheckOptional(secret) {
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
		if !KubeCheckOptional(secret) {
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
		if !KubeCheckOptional(secret) {
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
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return "", fmt.Errorf("failed to generate temp file for k8s secret %q content, %v", secretName, err)
	}

	// Write into a file
	err = ioutil.WriteFile(file.Name(), secretValue, 0444)
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
