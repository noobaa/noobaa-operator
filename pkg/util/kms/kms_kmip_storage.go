package kms

import (
	"bufio"
	"bytes"
	"fmt"
	"time"

	"crypto/tls"
	"crypto/x509"
	"encoding/base64"

	kmip "github.com/gemalto/kmip-go"
	"github.com/gemalto/kmip-go/kmip14"
	"github.com/gemalto/kmip-go/ttlv"
	"github.com/google/uuid"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/libopenstorage/secrets"
	corev1 "k8s.io/api/core/v1"
)

const (
	// KMIPSecretStorageName is KMS backend name
	KMIPSecretStorageName = "kmip"

	// KMIPDefaulReadTimeout is the default read network timeout
	KMIPDefaulReadTimeout = 10

	// KMIPDefaulWriteTimeout is the default write network timeout
	KMIPDefaulWriteTimeout = 10

	// KMIP version
	protocolMajor = 1
	protocolMinor = 4

	// Expected secret data length in bits
	cryptographicLength = 256
)

// KMIPSecretStorage is a KMIP backend Key Management Systems (KMS)
// which implements libopenstorage Secrets interface
type KMIPSecretStorage struct {
	endpoint     string
	tlsConfig    *tls.Config
	readTimeout  int
	writeTimeout int
	secret       *corev1.Secret
}

// NewKMIPSecretStorage is a constructor, returns a new instance of KMIPSecretStorage
func NewKMIPSecretStorage(
	secretConfig map[string]interface{}, // config
) (secrets.Secrets, error) {
	var value interface{}
	var exists bool
	if value, exists = secretConfig[KMIPEndpoint]; !exists {
		return nil, fmt.Errorf("%v is not set", KMIPEndpoint)
	}
	endpoint := value.(string)
	if value, exists = secretConfig[KMIPCACERT]; !exists {
		return nil, fmt.Errorf("%v is not set", KMIPCACERT)
	}
	caCert := value.(string)
	if value, exists = secretConfig[KMIPCLIENTCERT]; !exists {
		return nil, fmt.Errorf("%v is not set", KMIPCLIENTCERT)
	}
	clientCert := value.(string)
	if value, exists = secretConfig[KMIPCLIENTKEY]; !exists {
		return nil, fmt.Errorf("%v is not set", KMIPCLIENTKEY)
	}
	clientKey := value.(string)

	// server name is optional
	serverName := ""
	if value, exists = secretConfig[KMIPTLSServerName]; exists {
		serverName = value.(string)
	}

	readTimeout := KMIPDefaulReadTimeout
	if value, exists = secretConfig[KMIPReadTimeOut]; exists {
		readTimeout = value.(int)
	}

	writeTimeout := KMIPDefaulWriteTimeout
	if value, exists = secretConfig[KMIPWriteTimeOut]; exists {
		writeTimeout = value.(int)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(caCert))
	cert, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey))
	if err != nil {
		return nil, fmt.Errorf("Invalid X509 key pair: %v", err)
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		ServerName:   serverName,
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}

	secret, exists := secretConfig[KMPSecret]
	if !exists {
		return nil, fmt.Errorf("Missing KMS secret")
	}

	// returned instance
	r := &KMIPSecretStorage{endpoint, tlsConfig, readTimeout, writeTimeout, secret.(*corev1.Secret)}
	return r, nil
}

// String representation of this implementation
func (*KMIPSecretStorage) String() string {
	return KMIPSecretStorageName
}

