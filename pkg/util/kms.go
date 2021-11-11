package util

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/vault"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	corev1 "k8s.io/api/core/v1"
)

///////////////////////////////////
/////////// VAULT UTILS ///////////
///////////////////////////////////
const (
	rootSecretPath          = "NOOBAA_ROOT_SECRET_PATH"
	vaultCaCert             = "VAULT_CACERT"
	vaultClientCert         = "VAULT_CLIENT_CERT"
	vaultClientKey          = "VAULT_CLIENT_KEY"
	VaultAddr               = "VAULT_ADDR"
	vaultCaPath             = "VAULT_CAPATH"
	VaultBackendPath        = "VAULT_BACKEND_PATH"
	vaultToken              = "VAULT_TOKEN"
	KmsProvider             = "KMS_PROVIDER"
	KmsProviderVault        = "vault"
	defaultVaultBackendPath = "secret/"
)


// VerifyExternalSecretsDeletion checks if noobaa is on un-installation process
// if true, deletes secrets from external KMS
func VerifyExternalSecretsDeletion(kms nbv1.KeyManagementServiceSpec, namespace string, uid string) error {

	if len(kms.ConnectionDetails) == 0 {
		log.Infof("deleting root key locally")
		return nil
	}

	if !isVaultKMS(kms.ConnectionDetails[KmsProvider]) {
		log.Errorf("Unsupported KMS provider %v", kms.ConnectionDetails[KmsProvider])
		return fmt.Errorf("Unsupported KMS provider %v", kms.ConnectionDetails[KmsProvider])
	}

	c, err := InitVaultClient(kms.ConnectionDetails, kms.TokenSecretName, namespace)
	if err != nil {
		log.Errorf("deleting root key externally failed: init vault client: %v", err)
		return err
	}

	secretPath := BuildExternalSecretPath(kms, uid)
	err = DeleteSecret(c, secretPath)
	if err != nil {
		log.Errorf("deleting root key externally failed: %v", err)
		return err
	}

	return nil
}

// InitVaultClient inits the secret store
func InitVaultClient(config map[string]string, tokenSecretName string, namespace string) (secrets.Secrets, error) {

	// create a type correct copy of the configuration
	vaultConfig := make(map[string]interface{})
	for k, v := range config {
		vaultConfig[k] = v
	}

	// create TLS files out of secrets
	// replace in vaultConfig secret names with local temp files paths
	err := tlsConfig(vaultConfig, namespace)
	if err != nil {
		return nil, fmt.Errorf(`❌ Could not init vault tls config %q in namespace %q`, config, namespace)
	}

	// veify backend path, use default value if not set
	if b, ok := config[VaultBackendPath]; !ok || b == "" {
		log.Infof("KMS: using default backend path %v", defaultVaultBackendPath)
		vaultConfig[VaultBackendPath] = defaultVaultBackendPath
	}

	// fetch vault token from the secret
	secret := KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	secret.Namespace = namespace
	secret.Name = tokenSecretName
	if !KubeCheck(secret) {
		return nil, fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
	}
	token := secret.StringData["token"]
	vaultConfig[vaultToken] = token

	return vault.New(vaultConfig)
}

