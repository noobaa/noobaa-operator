package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
)

// Note 1: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Note 2: Add custom validation using kubebuilder tags: https://book.kubebuilder.io/reference/generating-crd.html

func init() {
	SchemeBuilder.Register(&NamespaceStore{}, &NamespaceStoreList{})
}

// NamespaceStore is the Schema for the namespacestores API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="Type"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type NamespaceStore struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the noobaa NamespaceStore.
	// +optional
	Spec NamespaceStoreSpec `json:"spec,omitempty"`

	// Most recently observed status of the noobaa NamespaceStore.
	// +optional
	Status NamespaceStoreStatus `json:"status,omitempty"`
}

// NamespaceStoreList contains a list of NamespaceStore
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NamespaceStoreList struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of NamespaceStores.
	Items []NamespaceStore `json:"items"`
}

// NamespaceStoreSpec defines the desired state of NamespaceStore
// +k8s:openapi-gen=true
type NamespaceStoreSpec struct {

	// Type is an enum of supported types
	Type NSType `json:"type"`

	//AccessMode is an enum of supported access modes
	// +optional
	AccessMode AccessModeType `json:"accessMode,omitempty"`

	// AWSS3Spec specifies a namespace store of type aws-s3
	// +optional
	AWSS3 *AWSS3Spec `json:"awsS3,omitempty"`

	// S3Compatible specifies a namespace store of type s3-compatible
	// +optional
	S3Compatible *S3CompatibleSpec `json:"s3Compatible,omitempty"`

	// IBMCos specifies a namespace store of type ibm-cos
	// +optional
	IBMCos *IBMCosSpec `json:"ibmCos,omitempty"`

	// AzureBlob specifies a namespace store of type azure-blob
	// +optional
	AzureBlob *AzureBlobSpec `json:"azureBlob,omitempty"`

	// GoogleCloudStorage specifies a namespace store of type google-cloud-storage
	// +optional
	GoogleCloudStorage *GoogleCloudStorageSpec `json:"googleCloudStorage,omitempty"`

	// NSFS specifies a namespace store of type nsfs
	// +optional
	NSFS *NSFSSpec `json:"nsfs,omitempty"`
}

// NamespaceStoreStatus defines the observed state of NamespaceStore
// +k8s:openapi-gen=true
type NamespaceStoreStatus struct {

	// Phase is a simple, high-level summary of where the namespace store is in its lifecycle
	// +optional
	Phase NamespaceStorePhase `json:"phase,omitempty"`

	// Conditions is a list of conditions related to operator reconciliation
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +optional
	Conditions []conditionsv1.Condition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`

	// RelatedObjects is a list of objects related to this operator.
	// +optional
	RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`
	// Mode specifies the updating mode of a NamespaceStore
	// +optional
	Mode NamespaceStoreMode `json:"mode,omitempty"`
}

// NamespaceStoreMode defines the updated Mode of NamespaceStore
type NamespaceStoreMode struct {
	// ModeCode specifies the updated mode of namespacestore
	// +optional
	ModeCode string `json:"modeCode,omitempty"`
	// TimeStamp specifies the update time of namespacestore new mode
	// +optional
	TimeStamp string `json:"timeStamp,omitempty"`
}

// NamespaceStorePhase is a string enum type for namespace store reconcile phases
type NamespaceStorePhase string

// These are the valid phases:
const (

	// NamespaceStorePhasePhaseRejected means the spec has been rejected by the operator,
	// this is most likely due to an incompatible configuration.
	// Use describe to see events.
	NamespaceStorePhaseRejected NamespaceStorePhase = "Rejected"

	// NamespacetorePhaseVerifying means the operator is verifying the spec
	NamespaceStorePhaseVerifying NamespaceStorePhase = "Verifying"

	// NamespaceStorePhaseConnecting means the operator is trying to connect to the system
	NamespaceStorePhaseConnecting NamespaceStorePhase = "Connecting"

	// NamespaceStorePhaseCreating means the operator is creating the resources on the cluster
	NamespaceStorePhaseCreating NamespaceStorePhase = "Creating"

	// NamespaceStorePhaseReady means the noobaa system has been created and ready to serve.
	NamespaceStorePhaseReady NamespaceStorePhase = "Ready"

	// NamespaceStorePhaseDeleting means the operator is deleting the resources on the cluster
	NamespaceStorePhaseDeleting NamespaceStorePhase = "Deleting"
)

// NSType is the backing store type enum
type NSType string

const (
	// NSStoreTypeAWSS3 is used to connect to AWS S3
	NSStoreTypeAWSS3 NSType = "aws-s3"

	// NSStoreTypeS3Compatible is used to connect to S3 compatible storage
	NSStoreTypeS3Compatible NSType = "s3-compatible"

	// NSStoreTypeIBMCos is used to connect to IBM cos storage
	NSStoreTypeIBMCos NSType = "ibm-cos"

	// NSStoreTypeAzureBlob is used to connect to Azure Blob
	NSStoreTypeAzureBlob NSType = "azure-blob"

	// NSStoreTypeGoogleCloudStorage is used to connect to Google Cloud Storage
	NSStoreTypeGoogleCloudStorage NSType = "google-cloud-storage"

	// NSStoreTypeNSFS is used to connect to a file system
	NSStoreTypeNSFS NSType = "nsfs"
)

// AccessModeType is the type of all the optional access modes
type AccessModeType string

const (
	// AccessModeReadWrite is the default access mode
	AccessModeReadWrite AccessModeType = "ReadWrite"
	// AccessModeReadOnly is a read only access mode
	AccessModeReadOnly AccessModeType = "ReadOnly"
)

// NSFSSpec specifies a namespace store of type nsfs
type NSFSSpec struct {
	// PvcName is the name of the pvc in which the file system resides
	PvcName string `json:"pvcName"`

	// SubPath is a path to a sub directory in the pvc file system
	// +optional
	SubPath string `json:"subPath"`

	// FsBackend is the backend type of the file system
	// +optional
	// +kubebuilder:validation:Enum=CEPH_FS;GPFS;NFSv4
	FsBackend string `json:"fsBackend,omitempty"`
}
