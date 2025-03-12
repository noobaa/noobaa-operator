package apis

import (
	"os"

	"github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)

	// skip if USE_CNPG_API_GROUP is true
	if os.Getenv("USE_CNPG_API_GROUP") != "true" {
		// Register the noobaa variant of cnpg types
		AddToSchemes = append(AddToSchemes, v1alpha1.CNPGSchemeBuilder.AddToScheme)
	}
}
