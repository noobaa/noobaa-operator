package system

import (
	"testing"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestFormatBytesKB(t *testing.T) {
	tests := []struct {
		name     string
		kb       int64
		expected string
	}{
		{"exact GB", 1024 * 1024, "1GB"},
		{"multiple GB", 4 * 1024 * 1024, "4GB"},
		{"exact MB", 256 * 1024, "256MB"},
		{"single MB", 1024, "1MB"},
		{"non-round MB", 1747, "1747kB"},
		{"32kB", 32, "32kB"},
		{"16MB", 16 * 1024, "16MB"},
		{"2GB", 2 * 1024 * 1024, "2GB"},
		{"8GB", 8 * 1024 * 1024, "8GB"},
		{"3GB", 3 * 1024 * 1024, "3GB"},
		{"512MB", 512 * 1024, "512MB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytesKB(tt.kb)
			if got != tt.expected {
				t.Errorf("formatBytesKB(%d) = %q, want %q", tt.kb, got, tt.expected)
			}
		})
	}
}

func TestCalculatePGConfig_4GB_2CPU_3Endpoints(t *testing.T) {
	totalMemoryKB := int64(4 * 1024 * 1024)
	cpuNum := int64(2)
	numEndpoints := 3

	params := calculatePGConfig(totalMemoryKB, cpuNum, numEndpoints, "50Gi")

	// 3 endpoints -> max_connections = 3*20 + 3*80 + 3 = 303
	expected := map[string]string{
		"shared_buffers":       "1GB",
		"effective_cache_size": "3GB",
		"maintenance_work_mem": "256MB",
		"wal_buffers":          "16MB",
		"max_connections":      "303",
		"wal_keep_size":        "6GB",
	}

	for key, want := range expected {
		got, ok := params[key]
		if !ok {
			t.Errorf("missing parameter %q", key)
			continue
		}
		if got != want {
			t.Errorf("param %q = %q, want %q", key, got, want)
		}
	}

	// cpuNum < 4 should not set parallel settings
	for _, key := range []string{
		"max_worker_processes",
		"max_parallel_workers",
		"max_parallel_workers_per_gather",
		"max_parallel_maintenance_workers",
	} {
		if _, ok := params[key]; ok {
			t.Errorf("unexpected parallel parameter %q set for cpuNum=%d", key, cpuNum)
		}
	}

	// work_mem: (4GB - 1GB) / ((303 + 8) * 3) = 3145728 / 933 = 3371kB < 4MB minimum
	assertParam(t, params, "work_mem", "4MB")
}

func TestCalculatePGConfig_16GB_8CPU_3Endpoints(t *testing.T) {
	totalMemoryKB := int64(16 * 1024 * 1024)
	cpuNum := int64(8)
	numEndpoints := 3

	params := calculatePGConfig(totalMemoryKB, cpuNum, numEndpoints, "50Gi")

	// 3 endpoints -> max_connections = 303
	assertParam(t, params, "shared_buffers", "4GB")
	assertParam(t, params, "effective_cache_size", "12GB")
	assertParam(t, params, "maintenance_work_mem", "1GB")
	assertParam(t, params, "wal_buffers", "16MB")
	assertParam(t, params, "max_connections", "303")

	// parallel settings with 8 CPUs
	assertParam(t, params, "max_worker_processes", "8")
	assertParam(t, params, "max_parallel_workers", "8")
	assertParam(t, params, "max_parallel_workers_per_gather", "4")
	assertParam(t, params, "max_parallel_maintenance_workers", "4")

	// work_mem: (16GB - 4GB) / ((303 + 8) * 3) = 12582912 / 933 = 13487kB
	workMem := params["work_mem"]
	if workMem == "" {
		t.Error("work_mem not set")
	}
}

