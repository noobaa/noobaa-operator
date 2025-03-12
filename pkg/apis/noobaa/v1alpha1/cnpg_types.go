package v1alpha1

import (
	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Register the noobaa variant of cnpg types
func init() {
	CNPGSchemeBuilder.Register(&cnpgv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "Cluster",
		},
	}, &cnpgv1.ImageCatalog{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "ImageCatalog",
		},
	})
}
