package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
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
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type BackingStore struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the noobaa BackingStore.
	// +optional
	Spec BackingStoreSpec `json:"spec,omitempty"`

	// Most recently observed status of the noobaa BackingStore.
	// +optional
	Status BackingStoreStatus `json:"status,omitempty"`
}

// BackingStoreList contains a list of BackingStore
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackingStoreList struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of BackingStores.
	Items []BackingStore `json:"items"`
}

// BackingStoreSpec defines the desired state of BackingStore
// +k8s:openapi-gen=true
type BackingStoreSpec struct {

	// Type is an enum of supported types
	Type StoreType `json:"type"`

	// AWSS3Spec specifies a backing store of type aws-s3
	// +optional
	AWSS3 *AWSS3Spec `json:"awsS3,omitempty"`

	// S3Compatible specifies a backing store of type s3-compatible
	// +optional
	S3Compatible *S3CompatibleSpec `json:"s3Compatible,omitempty"`

	// AzureBlob specifies a backing store of type azure-blob
	// +optional
	AzureBlob *AzureBlobSpec `json:"azureBlob,omitempty"`

	// GoogleCloudStorage specifies a backing store of type google-cloud-storage
	// +optional
	GoogleCloudStorage *GoogleCloudStorageSpec `json:"googleCloudStorage,omitempty"`

	// PVPool specifies a backing store of type pv-pool
	// +optional
	PVPool *PVPoolSpec `json:"pvPool,omitempty"`
}

// BackingStoreStatus defines the observed state of BackingStore
// +k8s:openapi-gen=true
type BackingStoreStatus struct {

	// Phase is a simple, high-level summary of where the backing store is in its lifecycle
	// +optional
	Phase BackingStorePhase `json:"phase,omitempty"`

	// Conditions is a list of conditions related to operator reconciliation
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +optional
	Conditions []conditionsv1.Condition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`

	// RelatedObjects is a list of objects related to this operator.
	// +optional
	RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`
}

// StoreType is the backing store type enum
type StoreType string

const (
	// StoreTypeAWSS3 is used to connect to AWS S3
	StoreTypeAWSS3 StoreType = "aws-s3"

	// StoreTypeS3Compatible is used to connect to S3 compatible storage
	StoreTypeS3Compatible StoreType = "s3-compatible"

	// StoreTypeGoogleCloudStorage is used to connect to Google Cloud Storage
	StoreTypeGoogleCloudStorage StoreType = "google-cloud-storage"

	// StoreTypeAzureBlob is used to connect to Azure Blob
	StoreTypeAzureBlob StoreType = "azure-blob"

	// StoreTypePVPool is used to allocate storage by dynamically allocating PVs (using PVCs)
	StoreTypePVPool StoreType = "pv-pool"
)

// AWSS3Spec specifies a backing store of type aws-s3
type AWSS3Spec struct {

	// TargetBucket is the name of the target S3 bucket
	TargetBucket string `json:"targetBucket"`

	// Secret refers to a secret that provides the credentials
	// The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	Secret corev1.SecretReference `json:"secret"`

	// Region is the AWS region
	// +optional
	Region string `json:"region,omitempty"`

	// SSLDisabled allows to disable SSL and use plain http
	// +optional
	SSLDisabled bool `json:"sslDisabled,omitempty"`
}

// S3CompatibleSpec specifies a backing store of type s3-compatible
type S3CompatibleSpec struct {

	// TargetBucket is the name of the target S3 bucket
	TargetBucket string `json:"targetBucket"`

	// Secret refers to a secret that provides the credentials
	// The secret should define AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	Secret corev1.SecretReference `json:"secret"`

	// Endpoint is the S3 compatible endpoint: http(s)://host:port
	Endpoint string `json:"endpoint"`

	// SignatureVersion specifies the client signature version to use when signing requests.
	// +optional
	SignatureVersion S3SignatureVersion `json:"signatureVersion,omitempty"`
}

// AzureBlobSpec specifies a backing store of type azure-blob
type AzureBlobSpec struct {

	// TargetBlobContainer is the name of the target Azure Blob container
	TargetBlobContainer string `json:"targetBlobContainer"`

	// Secret refers to a secret that provides the credentials
	// The secret should define AccountName and AccountKey as provided by Azure Blob.
	Secret corev1.SecretReference `json:"secret"`
}

// GoogleCloudStorageSpec specifies a backing store of type google-cloud-storage
type GoogleCloudStorageSpec struct {

	// TargetBucket is the name of the target S3 bucket
	TargetBucket string `json:"targetBucket"`

	// Secret refers to a secret that provides the credentials
	// The secret should define GoogleServiceAccountPrivateKeyJson containing the entire json string as provided by Google.
	Secret corev1.SecretReference `json:"secret"`
}

// PVPoolSpec specifies a backing store of type pv-pool
type PVPoolSpec struct {

	// StorageClass is the name of the storage class to use for the PV's
	StorageClass string `json:"storageClass,omitempty"`

	// NumVolumes is the number of volumes to allocate
	NumVolumes int `json:"numVolumes"`

	// VolumeResources represents the minimum resources each volume should have.
	VolumeResources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// S3SignatureVersion specifies the client signature version to use when signing requests.
type S3SignatureVersion string

const (
	// S3SignatureVersionV4 is aws v4
	S3SignatureVersionV4 S3SignatureVersion = "v4"
	// S3SignatureVersionV2 is aws v2
	S3SignatureVersionV2 S3SignatureVersion = "v2"
)

// BackingStorePhase is a string enum type for backing store reconcile phases
type BackingStorePhase string

// These are the valid phases:
const (

	// BackingStorePhaseRejected means the spec has been rejected by the operator,
	// this is most likely due to an incompatible configuration.
	// Use describe to see events.
	BackingStorePhaseRejected BackingStorePhase = "Rejected"

	// BackingStorePhaseVerifying means the operator is verifying the spec
	BackingStorePhaseVerifying BackingStorePhase = "Verifying"

	// BackingStorePhaseConnecting means the operator is trying to connect to the system
	BackingStorePhaseConnecting BackingStorePhase = "Connecting"

	// BackingStorePhaseCreating means the operator is creating the resources on the cluster
	BackingStorePhaseCreating BackingStorePhase = "Creating"

	// BackingStorePhaseReady means the noobaa system has been created and ready to serve.
	BackingStorePhaseReady BackingStorePhase = "Ready"

	// BackingStorePhaseDeleting means the operator is deleting the resources on the cluster
	BackingStorePhaseDeleting BackingStorePhase = "Deleting"
)
