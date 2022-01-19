package kms

import (
	"context"
	"fmt"
	"os"

	ibm "github.com/IBM/keyprotect-go-client"
	"github.com/libopenstorage/secrets"
	"github.com/sirupsen/logrus"
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
	kp  *ibm.API
}

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

	// returned instance
	r := &ibmKpSecretStorage{kp: kp}
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

	const limit = 2000 // same page size value
	                   // as used by the IBM KP client lib

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
) (map[string]interface{}, error) {
	// Find the key ID
	key, err := i.getKeyByName(secretID)
	if err != nil {
		return nil, err
	}

	// Fetch the key payload ( key value ) by ID
	keyPayload, err := i.kp.GetKey(context.TODO(), key.ID)
	if (err != nil) {
		return nil, err
	}

	// Return the fetched key value
	r := map[string]interface{}{secretID: keyPayload.Payload}

	return r, nil
}

// PutSecret will associate an secretId to its secret data
// provided in the arguments and store it into the secret backend
func (i *ibmKpSecretStorage) PutSecret(
	secretID string,
	plainText map[string]interface{},
	keyContext map[string]string,
) error {
	// Import the key value into IBM KP storage
	value := plainText[secretID].(string)
	_, err := i.kp.CreateImportedStandardKey(context.TODO(), secretID, nil, value)

	if err != nil {
		return err
	}

	return nil
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
