package system

import (
	"testing"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestFormatBytesKB(t *testing.T) {
	tests := []struct {
		kb   int64
		want string
	}{
		{1024 * 1024, "1GB"},
		{4 * 1024 * 1024, "4GB"},
		{256 * 1024, "256MB"},
		{512 * 1024, "512MB"},
		{16 * 1024, "16MB"},
		{4096, "4MB"},
		{1966, "1966kB"},
		{1024, "1MB"},
	}
	for _, tt := range tests {
		got := formatBytesKB(tt.kb)
		if got != tt.want {
			t.Errorf("formatBytesKB(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

func TestCalculateWalKeepSize(t *testing.T) {
	tests := []struct {
		name     string
		volSize  string
		expected string
	}{
		{"50Gi default", "50Gi", "6GB"},
		{"100Gi", "100Gi", "12GB"},
		{"200Gi hits 16GB cap", "200Gi", "16GB"},
		{"500Gi hits 16GB cap", "500Gi", "16GB"},
		{"10Gi small", "10Gi", "1228MB"},
		{"5Gi hits 1GB min", "5Gi", "1GB"},
		{"invalid falls back", "invalid", "6GB"},
		{"empty falls back", "", "6GB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateWalKeepSize(tt.volSize)
			if got != tt.expected {
				t.Errorf("calculateWalKeepSize(%q) = %q, want %q", tt.volSize, got, tt.expected)
			}
		})
	}
}

// TestDBClusterStatusAfterSpecUpdate covers the regression behind DFBUGS-8895: a cluster
// spec update must never overwrite a status that reports an in-flight initial deployment,
// otherwise the reported phase no longer reflects what the cluster is actually doing.
func TestDBClusterStatusAfterSpecUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current nbv1.DBClusterStatus
		want    nbv1.DBClusterStatus
	}{
		{"ready moves to updating", nbv1.DBClusterStatusReady, nbv1.DBClusterStatusUpdating},
		{"creating is preserved", nbv1.DBClusterStatusCreating, nbv1.DBClusterStatusCreating},
		{"importing is preserved", nbv1.DBClusterStatusImporting, nbv1.DBClusterStatusImporting},
		{"updating is preserved", nbv1.DBClusterStatusUpdating, nbv1.DBClusterStatusUpdating},
		{"none is preserved", nbv1.DBClusterStatusNone, nbv1.DBClusterStatusNone},
		{"failed is preserved", nbv1.DBClusterStatusFailed, nbv1.DBClusterStatusFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := dbClusterStatusAfterSpecUpdate(tt.current); got != tt.want {
				t.Fatalf("dbClusterStatusAfterSpecUpdate(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

// baseClusterSpec returns a cluster spec populated on every field that
// wasClusterSpecChanged compares, so each case can mutate exactly one of them.
func baseClusterSpec() cnpgv1.ClusterSpec {
	storageClass := "test-storage-class"
	apiGroup := "postgresql.cnpg.io"
	defaultQueriesEnabled := false

	return cnpgv1.ClusterSpec{
		InheritedMetadata: &cnpgv1.EmbeddedObjectMetadata{
			Labels: map[string]string{"app": "noobaa"},
		},
		ImageCatalogRef: &cnpgv1.ImageCatalogRef{
			TypedLocalObjectReference: corev1.TypedLocalObjectReference{
				Kind:     "ImageCatalog",
				APIGroup: &apiGroup,
				Name:     "noobaa-db-pg-image-catalog",
			},
			Major: 17,
		},
		Instances: 2,
		Affinity: cnpgv1.AffinityConfiguration{
			TopologyKey: "kubernetes.io/hostname",
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
		},
		StorageConfiguration: cnpgv1.StorageConfiguration{
			StorageClass: &storageClass,
			Size:         "50Gi",
			PersistentVolumeClaimTemplate: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOncePod},
			},
		},
		Monitoring: &cnpgv1.MonitoringConfiguration{
			DisableDefaultQueries: &defaultQueriesEnabled,
		},
		PostgresConfiguration: cnpgv1.PostgresConfiguration{
			Parameters: map[string]string{"wal_level": "replica", "jit": "off"},
		},
		FailoverDelay: defaultFailoverDelaySec,
	}
}

// TestWasClusterSpecChanged covers the gate that decides whether the operator issues an
// update at all. The "unchanged" cases matter as much as the rest: a spurious diff would
// make the operator rewrite the cluster CR on every reconcile.
func TestWasClusterSpecChanged(t *testing.T) {
	t.Parallel()

	otherStorageClass := "other-storage-class"
	disableDefaultQueries := true

	tests := []struct {
		name    string
		mutate  func(spec *cnpgv1.ClusterSpec)
		changed bool
	}{
		{"identical spec", func(*cnpgv1.ClusterSpec) {}, false},
		{"untracked field", func(s *cnpgv1.ClusterSpec) { s.LogLevel = "debug" }, false},
		{"inherited metadata", func(s *cnpgv1.ClusterSpec) { s.InheritedMetadata.Labels["app"] = "other" }, true},
		{"image catalog major version", func(s *cnpgv1.ClusterSpec) { s.ImageCatalogRef.Major = 18 }, true},
		{"instances", func(s *cnpgv1.ClusterSpec) { s.Instances = 3 }, true},
		{"affinity", func(s *cnpgv1.ClusterSpec) { s.Affinity.TopologyKey = "topology.kubernetes.io/zone" }, true},
		{"resources", func(s *cnpgv1.ClusterSpec) {
			s.Resources.Requests[corev1.ResourceCPU] = resource.MustParse("4")
		}, true},
		{"storage class", func(s *cnpgv1.ClusterSpec) {
			s.StorageConfiguration.StorageClass = &otherStorageClass
		}, true},
		{"storage size", func(s *cnpgv1.ClusterSpec) { s.StorageConfiguration.Size = "100Gi" }, true},
		{"pvc template", func(s *cnpgv1.ClusterSpec) {
			s.StorageConfiguration.PersistentVolumeClaimTemplate.AccessModes =
				[]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		}, true},
		{"monitoring", func(s *cnpgv1.ClusterSpec) { s.Monitoring.DisableDefaultQueries = &disableDefaultQueries }, true},
		{"postgres parameter value", func(s *cnpgv1.ClusterSpec) {
			s.PostgresConfiguration.Parameters["jit"] = "on"
		}, true},
		{"postgres parameter added", func(s *cnpgv1.ClusterSpec) {
			s.PostgresConfiguration.Parameters["work_mem"] = "8MB"
		}, true},
		{"failover delay", func(s *cnpgv1.ClusterSpec) { s.FailoverDelay = defaultFailoverDelaySec + 30 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			existing := baseClusterSpec()
			desired := existing.DeepCopy()
			tt.mutate(desired)

			r := &Reconciler{CNPGCluster: &cnpgv1.Cluster{Spec: *desired}}
			if got := r.wasClusterSpecChanged(&existing); got != tt.changed {
				t.Fatalf("wasClusterSpecChanged() = %v, want %v", got, tt.changed)
			}
		})
	}
}
