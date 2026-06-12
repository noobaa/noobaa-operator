package system

import "testing"

func assertParam(t *testing.T, params map[string]string, key, expected string) {
	t.Helper()
	if got := params[key]; got != expected {
		t.Errorf("%s = %q, want %q", key, got, expected)
	}
}

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
