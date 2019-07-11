package v1alpha1

import (
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
	Type StoreType `json:"type"`
}

type StoreType string

const (
	StoreTypeS3 StoreType = "aws-s3"
)

// BackingStoreStatus defines the observed state of BackingStore
// +k8s:openapi-gen=true
type BackingStoreStatus struct {
}
