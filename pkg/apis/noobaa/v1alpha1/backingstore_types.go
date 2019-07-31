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
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="Type"
// +kubebuilder:printcolumn:name="Bucket-Name",type="string",JSONPath=".spec.bucketName",description="Bucket Name"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
	S3Options *S3Options `json:"s3Options,omitempty"`
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
	Region string `json:"region,omitempty"`
	// Endpoint is the S3 endpoint to use
	// +optional
	Endpoint string `json:"endpoint,omitempty"`
	// SSLDisabled allows to disable SSL and use plain http
	// +optional
	SSLDisabled bool `json:"sslDisabled,omitempty"`
	// S3ForcePathStyle forces the client to send the bucket name in the path
	// aka path-style rather than as a subdomain of the endpoint.
	// +optional
	S3ForcePathStyle bool `json:"s3ForcePathStyle,omitempty"`
	// SignatureVersion specifies the client signature version to use when signing requests.
	// +optional
	SignatureVersion S3SignatureVersion `json:"signatureVersion,omitempty"`
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

	// Phase is a simple, high-level summary of where the System is in its lifecycle
	Phase BackingStorePhase `json:"phase"`

	// Current service state of the noobaa system.
	// Based on: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-conditions
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []SystemCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// BackingStorePhase is a string enum type for system phases
type BackingStorePhase string

// These are the valid phases:
const (

	// BackingStorePhaseRejected means the spec has been rejected by the operator,
	// this is most likely due to an incompatible configuration.
	// Describe the noobaa system to see events.
	BackingStorePhaseRejected BackingStorePhase = "Rejected"

	// BackingStorePhaseVerifying means the operator is verifying the spec
	BackingStorePhaseVerifying BackingStorePhase = "Verifying"

	// BackingStorePhaseCreating means the operator is creating the resources on the cluster
	BackingStorePhaseCreating BackingStorePhase = "Creating"

	// BackingStorePhaseConnecting means the operator is trying to connect to the pods and services it created
	BackingStorePhaseConnecting BackingStorePhase = "Connecting"

	// BackingStorePhaseReady means the noobaa system has been created and ready to serve.
	BackingStorePhaseReady BackingStorePhase = "Ready"
)
