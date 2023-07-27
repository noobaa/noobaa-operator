package v1alpha1

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note 1: Run "make gen-api" to regenerate code after modifying this file
// Note 2: Add custom validation using kubebuilder tags: https://book.kubebuilder.io/reference/generating-crd.html

func init() {
	SchemeBuilder.Register(&BucketClass{}, &BucketClassList{})
}

// BucketClass is the Schema for the bucketclasses API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Placement",type="string",JSONPath=".spec.placementPolicy",description="Placement"
// +kubebuilder:printcolumn:name="NamespacePolicy",type="string",JSONPath=".spec.namespacePolicy",description="NamespacePolicy"
// +kubebuilder:printcolumn:name="Quota",type="string",JSONPath=".spec.quota",description="Quota"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type BucketClass struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

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
	// +optional
	PlacementPolicy *PlacementPolicy `json:"placementPolicy,omitempty"`

	// NamespacePolicy specifies the namespace policy for the bucket class
	// +optional
	NamespacePolicy *NamespacePolicy `json:"namespacePolicy,omitempty"`

	// Quota specifies the quota configuration for the bucket class
	// +optional
	Quota *Quota `json:"quota,omitempty"`

	// ReplicationPolicy specifies a json of replication rules for the bucketclass
	// +optional
	ReplicationPolicy string `json:"replicationPolicy,omitempty"`
}

// BucketClassStatus defines the observed state of BucketClass
// +k8s:openapi-gen=true
type BucketClassStatus struct {
	// Phase is a simple, high-level summary of where the System is in its lifecycle
	// +optional
	Phase BucketClassPhase `json:"phase,omitempty"`

	// Conditions is a list of conditions related to operator reconciliation
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +optional
	Conditions []conditionsv1.Condition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`

	// RelatedObjects is a list of objects related to this operator.
	// +optional
	RelatedObjects []corev1.ObjectReference `json:"relatedObjects,omitempty"`
	// Mode is a simple, high-level summary of where the System is in its lifecycle
	// +optional
	Mode string `json:"mode,omitempty"`
}

// PlacementPolicy specifies the placement policy for the bucket class
type PlacementPolicy struct {

	// Tiers is an ordered list of tiers to use.
	// The model is a waterfall - push to first tier by default,
	// and when no more space spill "cold" storage to next tier.
	Tiers []Tier `json:"tiers,omitempty"`
}

// NamespacePolicy specifies the namespace policy for the bucket class
type NamespacePolicy struct {
	// Type is the namespace policy type
	Type NSBucketClassType `json:"type,omitempty"`

	// Single is a namespace policy configuration of type Single
	// +optional
	Single *SingleNamespacePolicy `json:"single,omitempty"`

	// Multi is a namespace policy configuration of type Multi
	// +optional
	Multi *MultiNamespacePolicy `json:"multi,omitempty"`

	// Cache is a namespace policy configuration of type Cache
	// +optional
	Cache *CacheNamespacePolicy `json:"cache,omitempty"`
}

// SingleNamespacePolicy specifies the configuration of namespace policy of type Single
type SingleNamespacePolicy struct {

	// Resource is the read and write resource name to use
	Resource string `json:"resource,omitempty"`
}

// MultiNamespacePolicy specifies the configuration of namespace policy of type Multi
type MultiNamespacePolicy struct {

	// ReadResources is an ordered list of read resources names to use
	ReadResources []string `json:"readResources,omitempty"`

	// WriteResource is the write resource name to use
	// +optional
	WriteResource string `json:"writeResource,omitempty"`
}

// CacheNamespacePolicy specifies the configuration of namespace policy of type Cache
type CacheNamespacePolicy struct {

	// HubResource is the read and write resource name to use
	HubResource string `json:"hubResource,omitempty"`

	// Caching is the cache specification for the ns policy
	Caching *CacheSpec `json:"caching,omitempty"`
}

// Quota bucket config
type Quota struct {

	//limits the max total size of objects per bucket
	MaxSize string `json:"maxSize,omitempty"`

	//limits the max total quantity of objects per bucket
	MaxObjects string `json:"maxObjects,omitempty"`
}

// CacheSpec specifies the cache specifications for the bucket class
type CacheSpec struct {

	// TTL specifies the cache ttl
	TTL int `json:"ttl,omitempty"`

	// Prefix is prefix of the future cached data
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// Tier specifies a storage tier
type Tier struct {

	// Placement specifies the type of placement for the tier
	// If empty it should have a single backing store.
	// +optional
	// +kubebuilder:validation:Enum=Spread;Mirror
	Placement TierPlacement `json:"placement,omitempty"`

	// BackingStores is an unordered list of backing store names.
	// The meaning of the list depends on the placement.
	// +optional
	BackingStores []BackingStoreName `json:"backingStores,omitempty"`
}

// TierPlacement is a string enum type for tier placement
type TierPlacement string

// These are the valid placement values:
const (

	// TierPlacementSingle stores the data on a single backing store.
	TierPlacementSingle TierPlacement = ""

	// TierPlacementMirror requires 2 or more backing store.
	// All mirrors should eventually store all the data of the tier.
	// The mirroring model is async so just a single mirror is required before the write can ack.
	// The first mirror is selected according to locality optimizations of the client endpoint.
	// The data is replicated to the rest of the mirrors in the background.
	TierPlacementMirror TierPlacement = "Mirror"

	// TierPlacementSpread requires 2 or more backing store.
	// The data is spread over the backing stores without any specific preference.
	// The spread is a simple aggregate of those backing stores capacity.
	TierPlacementSpread TierPlacement = "Spread"
)

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

	// BucketClassPhaseConfiguring means the operator is configuring the buckets as requested
	BucketClassPhaseConfiguring BucketClassPhase = "Configuring"

	// BucketClassPhaseReady means the noobaa system has been created and ready to serve.
	BucketClassPhaseReady BucketClassPhase = "Ready"

	// BucketClassPhaseDeleting means the operator is deleting the resources on the cluster
	BucketClassPhaseDeleting BucketClassPhase = "Deleting"
)

// NSBucketClassType is the namespace bucketclass type enum
type NSBucketClassType string

const (
	// NSBucketClassTypeSingle is used to configure namespace bucket class of type Single
	NSBucketClassTypeSingle NSBucketClassType = "Single"

	// NSBucketClassTypeMulti is used to configure namespace bucket class of type Multi
	NSBucketClassTypeMulti NSBucketClassType = "Multi"

	// NSBucketClassTypeCache is used to configure namespace bucket class of type Cache
	NSBucketClassTypeCache NSBucketClassType = "Cache"
)
