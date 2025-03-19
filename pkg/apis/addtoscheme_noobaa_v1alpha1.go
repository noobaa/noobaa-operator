package apis

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	// Register the noobaa variant of cnpg types
	AddToSchemes = append(AddToSchemes, v1alpha1.CNPGSchemeBuilder.AddToScheme)
}
