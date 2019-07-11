package v1alpha1

import (
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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketClassSpec   `json:"spec,omitempty"`
	Status BucketClassStatus `json:"status,omitempty"`
}

// BucketClassList contains a list of BucketClass
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BucketClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketClass `json:"items"`
}

// BucketClassSpec defines the desired state of BucketClass
// +k8s:openapi-gen=true
type BucketClassSpec struct {
}

// BucketClassStatus defines the observed state of BucketClass
// +k8s:openapi-gen=true
type BucketClassStatus struct {
}
