package util

import (
	"testing"

	ocstlsv1 "github.com/red-hat-storage/ocs-tls-profiles/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeProfile(version ocstlsv1.TLSProtocolVersion, ciphers []ocstlsv1.TLSCipherSuite, groups []ocstlsv1.TLSGroupName) *ocstlsv1.TLSProfile {
	return &ocstlsv1.TLSProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: ocstlsv1.TLSProfileSpec{
			Rules: []ocstlsv1.TLSProfileRules{
				{
					Selectors: []ocstlsv1.Selector{"noobaa.io"},
					Config: ocstlsv1.TLSConfig{
						Version: version,
						Ciphers: ciphers,
						Groups:  groups,
					},
				},
			},
		},
	}
}

func TestTLSConfigFromProfile_Nil(t *testing.T) {
	if TLSConfigFromProfile(nil) != nil {
		t.Fatal("expected nil for nil profile")
	}
}

func TestTLSConfigFromProfile_NoMatch(t *testing.T) {
	profile := &ocstlsv1.TLSProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: ocstlsv1.TLSProfileSpec{
			Rules: []ocstlsv1.TLSProfileRules{
				{
					Selectors: []ocstlsv1.Selector{"other.io"},
					Config: ocstlsv1.TLSConfig{
						Version: ocstlsv1.VersionTLS1_2,
						Ciphers: []ocstlsv1.TLSCipherSuite{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
						Groups:  []ocstlsv1.TLSGroupName{"secp256r1"},
					},
				},
			},
		},
	}
	if TLSConfigFromProfile(profile) != nil {
		t.Fatal("expected nil when no selector matches noobaa.io")
	}
}

func TestTLSConfigFromProfile_Match(t *testing.T) {
	profile := makeProfile(
		ocstlsv1.VersionTLS1_2,
		[]ocstlsv1.TLSCipherSuite{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		[]ocstlsv1.TLSGroupName{"secp256r1"},
	)
	cfg := TLSConfigFromProfile(profile)
	if cfg == nil {
		t.Fatal("expected non-nil TLSConfig")
	}
	if cfg.Version != ocstlsv1.VersionTLS1_2 {
		t.Fatalf("unexpected version: %s", cfg.Version)
	}
}

func TestApplyTLSEnvVars_NilProfile(t *testing.T) {
	env := []corev1.EnvVar{
		{Name: "TLS_MIN_VERSION", Value: "old"},
		{Name: "TLS_CIPHERS", Value: "old"},
		{Name: "TLS_GROUPS", Value: "old"},
	}
	if err := ApplyTLSEnvVars(&env, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range env {
		if e.Value != "" {
			t.Fatalf("expected empty value for %s, got %q", e.Name, e.Value)
		}
	}
}

func TestApplyTLSEnvVars_ValidProfile(t *testing.T) {
	profile := makeProfile(
		ocstlsv1.VersionTLS1_2,
		[]ocstlsv1.TLSCipherSuite{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		[]ocstlsv1.TLSGroupName{"secp256r1"},
	)
	env := []corev1.EnvVar{
		{Name: "TLS_MIN_VERSION"},
		{Name: "TLS_CIPHERS"},
		{Name: "TLS_GROUPS"},
		{Name: "OTHER", Value: "unchanged"},
	}
	if err := ApplyTLSEnvVars(&env, profile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vals := map[string]string{}
	for _, e := range env {
		vals[e.Name] = e.Value
	}
	if vals["TLS_MIN_VERSION"] != "TLSv1.2" {
		t.Errorf("unexpected TLS_MIN_VERSION: %q", vals["TLS_MIN_VERSION"])
	}
	if vals["TLS_CIPHERS"] == "" {
		t.Error("expected non-empty TLS_CIPHERS")
	}
	if vals["TLS_GROUPS"] == "" {
		t.Error("expected non-empty TLS_GROUPS")
	}
	if vals["OTHER"] != "unchanged" {
		t.Error("OTHER env var should not be modified")
	}
}

func TestApplyTLSEnvVars_InvalidProfile(t *testing.T) {
	// TLS 1.2 version with a TLS 1.3-only cipher — should fail validation
	profile := makeProfile(
		ocstlsv1.VersionTLS1_2,
		[]ocstlsv1.TLSCipherSuite{"TLS_AES_128_GCM_SHA256"},
		[]ocstlsv1.TLSGroupName{"secp256r1"},
	)
	env := []corev1.EnvVar{{Name: "TLS_MIN_VERSION"}, {Name: "TLS_CIPHERS"}, {Name: "TLS_GROUPS"}}
	if err := ApplyTLSEnvVars(&env, profile); err == nil {
		t.Fatal("expected validation error for incompatible TLS 1.3 cipher with TLS 1.2")
	}
}

func TestApplyTLSEnvVars_NoMatchingEnvVars(t *testing.T) {
	profile := makeProfile(
		ocstlsv1.VersionTLS1_2,
		[]ocstlsv1.TLSCipherSuite{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		[]ocstlsv1.TLSGroupName{"secp256r1"},
	)
	env := []corev1.EnvVar{{Name: "OTHER", Value: "val"}}
	if err := ApplyTLSEnvVars(&env, profile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env[0].Value != "val" {
		t.Error("non-TLS env var should not be modified")
	}
}
