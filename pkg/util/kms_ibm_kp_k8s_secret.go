package util

import (
	"fmt"

	"github.com/libopenstorage/secrets"
	"github.com/libopenstorage/secrets/ibm"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"

	corev1 "k8s.io/api/core/v1"
)

const (
	// IbmKpK8sSecretName is KMS backend name
	IbmKpK8sSecretName = ibm.Name + "-k8s-secret"

	// configuration keys
	secretName       = "K8S_SECRET_NAME"
	secretNamespace  = "K8S_SECRET_NAMESPACE"
)

// ibmKpK8sSecret is a IBM KP backend Key Management Systems (KMS)
// which implements libopenstorage Secrets interface
// in terms of libopenstorage/secrets/ibm client with in-memory encrypted secret storage
// and k8s secret encrypted secrets persistence
type ibmKpK8sSecret struct {
	ibmKp secrets.Secrets       // uses in memory kvdb storage
	tokenSecret *corev1.Secret  // store encrypted secrets here
}

// NewIBMKpK8sSecret is a constructor, returns a new instance of ibmKpK8sSecret
func NewIBMKpK8sSecret(
	c map[string]interface{}, // config
) (secrets.Secrets, error) {
	// returned instance
	r := &ibmKpK8sSecret{}

	// Create in-memory kvdb, used for encrypted secret storage by openstorage ibm-kp layer
	kv, err := kvdb.New(mem.Name, "noobaa/", nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("❌ Unable to create a IBM Key Protect kvdb instance: %v", err)
	}
	c[ibm.IbmKvdbKey] = kv

	// Create libopenstorage ibm kp client with in-memory encrypted secret storage
	ibmKp, err := secrets.New(secrets.TypeIBM, c)
	if err != nil {
		return nil, err
	}
	r.ibmKp = ibmKp

	// k8s secret where encrypted secrets are stored
	r.tokenSecret = &corev1.Secret{}
	r.tokenSecret.Name = c[secretName].(string)
	r.tokenSecret.Namespace = c[secretNamespace].(string)

	return r, nil
}

// String representation of this implementation
func (i *ibmKpK8sSecret) String() string {
	return IbmKpK8sSecretName
}

// public key context is used for encrypted secrets access
var publicCtx = map[string]string{ secrets.PublicSecretData: "true"}

// encryptedSecretKeyName returns key name for secretId's encrypted secret
// in k8s secret data map
func encryptedSecretKeyName(secretID string) string {
	return "wdek_" + secretID
}

// GetSecret returns the secret data associated with the
// supplied secretId.
func (i *ibmKpK8sSecret) GetSecret(
	secretID string,
	keyContext map[string]string,
) (map[string]interface{}, error) {
	// Fetch encrypted secret from K8S Secret persistence layer
	if _, _, err := KubeGet(i.tokenSecret); err != nil {
		return nil, err
	}
	encryptedSecret, ok := i.tokenSecret.Data[encryptedSecretKeyName(secretID)]
	if !ok {
		return nil, secrets.ErrInvalidSecretId
	}

	// If the encrypted secret is missing in the ibm-kp's in-memory kvdb
	// sync it using the value fetched from K8S secret
	// this value is used to Unwrap the plaintext secret
	if _, err := i.ibmKp.GetSecret(secretID, publicCtx); err == secrets.ErrInvalidSecretId {
		publicData := make(map[string]interface{})
		publicData[secretID] = encryptedSecret

		err := i.ibmKp.PutSecret(secretID, publicData, publicCtx);
		if  err != nil {
			return nil, err
		}
	} else if err != nil { // Unknown GetSecret() error
		return nil, err
	}

	// Unwrap the plaintext secret using IBM KP client
	return i.ibmKp.GetSecret(secretID, keyContext)
}

// PutSecret will associate an secretId to its secret data
// provided in the arguments and store it into the secret backend
func (i *ibmKpK8sSecret) PutSecret(
	secretID string,
	plainText map[string]interface{},
	keyContext map[string]string,
) error {
	// Wrap the plaintext secret using IBM KP client
	err := i.ibmKp.PutSecret(secretID, plainText, keyContext);
	if  err != nil {
		return err
	}

	// Fetch the encrypted secret value from ibmkp kvdb in-memory storage
	encryptedSecret, err := i.ibmKp.GetSecret(secretID, publicCtx);
	if  err != nil {
		return err
	}

	// Store the encrypted secret in k8s secret for future GetSecret() use
	i.tokenSecret.Data[encryptedSecretKeyName(secretID)] = encryptedSecret[secretID].([]uint8)
	if !KubeUpdate(i.tokenSecret) {
		return fmt.Errorf("❌ KMS IBM KP PutSecret Failed to update encrypted secret")
	}

	return nil
}

// DeleteSecret deletes the secret data associated with the
// supplied secretId.
func (i *ibmKpK8sSecret) DeleteSecret(
	secretID string,
	keyContext map[string]string,
) error {
	// Fetch DEK from K8S Secret persistence layer
	if _, _, err := KubeGet(i.tokenSecret); err != nil {
		return err
	}

	// Remove the encrypted secret from k8s secret
	delete(i.tokenSecret.Data, encryptedSecretKeyName(secretID))
	if !KubeUpdate(i.tokenSecret) {
		return fmt.Errorf("❌ KMS IBM KP DeleteSecret Failed to update encrypted secret secret")
	}

	// Remove secretId from ibm-kp
	return i.ibmKp.DeleteSecret(secretID, keyContext)
}

// ListSecrets is not implemented
func (*ibmKpK8sSecret) ListSecrets() ([]string, error) {
	return nil, secrets.ErrNotSupported
}

// Encrypt is passed to ibm-kp
func (i *ibmKpK8sSecret) Encrypt(
	secretID string,
	plaintTextData string,
	keyContext map[string]string,
) (string, error) {
	return i.ibmKp.Encrypt(secretID, plaintTextData, keyContext)
}

// Decrypt is passed to ibm-kp
func (i *ibmKpK8sSecret) Decrypt(
	secretID string,
	encryptedData string,
	keyContext map[string]string,
) (string, error) {
	return i.ibmKp.Decrypt(secretID, encryptedData, keyContext)
}

// Rencrypt is passed to ibm-kp
func (i *ibmKpK8sSecret) Rencrypt(
	originalSecretID string,
	newSecretID string,
	originalKeyContext map[string]string,
	newKeyContext map[string]string,
	encryptedData string,
) (string, error) {
	return i.ibmKp.Rencrypt(originalSecretID, newSecretID, originalKeyContext, newKeyContext, encryptedData)
}

// Register ibmKpK8sSecret backend with libopenstorage secrets layer
func init() {
	if err := secrets.Register(IbmKpK8sSecretName, NewIBMKpK8sSecret); err != nil {
		panic(err.Error())
	}
}