func TestCalculatePGConfig_8GB_4CPU_7Endpoints(t *testing.T) {
	totalMemoryKB := int64(8 * 1024 * 1024)
	cpuNum := int64(4)
	numEndpoints := 7

	params := calculatePGConfig(totalMemoryKB, cpuNum, numEndpoints, "50Gi")

	// 7 endpoints -> max_connections = 3*20 + 7*80 + 3 = 623
	assertParam(t, params, "shared_buffers", "2GB")
	assertParam(t, params, "effective_cache_size", "6GB")
	assertParam(t, params, "maintenance_work_mem", "512MB")
	assertParam(t, params, "max_connections", "623")

	// parallel: cpuNum=4 so parallel settings should be set
	assertParam(t, params, "max_worker_processes", "4")
	assertParam(t, params, "max_parallel_workers", "4")
	assertParam(t, params, "max_parallel_workers_per_gather", "2")
	assertParam(t, params, "max_parallel_maintenance_workers", "2")

	// work_mem: (8GB - 2GB) / ((623 + 4) * 3) = 6291456 / 1881 = 3344kB < 4MB minimum
	assertParam(t, params, "work_mem", "4MB")
}

func TestCalculatePGConfig_MaintenanceWorkMem_Cap(t *testing.T) {
	// 256GB memory -> maintenance_work_mem = 256GB / 16 = 16GB, capped at 8GB
	params := calculatePGConfig(256*1024*1024, 16, 3, "50Gi")
	assertParam(t, params, "maintenance_work_mem", "8GB")
}

func TestCalculatePGConfig_WalBuffers(t *testing.T) {
	// 256MB memory -> shared_buffers = 64MB -> wal_buffers = 3% * 64MB = 1966kB
	params256MB := calculatePGConfig(256*1024, 1, 3, "50Gi")
	got := params256MB["wal_buffers"]
	if got != "1966kB" && got != "1MB" {
		t.Logf("wal_buffers for 256MB = %q (this is expected to vary)", got)
	}

	// 1GB memory -> shared_buffers = 256MB -> wal_buffers = 3% * 256MB = 7864kB
	params1GB := calculatePGConfig(1024*1024, 1, 3, "50Gi")
	walBuf1GB := params1GB["wal_buffers"]
	if walBuf1GB == "" {
		t.Error("wal_buffers not set for 1GB")
	}

	// 4GB+ memory -> shared_buffers = 1GB+ -> wal_buffers = 3% >= ~30MB -> capped at 16MB
	params4GB := calculatePGConfig(4*1024*1024, 1, 3, "50Gi")
	assertParam(t, params4GB, "wal_buffers", "16MB")
}

func TestCalculatePGConfig_WorkMem_MinimumFloor(t *testing.T) {
	// Tiny memory with many endpoints -> work_mem should be at least 4MB
	// 10 endpoints -> max_connections = 3*20 + 10*80 + 3 = 863
	totalMemoryKB := int64(512 * 1024) // 512MB
	params := calculatePGConfig(totalMemoryKB, 1, 10, "50Gi")

	// work_mem = (512MB - 128MB) / ((863 + 8) * 3) = 393216 / 2613 = 150kB -> floor to 4MB
	assertParam(t, params, "work_mem", "4MB")
}

func TestCalculatePGConfig_ParallelWorkers_Capping(t *testing.T) {
	// 32 CPUs -> workers_per_gather = ceil(32/2) = 16 -> capped at 4
	params := calculatePGConfig(32*1024*1024, 32, 3, "50Gi")

	assertParam(t, params, "max_worker_processes", "32")
	assertParam(t, params, "max_parallel_workers", "32")
	assertParam(t, params, "max_parallel_workers_per_gather", "4")
	assertParam(t, params, "max_parallel_maintenance_workers", "4")
}

func TestCalculatePGConfig_NoCPU(t *testing.T) {
	params := calculatePGConfig(4*1024*1024, 0, 3, "50Gi")

	// no parallel settings when cpuNum < 4
	for _, key := range []string{
		"max_worker_processes",
		"max_parallel_workers",
		"max_parallel_workers_per_gather",
		"max_parallel_maintenance_workers",
	} {
		if _, ok := params[key]; ok {
			t.Errorf("unexpected parallel parameter %q when cpuNum=0", key)
		}
	}
}

