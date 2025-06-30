package kms

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/libopenstorage/secrets"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

// Version extracts version specific code
// for two existing KMS models: single secret and
// rotating secret, a.k.a. map
// Those two flavors implement KMS SecretStorage interface
type Version interface {
	SecretStorage
	Upgrade() error
}

// VersionBase contains the base fields
// of both string and map models of the KMS
type VersionBase struct {
	k    *KMS        // Pointer to the parent KMS
	data interface{} // Secret data stored in the KMS
	//Could be either string or map
}

// VersionSingleSecret implements Version interface
// for the single string KMS secret
type VersionSingleSecret VersionBase

// Reconcile sets the single string master root key
// with the system reconciler
func (v *VersionSingleSecret) Reconcile(r SecretReconciler) error {
	return r.ReconcileSecretString(v.data.(string))
}

// Get implements SecretStorage interface for single string secret
func (v *VersionSingleSecret) Get() error {
	s, _, err := v.k.GetSecret(v.k.driver.Path(), v.k.driver.GetContext())
	if err != nil {
		// handle k8s get from non-existent secret
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			return secrets.ErrInvalidSecretId
		}
		return err
	}

	value, ok := s[v.k.driver.Name()]
	if ok {
		switch value.(type) {
		case string:
			v.data = s[v.k.driver.Name()].(string)
			return nil
		}
	}
	return secrets.ErrInvalidSecretId
}

// Set implements SecretStorage interface for single string secret
func (v *VersionSingleSecret) Set(val string) error {
	data := map[string]interface{}{
		v.k.driver.Name(): val,
	}

	v.data = val
	_, err := v.k.PutSecret(v.k.driver.Path(), data, v.k.driver.SetContext())
	return err
}

// Delete implements SecretStorage interface for single string secret
func (v *VersionSingleSecret) Delete() error {
	return v.k.DeleteSecret(v.k.driver.Path(), v.k.driver.DeleteContext())
}

// Upgrade implements SecretStorage interface for single string secret
func (v *VersionSingleSecret) Upgrade() error {
	// NOOP
	return nil
}

// VersionRotatingSecret implements SecretStorage interface
// for the rotating root master key modeled as map
type VersionRotatingSecret struct {
	VersionBase
	name string
	ns   string
}

const (
	// ActiveRootKey - pointer to the current key name
	ActiveRootKey = "active_root_key"
)

// Reconcile sets the secret map, i.e. rotating master root key
// with the system reconciler
func (v *VersionRotatingSecret) Reconcile(r SecretReconciler) error {
	return r.ReconcileSecretMap(v.data.(map[string]string))
}

// Get implements SecretStorage interface for the secret map, i.e. rotating master root key
func (v *VersionRotatingSecret) Get() error {
	s, _, err := v.k.GetSecret(v.BackendSecretName(), v.k.driver.GetContext())
	if err != nil {
		// handle k8s get from non-existent secret
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			return secrets.ErrInvalidSecretId
		}
		return err
	}

	if (v.k.driver.Name() == "KMIPSecret") {
		encodedData, ok := s[v.BackendSecretName()]
		if !ok {
			return secrets.ErrInvalidSecretData
		}
		data := map[string]string{}
		decodedString, err := base64.StdEncoding.DecodeString(encodedData.(string))
		if err != nil {
			return secrets.ErrInvalidSecretData
		}
		err = json.Unmarshal(decodedString, &data)
		if err != nil {
			return secrets.ErrInvalidSecretData
		}
		v.data = data
		return nil
	}
	rc := map[string]string{}
	for k, v := range s {
		rc[k] = v.(string)
	}
	v.data = rc
	return nil
}

// BackendSecretName returns the rotating secret backend secret name
func (v *VersionRotatingSecret) BackendSecretName() string {
	return v.name + "-root-master-key-backend"
}