// Connect to the kmip endpoint, perform TLS and KMIP handshakes
func (k *KMIPSecretStorage) connect() (*tls.Conn, error) {
	conn, err := tls.Dial("tcp", k.endpoint, k.tlsConfig)
	if err != nil {
		return nil, err
	}
	if k.readTimeout != 0 {
		err = conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(k.readTimeout)))
		if err != nil {
			return nil, err
		}
	}
	if k.writeTimeout != 0 {
		err = conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(k.writeTimeout)))
		if err != nil {
			return nil, err
		}
	}
	if err = conn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	// KMIP handshake
	if err = k.discover(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// Send KMIP operation over tls connection, return response and error
func (k *KMIPSecretStorage) send(conn *tls.Conn, operation kmip14.Operation, payload interface{}) (*kmip.ResponseMessage, *ttlv.Decoder, []byte, error) {
	biID := uuid.New()

	msg := kmip.RequestMessage{
		RequestHeader: kmip.RequestHeader{
			ProtocolVersion: kmip.ProtocolVersion{
				ProtocolVersionMajor: protocolMajor,
				ProtocolVersionMinor: protocolMinor,
			},
			BatchCount: 1,
		},
		BatchItem: []kmip.RequestBatchItem{
			{
				UniqueBatchItemID: biID[:],
				Operation:         operation,
				RequestPayload:    payload,
			},
		},
	}

	req, err := ttlv.Marshal(msg)
	if err != nil {
		return nil, nil, nil, err
	}

	_, err = conn.Write(req)
	if err != nil {
		return nil, nil, nil, err
	}

	decoder := ttlv.NewDecoder(bufio.NewReader(conn))
	resp, err := decoder.NextTTLV()
	if err != nil {
		return nil, nil, nil, err
	}

	var respMsg kmip.ResponseMessage
	err = decoder.DecodeValue(&respMsg, resp)
	if err != nil {
		return nil, nil, nil, err
	}

	return &respMsg, decoder, biID[:], nil
}

// Verify the response success and return the batch item
func (k *KMIPSecretStorage) response(respMsg *kmip.ResponseMessage, operation kmip14.Operation, uniqueBatchItemID []byte) (*kmip.ResponseBatchItem, error) {
	if respMsg.ResponseHeader.BatchCount != 1 {
		return nil, fmt.Errorf("Batch count %v should be 1", respMsg.ResponseHeader.BatchCount)
	}
	if len(respMsg.BatchItem) != 1 {
		return nil, fmt.Errorf("Batch Intems list len %v should be 1", len(respMsg.BatchItem))
	}
	bi := respMsg.BatchItem[0]
	if operation != bi.Operation {
		return nil, fmt.Errorf("Unexpected operation, real %v expected %v", bi.Operation, operation)
	}
	if !bytes.Equal(uniqueBatchItemID, bi.UniqueBatchItemID) {
		return nil, fmt.Errorf("Unexpected uniqueBatchItemID, real %v expected %v", bi.UniqueBatchItemID, uniqueBatchItemID)
	}
	if kmip14.ResultStatusSuccess != bi.ResultStatus {
		return nil, fmt.Errorf("Unexpected result status %v expected success %v", bi.ResultStatus, kmip14.ResultStatusSuccess)
	}

	return &bi, nil
}

// KMIP handshake
func (k *KMIPSecretStorage) discover(conn *tls.Conn) error {
	respMsg, decoder, uniqueBatchItemID, err := k.send(conn, kmip14.OperationDiscoverVersions, kmip.DiscoverVersionsRequestPayload{
		ProtocolVersion: []kmip.ProtocolVersion{
			{ProtocolVersionMajor: protocolMajor, ProtocolVersionMinor: protocolMinor},
		},
	})
	if err != nil {
		return err
	}

	bi, err := k.response(respMsg, kmip14.OperationDiscoverVersions, uniqueBatchItemID)
	if err != nil {
		return err
	}

	var respDiscoverVersionsPayload kmip.DiscoverVersionsResponsePayload
	ttlvPayload, ok := bi.ResponsePayload.(ttlv.TTLV)
	if !ok {
		return fmt.Errorf("failed to parse responsePayload")
	}

	err = decoder.DecodeValue(&respDiscoverVersionsPayload, ttlvPayload)
	if err != nil {
		return err
	}

	if len(respDiscoverVersionsPayload.ProtocolVersion) != 1 {
		return fmt.Errorf("Invalid len of discovered protocol versions %v expected 1", len(respDiscoverVersionsPayload.ProtocolVersion))
	}
	pv := respDiscoverVersionsPayload.ProtocolVersion[0]
	if pv.ProtocolVersionMajor != protocolMajor || pv.ProtocolVersionMinor != protocolMinor {
		return fmt.Errorf("Invalid discovered protocol version %v.%v expected %v.%v", pv.ProtocolVersionMajor, pv.ProtocolVersionMinor, protocolMajor, protocolMinor)
	}
	return nil
}

// GetSecret returns the secret data associated with the
// supplied secretId.
func (k *KMIPSecretStorage) GetSecret(
	secretID string,
	keyContext map[string]string,
) (map[string]interface{}, error) {

	log := util.Logger()

	// KMIP key uniqueIdentifier
	uniqueIdentifier, exists := k.secret.StringData[KMIPUniqueID]
	if !exists {
		log.Errorf("KMIPSecretStorage.GetSecret() uniqueIdentifier %v does not exist in secret %v", KMIPUniqueID, k.secret)
		return nil, secrets.ErrInvalidSecretId
	}

	conn, err := k.connect()
	if err != nil {
		log.Errorf("KMIPSecretStorage.GetSecret() failed to connect %v", err)
		return nil, err
	}
	defer conn.Close()

	respMsg, decoder, uniqueBatchItemID, err := k.send(conn, kmip14.OperationGet, kmip.GetRequestPayload{
		UniqueIdentifier: uniqueIdentifier,
	})
	if err != nil {
		log.Errorf("KMIPSecretStorage.GetSecret() failed to send %v", err)
		return nil, err
	}
	bi, err := k.response(respMsg, kmip14.OperationGet, uniqueBatchItemID)
	if err != nil {
		log.Errorf("KMIPSecretStorage.GetSecret() failed to verify response %v", err)
		return nil, err
	}
	var getRespPayload kmip.GetResponsePayload
	err = decoder.DecodeValue(&getRespPayload, bi.ResponsePayload.(ttlv.TTLV))
	if err != nil {
		log.Errorf("KMIPSecretStorage.GetSecret() failed to decode response payload %v", err)
		return nil, err
	}
	if getRespPayload.UniqueIdentifier != uniqueIdentifier {
		return nil, fmt.Errorf("Unexpected get response uniqueIdentifier actual %v, expected %v", getRespPayload.UniqueIdentifier, uniqueIdentifier)
	}
	if getRespPayload.ObjectType != kmip14.ObjectTypeSymmetricKey {
		return nil, fmt.Errorf("Unexpected  get response object type actual %v, expected %v", getRespPayload.ObjectType, kmip14.ObjectTypeSymmetricKey)
	}
	if getRespPayload.SymmetricKey == nil {
		return nil, fmt.Errorf("Unexpected  get response SymmetricKey can not be nil")
	}
	if getRespPayload.SymmetricKey.KeyBlock.KeyFormatType != kmip14.KeyFormatTypeRaw {
		return nil, fmt.Errorf("Unexpected  KeyBlock format type actual %v, expected KeyFormatTypeRaw %v", getRespPayload.SymmetricKey.KeyBlock.KeyFormatType, kmip14.KeyFormatTypeRaw)
	}
	if getRespPayload.SymmetricKey.KeyBlock.CryptographicLength != cryptographicLength {
		return nil, fmt.Errorf("Unexpected  KeyBlock crypto len actual %v, expected %v", getRespPayload.SymmetricKey.KeyBlock.CryptographicLength, cryptographicLength)
	}
	if getRespPayload.SymmetricKey.KeyBlock.CryptographicAlgorithm != kmip14.CryptographicAlgorithmAES {
		return nil, fmt.Errorf("Unexpected  KeyBlock crypto algo actual %v, expected CryptographicAlgorithmAES %v", getRespPayload.SymmetricKey.KeyBlock.CryptographicAlgorithm, kmip14.CryptographicAlgorithmAES)
	}

	secretBytes := getRespPayload.SymmetricKey.KeyBlock.KeyValue.KeyMaterial.([]byte)
	secretBase64 := base64.StdEncoding.EncodeToString(secretBytes)

	// Return the fetched key value
	r := map[string]interface{}{secretID: secretBase64}

	return r, nil
}

// PutSecret will associate an secretId to its secret data
// provided in the arguments and store it into the secret backend
func (k *KMIPSecretStorage) PutSecret(
	secretID string,
	plainText map[string]interface{},
	keyContext map[string]string,
) error {
	log := util.Logger()
	if _, exists := k.secret.StringData[KMIPUniqueID]; exists {
		log.Errorf("KMIPSecretStorage.PutSecret() Key UniqueIdentifier %v was not found in the secret", KMIPUniqueID)
		return secrets.ErrSecretExists
	}

	// Register the key value the KMIP endpoint
	value := plainText[secretID].(string)
	valueBytes, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return err
	}

	conn, err := k.connect()
	if err != nil {
		log.Errorf("KMIPSecretStorage.PutSecret() can not connect %v", err)
		return err
	}
	defer conn.Close()

	registerPayload := kmip.RegisterRequestPayload{
		ObjectType: kmip14.ObjectTypeSymmetricKey,
		SymmetricKey: &kmip.SymmetricKey{
			KeyBlock: kmip.KeyBlock{
				KeyFormatType: kmip14.KeyFormatTypeRaw,
				KeyValue: &kmip.KeyValue{
					KeyMaterial: valueBytes,
				},
				CryptographicLength:    cryptographicLength,
				CryptographicAlgorithm: kmip14.CryptographicAlgorithmAES,
			},
		},
	}
	registerPayload.TemplateAttribute.Append(kmip14.TagCryptographicUsageMask, kmip14.CryptographicUsageMaskExport)
	respMsg, decoder, uniqueBatchItemID, err := k.send(conn, kmip14.OperationRegister, registerPayload)
	if err != nil {
		log.Errorf("KMIPSecretStorage.PutSecret() can send %v", err)
		return err
	}
	bi, err := k.response(respMsg, kmip14.OperationRegister, uniqueBatchItemID)
	if err != nil {
		log.Errorf("KMIPSecretStorage.PutSecret() can verify response %v", err)
		return err
	}

	var registerRespPayload kmip.RegisterResponsePayload
	err = decoder.DecodeValue(&registerRespPayload, bi.ResponsePayload.(ttlv.TTLV))
	if err != nil {
		log.Errorf("KMIPSecretStorage.PutSecret() can decode response payload %v", err)
		return err
	}

	k.secret.StringData[KMIPUniqueID] = registerRespPayload.UniqueIdentifier
	if !util.KubeUpdate(k.secret) {
		log.Errorf("Failed to update KMS secret %v in ns %v", k.secret.Name, k.secret.Namespace)
		return fmt.Errorf("Failed to update KMS secret %v in ns %v", k.secret.Name, k.secret.Namespace)
	}
	return nil
}

