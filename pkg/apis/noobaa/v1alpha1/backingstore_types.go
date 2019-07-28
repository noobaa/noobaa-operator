package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note 1: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Note 2: Add custom validation using kubebuilder tags: https://book.kubebuilder.io/reference/generating-crd.html

func init() {
	SchemeBuilder.Register(&BackingStore{}, &BackingStoreList{})
}

// BackingStore is the Schema for the backingstores API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type BackingStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackingStoreSpec   `json:"spec,omitempty"`
	Status BackingStoreStatus `json:"status,omitempty"`
}

// BackingStoreList contains a list of BackingStore
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackingStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackingStore `json:"items"`
}

// BackingStoreSpec defines the desired state of BackingStore
// +k8s:openapi-gen=true
type BackingStoreSpec struct {

	// Type
	Type StoreType `json:"type"`

	BucketName string `json:"bucketName"`

	// Secret refers to a secret that provides the credentials
	Secret corev1.SecretReference `json:"secret"`

	// S3Options specifies client options for the backing store
	// +optional
	S3Options *S3Options `json:"s3Options"`
}

// StoreType is the backing store type enum
type StoreType string

const (
	// StoreTypeAWSS3 is used to connect to AWS S3
	StoreTypeAWSS3 StoreType = "aws-s3"
	// StoreTypeGoogleCloudStorage is used to connect to Google Cloud Storage
	StoreTypeGoogleCloudStorage StoreType = "google-cloud-storage"
	// StoreTypeAzureBlob is used to connect to Azure Blob
	StoreTypeAzureBlob StoreType = "azure-blob"
	// StoreTypeS3Compatible is used to connect to S3 compatible storage
	StoreTypeS3Compatible StoreType = "s3-compatible"
)

// S3Options specifies client options for the backing store
type S3Options struct {
	// Region is the AWS region
	// +optional
	Region string `json:"region"`
	// Endpoint is the S3 endpoint to use
	// +optional
	Endpoint string `json:"endpoint"`
	// SSLDisabled allows to disable SSL and use plain http
	// +optional
	SSLDisabled bool `json:"sslDisabled"`
	// S3ForcePathStyle forces the client to send the bucket name in the path
	// aka path-style rather than as a subdomain of the endpoint.
	// +optional
	S3ForcePathStyle bool `json:"s3ForcePathStyle"`
	// SignatureVersion specifies the client signature version to use when signing requests.
	// +optional
	SignatureVersion S3SignatureVersion `json:"signatureVersion"`
}

// S3SignatureVersion specifies the client signature version to use when signing requests.
type S3SignatureVersion string

const (
	// S3SignatureVersionV4 is aws v4
	S3SignatureVersionV4 S3SignatureVersion = "v4"
	// S3SignatureVersionV2 is aws v2
	S3SignatureVersionV2 S3SignatureVersion = "v2"
)

// BackingStoreStatus defines the observed state of BackingStore
// +k8s:openapi-gen=true
type BackingStoreStatus struct {
}
