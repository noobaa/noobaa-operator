package v1alpha1

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note 1: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Note 2: Add custom validation using kubebuilder tags: https://book.kubebuilder.io/reference/generating-crd.html

func init() {
	SchemeBuilder.Register(&BucketClass{}, &BucketClassList{})
}

// BucketClass is the Schema for the bucketclasses API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type BucketClass struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the noobaa BucketClass.
	// +optional
	Spec BucketClassSpec `json:"spec,omitempty"`

	// Most recently observed status of the noobaa BackingStore.
	// +optional
	Status BucketClassStatus `json:"status,omitempty"`
}

// BucketClassList contains a list of BucketClass
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BucketClassList struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of BucketClasses.
	Items []BucketClass `json:"items"`
}

// BucketClassSpec defines the desired state of BucketClass
// +k8s:openapi-gen=true
type BucketClassSpec struct {

	// PlacementPolicy specifies the placement policy for the bucket class
	PlacementPolicy PlacementPolicy `json:"placementPolicy"`
}

// BucketClassStatus defines the observed state of BucketClass
// +k8s:openapi-gen=true
type BucketClassStatus struct {
	// Phase is a simple, high-level summary of where the System is in its lifecycle
	Phase BucketClassPhase `json:"phase"`

	// Conditions is a list of conditions related to operator reconciliation
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +optional
	Conditions []conditionsv1.Condition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`

	// RelatedObjects is a list of objects that are "interesting" or related to this operator.
	RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`
}

// PlacementPolicy specifies the placement policy for the bucket class
type PlacementPolicy struct {

	// Tiers is an ordered list of tiers to use.
	// The model is a waterfall - push to first tier by default,
	// and when no more space spill "cold" storage to next tier.
	Tiers []TierItem `json:"tiers"`
}

// TierItem specifies a storage tier
type TierItem struct {

	// Tier specifies a storage tier
	Tier Tier `json:"tier"`
}

// Tier specifies a storage tier
type Tier struct {

	// Mirrors is an unordered list of mirrors - all mirrors should eventually store all the data of the tier.
	// The mirroring model is async so just a single mirror is required before the write can ack.
	// The first mirror is selected according to locality optimizations of the client endpoint.
	// The data is replicated to the rest of the mirrors in the background.
	Mirrors []MirrorItem `json:"mirrors"`
}

// MirrorItem specifies a mirror item in a tier
type MirrorItem struct {

	// Mirror specifies a mirror group
	Mirror Mirror `json:"mirror"`
}

// Mirror is a single mirror group storage target,
// which can aggregate multiple backing stores in a "spread" model.
// The spread model means that just one of the stores is selected,
// with no particular policy, based on internal optimizations and available space.
type Mirror struct {
	Spread []BackingStoreName `json:"spread"`
}

// BackingStoreName is just a name-reference to a BackingStore
type BackingStoreName = string

// BucketClassPhase is a string enum type for system phases
type BucketClassPhase string

// These are the valid phases:
const (

	// BucketClassPhaseRejected means the spec has been rejected by the operator,
	// this is most likely due to an incompatible configuration.
	// Use describe to see events.
	BucketClassPhaseRejected BucketClassPhase = "Rejected"

	// BucketClassPhaseVerifying means the operator is verifying the spec
	BucketClassPhaseVerifying BucketClassPhase = "Verifying"

	// BucketClassPhaseReady means the noobaa system has been created and ready to serve.
	BucketClassPhaseReady BucketClassPhase = "Ready"

	// BucketClassPhaseDeleting means the operator is deleting the resources on the cluster
	BucketClassPhaseDeleting BucketClassPhase = "Deleting"
)
