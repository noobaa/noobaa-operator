package kms

import (
	"context"
	"fmt"
	"os"
	"strings"

	ibm "github.com/IBM/keyprotect-go-client"
	"github.com/libopenstorage/secrets"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

const (
	// IbmKpSecretStorageName is KMS backend name
	IbmKpSecretStorageName = "ibmkeyprotect"
	// IbmServiceAPIKey is the service ID API Key
	IbmServiceAPIKey = "IBM_KP_SERVICE_API_KEY"
	// IbmInstanceIDKey is the Key Protect Service's Instance ID
	IbmInstanceIDKey = "IBM_KP_SERVICE_INSTANCE_ID"
	// IbmBaseURLKey is the Key Protect Service's Base URL
	IbmBaseURLKey = "IBM_KP_BASE_URL"
	// IbmTokenURLKey is the Key Protect Service's Token URL
	IbmTokenURLKey = "IBM_KP_TOKEN_URL"
	// kpClientTimeout is key protect client network timeout
	kpClientTimeout = 10
)

// ibmKpSecretStorage is a IBM KP backend Key Management Systems (KMS)
// which implements libopenstorage Secrets interface
// using IBM KP secret storage - standard keys interface using
// latest "github.com/IBM/keyprotect-go-client" version 0.7.0
type ibmKpSecretStorage struct {
	kp     *ibm.API
	secret *corev1.Secret
}

const (
	// IBMKPActiveKeyID is the key in the secret that stores the active key ID
	IBMKPActiveKeyID = "IBM_KP_ACTIVE_KEY_ID"
	// IBMKPKeyPrefix is the prefix for key IDs stored in the secret
	IBMKPKeyPrefix = "IBM_KP_KEY_"
)

func getIbmParam(secretConfig map[string]interface{}, name string) string {
	if tokenIntf, exists := secretConfig[name]; exists {
		return tokenIntf.(string)
	}

	return os.Getenv(name)
}

// NewIBMKpSecretStorage is a constructor, returns a new instance of ibmKpSecretStorage
func NewIBMKpSecretStorage(
	secretConfig map[string]interface{}, // config
) (secrets.Secrets, error) {
	serviceAPIKey := getIbmParam(secretConfig, IbmServiceAPIKey)
	if serviceAPIKey == "" {
		return nil, fmt.Errorf("%v is not set", IbmServiceAPIKey)
	}

	instanceID := getIbmParam(secretConfig, IbmInstanceIDKey)
	if instanceID == "" {
		return nil, fmt.Errorf("%v is not set", IbmInstanceIDKey)
	}

	baseURL := getIbmParam(secretConfig, IbmBaseURLKey)
	if baseURL == "" {
		baseURL = ibm.DefaultBaseURL
	}

	tokenURL := getIbmParam(secretConfig, IbmTokenURLKey)
	if tokenURL == "" {
		tokenURL = ibm.DefaultTokenURL
	}

	cc := ibm.ClientConfig{
		BaseURL:    baseURL,
		APIKey:     serviceAPIKey,
		TokenURL:   tokenURL,
		InstanceID: instanceID,
		Verbose:    ibm.VerboseAll,
		Timeout:    kpClientTimeout,
	}
	kp, err := ibm.NewWithLogger(cc, nil, logrus.StandardLogger())
	if err != nil {
		return nil, err
	}

	secret, exists := secretConfig[IBMSecret]
	if !exists {
		return nil, fmt.Errorf("Missing IBM secret")
	}

	// returned instance
	r := &ibmKpSecretStorage{
		kp:     kp,
		secret: secret.(*corev1.Secret),
	}
	return r, nil
}

// String representation of this implementation
func (i *ibmKpSecretStorage) String() string {
	return IbmKpSecretStorageName
}

// getKeyByName finds the key, specifically ID by name
// by iterating over of all keys
func (i *ibmKpSecretStorage) getKeyByName(
	secretID string,
) (*ibm.Key, error) {

	// same page size value as used by the IBM KP client lib
	const limit = 2000

	for pageNo := 0; ; pageNo++ {
		// Get current page keys
		offset := pageNo * limit
		keys, err := i.kp.GetKeys(context.TODO(), limit, offset)
		if err != nil {
			return nil, err
		}

		// No more keys, key is not found
		// out of here
		if len(keys.Keys) == 0 {
			return nil, secrets.ErrInvalidSecretId
		}

		// Find key by name
		for _, k := range keys.Keys {
			if k.Name == secretID {
				return &k, nil
			}
		}
	}
}

// GetSecret returns the secret data associated with the
// supplied secretId.
func (i *ibmKpSecretStorage) GetSecret(
	secretID string,
	keyContext map[string]string,
) (map[string]interface{}, secrets.Version, error) {
	log := util.Logger()

	// Check if this is a rotating secret (backend secret)
	var activeKeyID string
	util.KubeCheck(i.secret)
	if strings.HasSuffix(secretID, "-root-master-key-backend") {
		// Rotating secret format - get the active key ID from the tracking secret
		exists := false
		activeKeyID, exists = i.secret.StringData[IBMKPActiveKeyID]
		if !exists {
			log.Errorf("IBM KeyProtect GetSecret() activeKeyID does not exist in secret %v", i.secret.Name)
			return nil, secrets.NoVersion, secrets.ErrInvalidSecretId
		}
	}

	// Determine which key name to look for
	var keyNameToFind string
	if len(activeKeyID) > 0 {
		// Rotating format: look for the IBM KP key ID stored in the secret
		keyIDInKP, exists := i.secret.StringData[IBMKPKeyPrefix+activeKeyID]
		if !exists {
			log.Errorf("IBM KeyProtect GetSecret() key ID mapping not found for %v", activeKeyID)
			return nil, secrets.NoVersion, secrets.ErrInvalidSecretId
		}
		keyNameToFind = keyIDInKP
	} else {
		// Single format (for backward compatibility during upgrade)
		keyNameToFind = secretID
	}

	// Find the key by name in IBM KeyProtect
	key, err := i.getKeyByName(keyNameToFind)
	if err != nil {
		log.Errorf("IBM KeyProtect GetSecret failed to find key by name %v: %v", keyNameToFind, err)
		return nil, secrets.NoVersion, err
	}

	// Fetch the key payload (key value) by ID
	keyPayload, err := i.kp.GetKey(context.TODO(), key.ID)
	if err != nil {
		log.Errorf("IBM KeyProtect GetSecret failed to get key %v: %v", key.ID, err)
		return nil, secrets.NoVersion, err
	}

	// Return in the appropriate format
	if len(activeKeyID) > 0 {
		// Rotating format: return active_root_key pointer and the key value
		r := map[string]interface{}{ActiveRootKey: activeKeyID, activeKeyID: keyPayload.Payload}
		return r, secrets.NoVersion, nil
	} else {
		// Single format: return just the key value, Might not need
		r := map[string]interface{}{secretID: keyPayload.Payload}
		return r, secrets.NoVersion, nil
	}
}

// PutSecret will associate an secretId to its secret data
// provided in the arguments and store it into the secret backend
func (i *ibmKpSecretStorage) PutSecret(
	secretID string,
	plainText map[string]interface{},
	keyContext map[string]string,
) (secrets.Version, error) {
	log := util.Logger()

	// Check if this is a rotating secret format
	var activeKey string
	var value string
	if strings.HasSuffix(secretID, "-root-master-key-backend") {
		// Rotating secret format - extract active key and its value
		activeKeyIntf, exists := plainText[ActiveRootKey]
		if !exists {
			log.Errorf("IBM KeyProtect PutSecret missing active_root_key in plainText")
			return secrets.NoVersion, fmt.Errorf("missing active_root_key in rotating secret format")
		}
		activeKey = activeKeyIntf.(string)

		valueIntf, exists := plainText[activeKey]
		if !exists {
			log.Errorf("IBM KeyProtect PutSecret missing value for active key %v", activeKey)
			return secrets.NoVersion, fmt.Errorf("missing value for active key %v", activeKey)
		}
		value = valueIntf.(string)
	} else {
		// Single format (for backward compatibility)
		valueIntf, ok := plainText[secretID]
		if !ok {
			log.Errorf("IBM KeyProtect PutSecret failed to get value for secretID %v", secretID)
			return secrets.NoVersion, fmt.Errorf("invalid secret value format for secretID %v", secretID)
		}
		value = valueIntf.(string)
	}

	// Generate a unique key name in IBM KeyProtect
	// For rotating keys, use the active key name; for single keys, use secretID
	var keyNameInKP string
	if len(activeKey) > 0 {
		// Use the timestamp-based key name from the active key
		keyNameInKP = activeKey
	} else {
		keyNameInKP = secretID
	}

	// Import the key value into IBM KP storage
	key, err := i.kp.CreateImportedStandardKey(context.TODO(), keyNameInKP, nil, value)
	if err != nil {
		log.Errorf("IBM KeyProtect PutSecret failed to create key %v: %v", keyNameInKP, err)
		return secrets.NoVersion, err
	}

	// For rotating format, store the mapping in the tracking secret
	if len(activeKey) > 0 {
		i.secret.StringData[IBMKPActiveKeyID] = activeKey
		i.secret.StringData[IBMKPKeyPrefix+activeKey] = key.Name

		if !util.KubeUpdate(i.secret) {
			log.Errorf("Failed to update IBM KP tracking secret %v in ns %v", i.secret.Name, i.secret.Namespace)
			return secrets.NoVersion, fmt.Errorf("failed to update IBM KP tracking secret %v in ns %v", i.secret.Name, i.secret.Namespace)
		}
	}

	log.Infof("IBM KeyProtect PutSecret successfully created key %v (IBM KP name: %v)", activeKey, keyNameInKP)
	return secrets.NoVersion, nil
}

// DeleteSecret deletes the secret data associated with the
// supplied secretId.
func (i *ibmKpSecretStorage) DeleteSecret(
	secretID string,
	keyContext map[string]string,
) error {
	// Find the key ID
	k, err := i.getKeyByName(secretID)
	if err != nil {
		return err
	}

	// Delete the key by ID
	_, err = i.kp.DeleteKey(context.TODO(), k.ID, ibm.ReturnRepresentation, []ibm.CallOpt{ibm.ForceOpt{Force: true}}...)
	if err != nil {
		return err
	}

	return nil
}

// ListSecrets is no supported
func (*ibmKpSecretStorage) ListSecrets() ([]string, error) {
	return nil, secrets.ErrNotSupported
}

// Encrypt is no supported
func (i *ibmKpSecretStorage) Encrypt(
	secretID string,
	plaintTextData string,
	keyContext map[string]string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

// Decrypt is no supported
func (i *ibmKpSecretStorage) Decrypt(
	secretID string,
	encryptedData string,
	keyContext map[string]string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

// Rencrypt is no supported
func (i *ibmKpSecretStorage) Rencrypt(
	originalSecretID string,
	newSecretID string,
	originalKeyContext map[string]string,
	newKeyContext map[string]string,
	encryptedData string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

// Register ibmKpSecretStorage backend with libopenstorage secrets layer
func init() {
	if err := secrets.Register(IbmKpSecretStorageName, NewIBMKpSecretStorage); err != nil {
		panic(err.Error())
	}
}