// DeleteSecret deletes the secret data associated with the
// supplied secretId.
func (k *KMIPSecretStorage) DeleteSecret(
	secretID string,
	keyContext map[string]string,
) error {
	log := util.Logger()

	// Find the key ID
	uniqueIdentifier, exists := k.secret.StringData[KMIPUniqueID]
	if !exists {
		log.Errorf("KMIPSecretStorage.DeleteSecret() No uniqueIdentifier in the secret")
		return secrets.ErrInvalidSecretId
	}

	conn, err := k.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	respMsg, decoder, uniqueBatchItemID, err := k.send(conn, kmip14.OperationDestroy, kmip.DestroyRequestPayload{
		UniqueIdentifier: uniqueIdentifier,
	})
	if err != nil {
		log.Errorf("KMIPSecretStorage.DeleteSecret() can not send %v", err)
		return err
	}
	bi, err := k.response(respMsg, kmip14.OperationDestroy, uniqueBatchItemID)
	if err != nil {
		log.Errorf("KMIPSecretStorage.DeleteSecret() can verify respnse %v", err)
		return err
	}

	var destroyRespPayload kmip.DestroyResponsePayload
	ttlvPayload, ok := bi.ResponsePayload.(ttlv.TTLV)
	if !ok {
		return fmt.Errorf("failed to parse responsePayload")
	}

	err = decoder.DecodeValue(&destroyRespPayload, ttlvPayload)
	if err != nil {
		log.Errorf("KMIPSecretStorage.DeleteSecret() can decode respnse payload %v", err)
		return err
	}
	if uniqueIdentifier != destroyRespPayload.UniqueIdentifier {
		return fmt.Errorf("Unexpected uniqueIdentifier %v in destroy response , expected %v", destroyRespPayload.UniqueIdentifier, uniqueIdentifier)
	}

	delete(k.secret.Data, KMIPUniqueID)
	delete(k.secret.StringData, KMIPUniqueID)
	if !util.KubeUpdate(k.secret) {
		log.Errorf("Failed to update KMS secret %v in ns %v", k.secret.Name, k.secret.Namespace)
		return fmt.Errorf("Failed to update KMS secret %v in ns %v", k.secret.Name, k.secret.Namespace)
	}

	return nil
}

// ListSecrets is no supported
func (*KMIPSecretStorage) ListSecrets() ([]string, error) {
	return nil, secrets.ErrNotSupported
}

// Encrypt is no supported
func (*KMIPSecretStorage) Encrypt(
	secretID string,
	plaintTextData string,
	keyContext map[string]string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

// Decrypt is no supported
func (*KMIPSecretStorage) Decrypt(
	secretID string,
	encryptedData string,
	keyContext map[string]string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

// Rencrypt is no supported
func (*KMIPSecretStorage) Rencrypt(
	originalSecretID string,
	newSecretID string,
	originalKeyContext map[string]string,
	newKeyContext map[string]string,
	encryptedData string,
) (string, error) {
	return "", secrets.ErrNotSupported
}

// Register KMIP secret storage backend with libopenstorage secrets layer
func init() {
	if err := secrets.Register(KMIPSecretStorageName, NewKMIPSecretStorage); err != nil {
		panic(err.Error())
	}
}
