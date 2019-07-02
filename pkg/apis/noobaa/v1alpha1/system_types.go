package v1alpha1

// Note 1: Run "operator-sdk generate k8s" to regenerate code after modifying this file
// Note 2: Add custom validation using kubebuilder tags: https://book.kubebuilder.io/reference/generating-crd.html

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&System{}, &SystemList{})
}

// System is the Schema for the systems API
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=noobaa;nb
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Mgmt-Endpoints",type="string",JSONPath=".status.services.serviceMgmt.nodePorts",description="Mgmt Endpoints"
// +kubebuilder:printcolumn:name="S3-Endpoints",type="string",JSONPath=".status.services.serviceS3.nodePorts",description="S3 Endpoints"
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".status.actualImage",description="Actual Image"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type System struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the System.
	// +optional
	Spec SystemSpec `json:"spec,omitempty"`

	// Most recently observed status of the System.
	// +optional
	Status SystemStatus `json:"status,omitempty"`
}

// SystemList contains a list of System
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SystemList struct {

	// Standard type metadata.
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Systems.
	Items []System `json:"items"`
}

// SystemSpec defines the desired state of System
// +k8s:openapi-gen=true
type SystemSpec struct {

	// Image (optional) overrides the default image for server pods
	// +optional
	Image string `json:"image,omitempty"`
}

// SystemStatus defines the observed state of System
// +k8s:openapi-gen=true
type SystemStatus struct {

	// ObservedGeneration is the most recent generation observed for this System.
	// It corresponds to the System's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration"`

	// Phase is a simple, high-level summary of where the System is in its lifecycle
	Phase SystemPhase `json:"phase"`

	// ActualImage is set to report which image the operator is using
	ActualImage string `json:"actualImage"`

	Accounts SystemAccountsStatus `json:"accounts"`

	Services SystemServicesStatus `json:"services"`

	// Readme is a user readable string with explanations on the system
	Readme string `json:"readme"`
}

// SystemPhase is a string enum type for system phases
type SystemPhase string

// These are the valid phases of systems:
const (

	// SystemPhaseRejected means the system spec has been rejected by the operator,
	// this is most likely due to an incompatible configuration.
	// Describe the system to see events.
	SystemPhaseRejected SystemPhase = "Rejected"

	// SystemPhaseVerifying means the operator is verifying the system spec
	SystemPhaseVerifying SystemPhase = "Verifying"

	// SystemPhaseCreating means the operator is creating the system resources on the cluster
	SystemPhaseCreating SystemPhase = "Creating"

	// SystemPhaseConfiguring means the operator is configuring the system as requested
	SystemPhaseConfiguring SystemPhase = "Configuring"

	// SystemPhaseReady means the system has been created and ready to serve.
	SystemPhaseReady SystemPhase = "Ready"
)

// SystemAccountsStatus is the status info of admin account
type SystemAccountsStatus struct {
	Admin SystemUserStatus `json:"admin"`
}

// SystemServicesStatus is the status info of the system's services
type SystemServicesStatus struct {
	ServiceMgmt SystemServiceStatus `json:"serviceMgmt"`
	ServiceS3   SystemServiceStatus `json:"serviceS3"`
}

// SystemUserStatus is the status info of a user secret
type SystemUserStatus struct {
	SecretRef corev1.SecretReference `json:"secretRef"`
}

// SystemServiceStatus is the status info and network addresses of a service
type SystemServiceStatus struct {

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
