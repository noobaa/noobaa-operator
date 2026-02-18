// NOTE: Boilerplate only.  Ignore this file.
// Package v1alpha1 contains API Schema definitions for the noobaa v1alpha1 API group
// +groupName=noobaa.io

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "noobaa.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// GroupVersion is group version used to register these objects
	CNPGGroupVersion = schema.GroupVersion{Group: "postgresql.cnpg.noobaa.io", Version: "v1"}

	// CNPGSchemeBuilder is used to add go types to the GroupVersionKind scheme
	CNPGSchemeBuilder = &scheme.Builder{GroupVersion: CNPGGroupVersion}
)
