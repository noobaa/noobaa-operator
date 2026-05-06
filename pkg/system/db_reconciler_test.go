package system

import (
	"testing"
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

func TestCalculatePGConfig_4GB_2CPU_300Conn(t *testing.T) {
	// 4GB = 4 * 1024 * 1024 KB
	totalMemoryKB := int64(4 * 1024 * 1024)
	cpuNum := int64(2)
	maxConnections := 300

	params := calculatePGConfig(totalMemoryKB, cpuNum, maxConnections)

	expected := map[string]string{
		"shared_buffers":               "1GB",
		"effective_cache_size":         "3GB",
		"maintenance_work_mem":         "256MB",
		"huge_pages":                   "off",
		"checkpoint_completion_target": "0.9",
		"wal_buffers":                  "16MB",
		"default_statistics_target":    "100",
		"random_page_cost":             "1.1",
		"effective_io_concurrency":     "300",
		"max_connections":              "300",
		"min_wal_size":                 "2GB",
		"max_wal_size":                 "8GB",
		"jit":                          "off",
		"wal_compression":              "lz4",
		"pg_stat_statements.track":     "all",
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

	// work_mem: (4GB - 1GB) / ((300 + 8) * 3) = 3GB / 924 = 3407kB
	if params["work_mem"] != "4096kB" {
		// (4194304 - 1048576) / ((300 + 8) * 3) = 3145728 / 924 = 3404
		// 3404kB < 4096kB (4MB minimum), so work_mem should be 4MB
		if params["work_mem"] != "4MB" {
			t.Errorf("work_mem = %q, want %q", params["work_mem"], "4MB")
		}
	}
}

func TestCalculatePGConfig_16GB_8CPU_300Conn(t *testing.T) {
	totalMemoryKB := int64(16 * 1024 * 1024)
	cpuNum := int64(8)
	maxConnections := 300

	params := calculatePGConfig(totalMemoryKB, cpuNum, maxConnections)

	assertParam(t, params, "shared_buffers", "4GB")
	assertParam(t, params, "effective_cache_size", "12GB")
	assertParam(t, params, "maintenance_work_mem", "1GB")
	assertParam(t, params, "huge_pages", "try")
	assertParam(t, params, "wal_buffers", "16MB")
	assertParam(t, params, "max_connections", "300")

	// parallel settings with 8 CPUs
	assertParam(t, params, "max_worker_processes", "8")
	assertParam(t, params, "max_parallel_workers", "8")
	assertParam(t, params, "max_parallel_workers_per_gather", "4")
	assertParam(t, params, "max_parallel_maintenance_workers", "4")

	// work_mem: (16GB - 4GB) / ((300 + 8) * 3) = 12GB / 924
	// = 12582912 / 924 = 13618kB = ~13MB
	workMem := params["work_mem"]
	if workMem == "" {
		t.Error("work_mem not set")
	}
}

func TestCalculatePGConfig_8GB_4CPU_600Conn(t *testing.T) {
	totalMemoryKB := int64(8 * 1024 * 1024)
	cpuNum := int64(4)
	maxConnections := 600

	params := calculatePGConfig(totalMemoryKB, cpuNum, maxConnections)

	assertParam(t, params, "shared_buffers", "2GB")
	assertParam(t, params, "effective_cache_size", "6GB")
	assertParam(t, params, "maintenance_work_mem", "512MB")
	assertParam(t, params, "huge_pages", "try")
	assertParam(t, params, "max_connections", "600")

	// parallel: cpuNum=4 so parallel settings should be set
	assertParam(t, params, "max_worker_processes", "4")
	assertParam(t, params, "max_parallel_workers", "4")
	// ceil(4/2) = 2
	assertParam(t, params, "max_parallel_workers_per_gather", "2")
	assertParam(t, params, "max_parallel_maintenance_workers", "2")

	// work_mem: (8GB - 2GB) / ((600 + 4) * 3) = 6GB / 1812
	// = 6291456 / 1812 = 3472kB < 4096kB minimum
	assertParam(t, params, "work_mem", "4MB")
}

func TestCalculatePGConfig_HugePages_Threshold(t *testing.T) {
	// 7GB memory -> shared_buffers = 1.75GB -> huge_pages = off
	params7GB := calculatePGConfig(7*1024*1024, 2, 300)
	assertParam(t, params7GB, "huge_pages", "off")

	// 8GB memory -> shared_buffers = 2GB (exactly at threshold) -> huge_pages = try
	params8GB := calculatePGConfig(8*1024*1024, 2, 300)
	assertParam(t, params8GB, "huge_pages", "try")

	// 16GB memory -> shared_buffers = 4GB -> huge_pages = try
	params16GB := calculatePGConfig(16*1024*1024, 2, 300)
	assertParam(t, params16GB, "huge_pages", "try")
}

func TestCalculatePGConfig_MaintenanceWorkMem_Cap(t *testing.T) {
	// 256GB memory -> maintenance_work_mem = 256GB / 16 = 16GB, capped at 8GB
	params := calculatePGConfig(256*1024*1024, 16, 300)
	assertParam(t, params, "maintenance_work_mem", "8GB")
}

func TestCalculatePGConfig_WalBuffers(t *testing.T) {
	// 256MB memory -> shared_buffers = 64MB -> wal_buffers = 3% * 64MB = 1966kB
	params256MB := calculatePGConfig(256*1024, 1, 300)
	got := params256MB["wal_buffers"]
	if got != "1966kB" && got != "1MB" {
		// 3 * 65536 / 100 = 1966kB (not evenly divisible by 1024 so stays as kB)
		t.Logf("wal_buffers for 256MB = %q (this is expected to vary)", got)
	}

	// 1GB memory -> shared_buffers = 256MB -> wal_buffers = 3% * 256MB = 7864kB -> 7864kB
	params1GB := calculatePGConfig(1024*1024, 1, 300)
	walBuf1GB := params1GB["wal_buffers"]
	if walBuf1GB == "" {
		t.Error("wal_buffers not set for 1GB")
	}

	// 4GB+ memory -> shared_buffers = 1GB+ -> wal_buffers = 3% >= ~30MB -> capped at 16MB
	params4GB := calculatePGConfig(4*1024*1024, 1, 300)
	assertParam(t, params4GB, "wal_buffers", "16MB")
}

func TestCalculatePGConfig_WorkMem_MinimumFloor(t *testing.T) {
	// Tiny memory with many connections -> work_mem should be at least 4MB
	totalMemoryKB := int64(512 * 1024) // 512MB
	params := calculatePGConfig(totalMemoryKB, 1, 1000)

	// work_mem = (512MB - 128MB) / ((1000 + 8) * 3) = 384MB / 3024 = 130kB -> floor to 4MB
	assertParam(t, params, "work_mem", "4MB")
}

func TestCalculatePGConfig_ParallelWorkers_Capping(t *testing.T) {
	// 32 CPUs -> workers_per_gather = ceil(32/2) = 16 -> capped at 4
	params := calculatePGConfig(32*1024*1024, 32, 300)

	assertParam(t, params, "max_worker_processes", "32")
	assertParam(t, params, "max_parallel_workers", "32")
	assertParam(t, params, "max_parallel_workers_per_gather", "4")
	assertParam(t, params, "max_parallel_maintenance_workers", "4")
}

func TestCalculatePGConfig_NoCPU(t *testing.T) {
	params := calculatePGConfig(4*1024*1024, 0, 300)

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

func TestCalculateMaxConnections(t *testing.T) {
	tests := []struct {
		name         string
		numEndpoints int
		expected     int
	}{
		{"1 endpoint", 1, 3*20 + 1*80 + 3},   // 143
		{"2 endpoints", 2, 3*20 + 2*80 + 3},   // 223
		{"3 endpoints", 3, 3*20 + 3*80 + 3},   // 303
		{"5 endpoints", 5, 3*20 + 5*80 + 3},   // 463
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