// tlsConfig create temp files with TLS keys and certs
func tlsConfig(config map[string]interface{}, namespace string) (error) {
	secret := KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	secret.Namespace = namespace

	if caCertSecretName, ok := config[vaultCaCert]; ok {
		secret.Name = caCertSecretName.(string)
		if !KubeCheckOptional(secret) {
			return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		caFileAddr, err := writeCrtsToFile(secret.Name, namespace, secret.Data["cert"], vaultCaCert)
		if err != nil {
			return fmt.Errorf("can not write crt %v to file %v", vaultCaCert, err)
		}
		config[vaultCaCert] = caFileAddr
	}

	if clientCertSecretName, ok := config[vaultClientCert]; ok {
		secret.Name = clientCertSecretName.(string)
		if !KubeCheckOptional(secret) {
			return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		clientCertFileAddr, err := writeCrtsToFile(secret.Name, namespace, secret.Data["cert"], vaultClientCert)
		if err != nil {
			return fmt.Errorf("can not write crt %v to file %v", vaultClientCert, err)
		}
		config[vaultClientCert] = clientCertFileAddr

	}
	if clientKeySecretName, ok := config[vaultClientKey]; ok {
		secret.Name = clientKeySecretName.(string)
		if !KubeCheckOptional(secret) {
			return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
		}
		clientKeyFileAddr, err := writeCrtsToFile(secret.Name, namespace, secret.Data["key"], vaultClientKey)
		if err != nil {
			return fmt.Errorf("can not write crt %v to file %v", vaultClientKey, err)
		}
		config[vaultClientKey] = clientKeyFileAddr
	}
	return nil
}

// PutSecret writes the secret to the secrets store
func PutSecret(client secrets.Secrets, secretName, secretValue, secretPath string) error {

	keyContext := map[string]string{}
	data := make(map[string]interface{})
	data[secretName] = secretValue

	err := client.PutSecret(secretPath, data, keyContext)
	if err != nil {
		log.Errorf("KMS PutSecret: secret path %v value %v, error %v", secretPath, secretValue, err)
		return err
	}

	return nil
}

// GetSecret reads the secret to the secrets store
func GetSecret(client secrets.Secrets, secretName, secretPath string) (string, error) {
	keyContext := map[string]string{}
	s, err := client.GetSecret(secretPath, keyContext)
	if err != nil {
		log.Errorf("KMS GetSecret: secret path %v, error %v", secretPath, err)
		return "", err
	}

	return s[secretName].(string), nil
}

// DeleteSecret deletes the secret from the secrets store
func DeleteSecret(client secrets.Secrets, secretPath string) error {
	keyContext := map[string]string{}
	err := client.DeleteSecret(secretPath, keyContext)
	if err != nil {
		log.Errorf("KMS DeleteSecret: secret path %v, error %v", secretPath, err)
		return err
	}
	return nil
}

// BuildExternalSecretPath builds a string that specifies the root key secret path
func BuildExternalSecretPath(kms nbv1.KeyManagementServiceSpec, uid string) (string) {
	secretPath := rootSecretPath + "/rootkeyb64-" + uid
	return secretPath
}

// isVaultKMS return true if kms provider is vault
func isVaultKMS(provider string) bool {
	return provider == KmsProviderVault
}

// ValidateConnectionDetails return error if kms connection details are faulty
func ValidateConnectionDetails(config map[string]string, tokenSecretName string, namespace string) error {
	// validate auth token
	if tokenSecretName == "" {
		return fmt.Errorf("kms token is missing")
	}
	secret := KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	secret.Namespace = namespace
	secret.Name = tokenSecretName

	if !KubeCheck(secret) {
		return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
	}

	token, ok := secret.StringData["token"]
	if !ok || token == "" {
		return fmt.Errorf("kms token in token secret is missing")
	}
	// validate connection details
	providerType := config[KmsProvider]
	if !isVaultKMS(providerType) {
		return fmt.Errorf("Unsupported kms type: %v", providerType)
	}

	return validateVaultConnectionDetails(config, tokenSecretName, namespace)
}

// validateVaultConnectionDetails return error if vault connection details are faulty
func validateVaultConnectionDetails(config map[string]string, tokenName string, namespace string) error {
	if addr, ok := config[VaultAddr]; !ok || addr == "" {
		return fmt.Errorf("failed to validate vault connection details: vault address is missing")
	}
	if capPath, ok := config[vaultCaPath]; ok && capPath != "" {
		// We do not support a directory with multiple CA since we fetch a k8s Secret and read its content
		// So we operate with a single CA only
		return fmt.Errorf("failed to validate vault connection details: multiple CA is unsupported")
	}
	secret := KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	secret.Namespace = namespace

	vaultTLSConnectionDetailsMap := map[string]string{vaultCaCert: "cert",
		vaultClientCert: "cert", vaultClientKey: "key"}

	for tlsOption, fieldInSecret := range vaultTLSConnectionDetailsMap {
		tlsOptionSecretName, ok := config[tlsOption]
		if ok && tlsOptionSecretName != "" {
			secret.Name = tlsOptionSecretName
			if !KubeCheckOptional(secret) {
				return fmt.Errorf(`❌ Could not find secret %q in namespace %q`, secret.Name, secret.Namespace)
			}
			if tlsOptionValue, ok := secret.Data[fieldInSecret]; !ok || len(tlsOptionValue) == 0 {
				return fmt.Errorf("failed to validate vault connection details: vault %v is missing in secret %q in namespace %q",
					tlsOption, secret.Name, secret.Namespace)
			}
		}
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
