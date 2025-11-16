package v1alpha1

import (
	"os"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Register the noobaa variant of cnpg types
func init() {
	// skip if USE_CNPG_API_GROUP is true
	if os.Getenv("USE_CNPG_API_GROUP") == "true" {
		return
	}
	CNPGSchemeBuilder.Register(&cnpgv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "Cluster",
		},
	}, &cnpgv1.ClusterList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "ClusterList",
		},
	}, &cnpgv1.ImageCatalog{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "ImageCatalog",
		},
	}, &cnpgv1.ImageCatalogList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "ImageCatalogList",
		},
	}, &cnpgv1.ScheduledBackup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "ScheduledBackup",
		},
	}, &cnpgv1.ScheduledBackupList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "ScheduledBackupList",
		},
	}, &cnpgv1.Backup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "Backup",
		},
	}, &cnpgv1.BackupList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "postgresql.cnpg.noobaa.io/v1",
			Kind:       "BackupList",
		},
	})
}