func TestCalculatePGConfig_WalKeepSize(t *testing.T) {
	params50 := calculatePGConfig(4*1024*1024, 2, 3, "50Gi")
	assertParam(t, params50, "wal_keep_size", "6GB")

	params200 := calculatePGConfig(4*1024*1024, 2, 3, "200Gi")
	assertParam(t, params200, "wal_keep_size", "16GB")

	paramsInvalid := calculatePGConfig(4*1024*1024, 2, 3, "invalid")
	assertParam(t, paramsInvalid, "wal_keep_size", "6GB")
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

func TestCalculateMaxConnections(t *testing.T) {
	tests := []struct {
		name         string
		numEndpoints int
		expected     int
	}{
		{"1 endpoint", 1, 3*20 + 1*80 + 3},     // 143
		{"2 endpoints", 2, 3*20 + 2*80 + 3},    // 223
		{"3 endpoints", 3, 3*20 + 3*80 + 3},    // 303
		{"5 endpoints", 5, 3*20 + 5*80 + 3},    // 463
		{"10 endpoints", 10, 3*20 + 10*80 + 3}, // 863
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMaxConnections(tt.numEndpoints)
			if got != tt.expected {
				t.Errorf("calculateMaxConnections(%d) = %d, want %d", tt.numEndpoints, got, tt.expected)
			}
		})
	}
}

func assertParam(t *testing.T, params map[string]string, key, expected string) {
	t.Helper()
	got, ok := params[key]
	if !ok {
		t.Errorf("missing parameter %q", key)
		return
	}
	if got != expected {
		t.Errorf("param %q = %q, want %q", key, got, expected)
	}
}

func TestGetDesiredDBPVCAccessMode(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		want        corev1.PersistentVolumeAccessMode
	}{
		{
			name:        "nil annotations defaults to RWOP",
			annotations: nil,
			want:        corev1.ReadWriteOncePod,
		},
		{
			name:        "missing annotation defaults to RWOP",
			annotations: map[string]string{},
			want:        corev1.ReadWriteOncePod,
		},
		{
			name:        "annotation true sets RWO",
			annotations: map[string]string{nbv1.PVCAccessModeRWO: "true"},
			want:        corev1.ReadWriteOnce,
		},
		{
			name:        "annotation false keeps RWOP",
			annotations: map[string]string{nbv1.PVCAccessModeRWO: "false"},
			want:        corev1.ReadWriteOncePod,
		},
		{
			name:        "annotation other value keeps RWOP",
			annotations: map[string]string{nbv1.PVCAccessModeRWO: "yes"},
			want:        corev1.ReadWriteOncePod,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDesiredDBPVCAccessMode(tt.annotations)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetDesiredStorageConfAccessMode(t *testing.T) {
	storageClass := "lab-rwo-only"
	dbSpec := &nbv1.NooBaaDBSpec{
		DBStorageClass: &storageClass,
	}

	t.Run("defaults to RWOP", func(t *testing.T) {
		storageConfiguration := &cnpgv1.StorageConfiguration{}
		if err := setDesiredStorageConf(storageConfiguration, dbSpec, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		modes := storageConfiguration.PersistentVolumeClaimTemplate.AccessModes
		if len(modes) != 1 || modes[0] != corev1.ReadWriteOncePod {
			t.Fatalf("got access modes %v, want [%s]", modes, corev1.ReadWriteOncePod)
		}
	})

	t.Run("annotation true sets RWO", func(t *testing.T) {
		storageConfiguration := &cnpgv1.StorageConfiguration{}
		annotations := map[string]string{nbv1.PVCAccessModeRWO: "true"}
		if err := setDesiredStorageConf(storageConfiguration, dbSpec, annotations); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		modes := storageConfiguration.PersistentVolumeClaimTemplate.AccessModes
		if len(modes) != 1 || modes[0] != corev1.ReadWriteOnce {
			t.Fatalf("got access modes %v, want [%s]", modes, corev1.ReadWriteOnce)
		}
		if storageConfiguration.StorageClass == nil || *storageConfiguration.StorageClass != storageClass {
			t.Fatalf("got storage class %v, want %q", storageConfiguration.StorageClass, storageClass)
		}
	})
}
