package system

import (
	"strings"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

func tlsVersionPtr(v nbv1.TLSProtocolVersion) *nbv1.TLSProtocolVersion {
	return &v
}

func TestMapTLSVersion(t *testing.T) {
	if v := MapTLSVersion(nil); v != "" {
		t.Errorf("nil should map to empty, got %q", v)
	}
	if v := MapTLSVersion(tlsVersionPtr(nbv1.TLSVersionTLS12)); v != "TLSv1.2" {
		t.Errorf("expected TLSv1.2, got %q", v)
	}
	if v := MapTLSVersion(tlsVersionPtr(nbv1.TLSVersionTLS13)); v != "TLSv1.3" {
		t.Errorf("expected TLSv1.3, got %q", v)
	}
}

func TestCipherJoin(t *testing.T) {
	if v := strings.Join([]string{}, ":"); v != "" {
		t.Errorf("empty slice should join to empty, got %q", v)
	}
	if v := strings.Join([]string{"TLS_AES_128_GCM_SHA256"}, ":"); v != "TLS_AES_128_GCM_SHA256" {
		t.Errorf("single cipher should have no separator, got %q", v)
	}
	if v := strings.Join([]string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"}, ":"); v != "TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384" {
		t.Errorf("expected colon-joined, got %q", v)
	}
}

func TestCurveJoin(t *testing.T) {
	if v := strings.Join([]string{"X25519MLKEM768", "X25519", "P-256"}, ":"); v != "X25519MLKEM768:X25519:P-256" {
		t.Errorf("expected colon-joined curves, got %q", v)
	}
	var nilSlice []string
	if v := strings.Join(nilSlice, ":"); v != "" {
		t.Errorf("nil slice should join to empty, got %q", v)
	}
}