// Set implements SecretStorage interface for the secret map, i.e. rotating master root key
func (v *VersionRotatingSecret) Set(val string) error {
	key := keyName()
	var s map[string]string
	if v.data == nil {
		s = map[string]string{}
	} else {
		s = v.data.(map[string]string)
	}
	s[ActiveRootKey] = key
	s[key] = val
	v.data = s
	var err error
	if (v.k.driver.Name() == "KMIPSecret") {
		jsonData, err := json.Marshal(s)
		encodedString := base64.StdEncoding.EncodeToString(jsonData)
		if err != nil {
			return err
		}
		_, err = v.k.PutSecret(v.BackendSecretName(), map[string]interface{}{v.BackendSecretName(): encodedString}, v.k.driver.SetContext())
		return err
	}
	_, err = v.k.PutSecret(v.BackendSecretName(), toInterfaceMap(s), v.k.driver.SetContext())
	return err
}

// deleteSingleStringSecret removes old format secret during upgrade
func (v *VersionRotatingSecret) deleteSingleStringSecret() bool {
	// Make sure single secret backend is deleted
	v1BackendSecret := &corev1.Secret{}
	v1BackendSecret.Name = v.k.driver.Path()
	v1BackendSecret.Namespace = v.ns
	return util.KubeDelete(v1BackendSecret)
}

// Delete implements SecretStorage interface for the secret map, i.e. rotating master root key
func (v *VersionRotatingSecret) Delete() error {
	// Delete rotating secret backend
	backendSecret := &corev1.Secret{}
	backendSecret.Name = v.BackendSecretName()
	backendSecret.Namespace = v.ns
	if !util.KubeDelete(backendSecret) {
		return fmt.Errorf("KMS Delete error for the rotating master root secret backend")
	}

	err := v.k.DeleteSecret(v.BackendSecretName(), v.k.driver.DeleteContext())
	if err != nil {
		return err
	}

	return nil
}

// Upgrade implements SecretStorage interface for the secret map, i.e. rotating master root key
func (v *VersionRotatingSecret) Upgrade() error {
	v1 := VersionSingleSecret{v.k, nil}
	err := v1.Get()
	if err != nil {
		if err == secrets.ErrInvalidSecretId {
			return nil // nothing to upgrade
		}
		return fmt.Errorf("KMS Upgrade Get error %w", err)
	}

	err = v.k.Set(v1.data.(string))
	if err != nil {
		return fmt.Errorf("KMS Upgrade Set error %w", err)
	}

	err = v.k.DeleteSecret(v.k.driver.Path(), v.k.driver.DeleteContext())
	if err != nil {
		return fmt.Errorf("KMS Upgrade Delete error %w", err)
	}

	// Make sure single secret backend is deleted
	if !v.deleteSingleStringSecret() {
		return fmt.Errorf("KMS Delete error for single secret")
	}

	return nil
}

// keyName generates a new timestamped key name,
// for secret map, i.e. rotating master root key
func keyName() string {
	return fmt.Sprintf("key-%v", time.Now().UnixNano())
}

// RemoveOldKeysFromSecret removes keys that are older than given time, and will leave at least min_keys keys
func RemoveOldKeysFromSecret(data map[string]string, givenTime time.Time, minKeys int) (int, error) {
	leftKeys := len(data)
	if (leftKeys <= minKeys) {
		return 0, nil
	}
	var entries []struct {
		key   string
		value int64
	}
	for k := range data {
		if k == ActiveRootKey {
			continue
		}
		// We want to split "key-<timestamp>" to get the timestamp
		splitKey := strings.Split(k, "-");
		if len(splitKey) != 2 {
			return 0, fmt.Errorf("KMS Key is not in the correct format")
		}	
		keyDate, err := strconv.ParseInt(splitKey[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("KMS Key is not in the correct format %w", err)
		}
		entries = append(entries, struct {
			key   string
			value int64
		}{key: k, value: keyDate})
	}
	// Sort entries by timestamp (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].value < entries[j].value
	})
	deletedKeys := 0
	leftKeys = len(entries)
	for _, entry := range entries {
		if leftKeys <= minKeys {
			break
		}
		old_key_timestamp := time.Unix(0, entry.value)
		if old_key_timestamp.Before(givenTime) {
			delete(data, entry.key)
			leftKeys--
			deletedKeys++
		}
	}
	return deletedKeys, nil
}

// toInterfaceMap converts map of string to string to map of string to interface
func toInterfaceMap(s map[string]string) map[string]interface{} {
	data := map[string]interface{}{}
	for k, v := range s {
		data[k] = v
	}
	return data
}
