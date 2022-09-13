package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
)

// Note 1: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Note 2: Add custom validation using kubebuilder tags: https://book.kubebuilder.io/reference/generating-crd.html

func init() {
	SchemeBuilder.Register(&NooBaaAccount{}, &NooBaaAccountList{})
}

// NooBaaAccount is the Schema for the NooBaaAccounts API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type NooBaaAccount struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the NooBaaAccount.
	// +optional
	Spec NooBaaAccountSpec `json:"spec,omitempty"`

	// Most recently observed status of the NooBaaAccount.
	// +optional
	Status NooBaaAccountStatus `json:"status,omitempty"`
}

// NooBaaAccountList contains a list of NooBaaAccount
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NooBaaAccountList struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of NooBaaAccounts.
	Items []NooBaaAccount `json:"items"`
}

// NooBaaAccountSpec defines the desired state of NooBaaAccount
// +k8s:openapi-gen=true
type NooBaaAccountSpec struct {
	// AllowBucketCreate specifies if new buckets can be created by this account
	AllowBucketCreate bool `json:"allow_bucket_creation"`

	// NsfsAccountConfig specifies the configurations on Namespace FS
	// +nullable
	// +optional
	NsfsAccountConfig *AccountNsfsConfig `json:"nsfs_account_config,omitempty"`

	// DefaultResource specifies which backingstore this account will use to create new buckets
	// +optional
	DefaultResource string `json:"default_resource,omitempty"`
}

// AccountNsfsConfig is the configuration of NSFS of CreateAccountParams
type AccountNsfsConfig struct {
	UID            int    `json:"uid"`
	GID            int    `json:"gid"`
	NewBucketsPath string `json:"new_buckets_path"`
	NsfsOnly       bool   `json:"nsfs_only"`
}

// NooBaaAccountStatus defines the observed state of NooBaaAccount
// +k8s:openapi-gen=true
type NooBaaAccountStatus struct {

	// Phase is a simple, high-level summary of where the noobaa user is in its lifecycle
	// +optional
	Phase NooBaaAccountPhase `json:"phase,omitempty"`

	// Conditions is a list of conditions related to operator reconciliation
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +optional
	Conditions []conditionsv1.Condition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`

	// RelatedObjects is a list of objects related to this operator.
	// +optional
	RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`
}

// NooBaaAccountPhase is a string enum type for backing store reconcile phases
type NooBaaAccountPhase string

// These are the valid phases:
const (

	// NooBaaAccountPhaseRejected means the spec has been rejected by the operator,
	// this is most likely due to an incompatible configuration.
	// Use describe to see events.
	NooBaaAccountPhaseRejected NooBaaAccountPhase = "Rejected"

	// NooBaaAccountPhaseVerifying means the operator is verifying the spec
	NooBaaAccountPhaseVerifying NooBaaAccountPhase = "Verifying"

	// NooBaaAccountPhaseConfiguring means the operator is trying to connect to the system
	NooBaaAccountPhaseConnecting NooBaaAccountPhase = "Connecting"

	// NooBaaAccountPhaseConfiguring means the operator is configuring the account as requested
	NooBaaAccountPhaseConfiguring NooBaaAccountPhase = "Configuring"

	// NooBaaAccountPhaseReady means the noobaa user has been created and ready to serve.
	NooBaaAccountPhaseReady NooBaaAccountPhase = "Ready"

	// NooBaaAccountPhaseDeleting means the operator is deleting the resources on the cluster
	NooBaaAccountPhaseDeleting NooBaaAccountPhase = "Deleting"
)
