package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Note 1: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Note 2: Add custom validation using kubebuilder tags: https://book.kubebuilder.io/reference/generating-crd.html

func init() {
	SchemeBuilder.Register(&NooBaa{}, &NooBaaList{})
}

// NooBaa is the Schema for the NooBaas API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=nb
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Mgmt-Endpoints",type="string",JSONPath=".status.services.serviceMgmt.nodePorts",description="Mgmt Endpoints"
// +kubebuilder:printcolumn:name="S3-Endpoints",type="string",JSONPath=".status.services.serviceS3.nodePorts",description="S3 Endpoints"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.actualImage",description="Actual Image"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type NooBaa struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the noobaa system.
	// +optional
	Spec NooBaaSpec `json:"spec,omitempty"`

	// Most recently observed status of the noobaa system.
	// +optional
	Status NooBaaStatus `json:"status,omitempty"`
}

// NooBaaList contains a list of noobaa systems
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NooBaaList struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Systems.
	Items []NooBaa `json:"items"`
}

// NooBaaSpec defines the desired state of System
// +k8s:openapi-gen=true
type NooBaaSpec struct {

	// Image (optional) overrides the default image for server container
	// +optional
	Image *string `json:"image,omitempty"`

	// MongoImage (optional) overrides the default image for mongodb container
	// +optional
	MongoImage *string `json:"mongoImage,omitempty"`

	// ImagePullSecret (optional) sets a pull secret for the system image
	// +optional
	ImagePullSecret *corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`

	// StorageClassName (optional) overrides the default StorageClass
	// for the PVC that the operator creates, this affects where the
	// system stores its database which contains system config,
	// buckets, objects meta-data and mapping file parts to storage locations.
	// +optional
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// NooBaaStatus defines the observed state of System
// +k8s:openapi-gen=true
type NooBaaStatus struct {

	// ObservedGeneration is the most recent generation observed for this noobaa system.
	// It corresponds to the CR generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration"`

	// Phase is a simple, high-level summary of where the System is in its lifecycle
	Phase SystemPhase `json:"phase"`

	// Current service state of the noobaa system.
	// Based on: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-conditions
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []SystemCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// ActualImage is set to report which image the operator is using
	ActualImage string `json:"actualImage"`

	Accounts AccountsStatus `json:"accounts"`

	Services ServicesStatus `json:"services"`

	// Readme is a user readable string with explanations on the system
	Readme string `json:"readme"`
}

// SystemPhase is a string enum type for system phases
type SystemPhase string

// These are the valid phases:
const (

	// SystemPhaseRejected means the spec has been rejected by the operator,
	// this is most likely due to an incompatible configuration.
	// Describe the noobaa system to see events.
	SystemPhaseRejected SystemPhase = "Rejected"

	// SystemPhaseVerifying means the operator is verifying the spec
	SystemPhaseVerifying SystemPhase = "Verifying"

	// SystemPhaseCreating means the operator is creating the resources on the cluster
	SystemPhaseCreating SystemPhase = "Creating"

	// SystemPhaseWaitingToConnect means the operator is waiting to connect to the pods and services it created
	SystemPhaseWaitingToConnect SystemPhase = "WaitingToConnect"

	// SystemPhaseConfiguring means the operator is configuring the as requested
	SystemPhaseConfiguring SystemPhase = "Configuring"

	// SystemPhaseReady means the noobaa system has been created and ready to serve.
	SystemPhaseReady SystemPhase = "Ready"
)

// SystemCondition contains details for the current condition of this system.
type SystemCondition struct {
	// Type is the type of the condition.
	Type ConditionType `json:"type"`
	// Status is the status of the condition.
	Status ConditionStatus `json:"status"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// ConditionType is a simple string type.
// Types should be used from the enum below.
type ConditionType string

// These are the valid conditions types and statuses:
const (
	ConditionTypePhase ConditionType = "Phase"
)

// ConditionStatus is a simple string type.
// In addition to the generic True/False/Unknown it also can accept SystemPhase enums
type ConditionStatus string

// These are general valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// AccountsStatus is the status info of admin account
type AccountsStatus struct {
	Admin UserStatus `json:"admin"`
}

// ServicesStatus is the status info of the system's services
type ServicesStatus struct {
	ServiceMgmt ServiceStatus `json:"serviceMgmt"`
	ServiceS3   ServiceStatus `json:"serviceS3"`
}

// UserStatus is the status info of a user secret
type UserStatus struct {
	SecretRef corev1.SecretReference `json:"secretRef"`
}

// ServiceStatus is the status info and network addresses of a service
type ServiceStatus struct {

	// NodePorts are the most basic network available
	// it uses the networks available on the hosts of kubernetes nodes.
	// This generally works from within a pod, and from the internal
	// network of the nodes, but may fail from public network.
	// https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
	// +optional
	NodePorts []string `json:"nodePorts,omitempty"`

	// PodPorts are the second most basic network address
	// every pod has an IP in the cluster and the pods network is a mesh
	// so the operator running inside a pod in the cluster can use this address.
	// Note: pod IPs are not guaranteed to persist over restarts, so should be rediscovered.
	// Note2: when running the operator outside of the cluster, pod IP is not accessible.
	// +optional
	PodPorts []string `json:"podPorts,omitempty"`

	// InternalIP are internal addresses of the service inside the cluster
	// https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	// +optional
	InternalIP []string `json:"internalIP,omitempty"`

	// InternalDNS are internal addresses of the service inside the cluster
	// +optional
	InternalDNS []string `json:"internalDNS,omitempty"`

	// ExternalIP are external public addresses for the service
	// LoadBalancerPorts such as AWS ELB provide public address and load balancing for the service
	// IngressPorts are manually created public addresses for the service
	// https://kubernetes.io/docs/concepts/services-networking/service/#external-ips
	// https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
	// https://kubernetes.io/docs/concepts/services-networking/ingress/
	// +optional
	ExternalIP []string `json:"externalIP,omitempty"`

	// ExternalDNS are external public addresses for the service
	// +optional
	ExternalDNS []string `json:"externalDNS,omitempty"`
}
