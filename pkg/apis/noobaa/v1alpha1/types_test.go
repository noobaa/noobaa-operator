package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"
)

func tlsVersionPtr(v TLSProtocolVersion) *TLSProtocolVersion {
	return &v
}

func TestTLSSecuritySpec_JSONRoundTrip(t *testing.T) {
	spec := TLSSecuritySpec{
		TLSVersion:          tlsVersionPtr(TLSVersionTLS13),
		TLSCipherSuites:     []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
		TLSCurvePreferences: []string{"X25519MLKEM768", "X25519", "P-256"},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded TLSSecuritySpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(spec, decoded) {
		t.Errorf("round-trip mismatch:\n  original: %+v\n  decoded:  %+v", spec, decoded)
	}
}

func TestTLSSecuritySpec_JSONFieldNames(t *testing.T) {
	spec := TLSSecuritySpec{
		TLSVersion:          tlsVersionPtr(TLSVersionTLS12),
		TLSCipherSuites:     []string{"cipher1"},
		TLSCurvePreferences: []string{"curve1"},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map failed: %v", err)
	}

	for _, field := range []string{"tlsVersion", "tlsCipherSuites", "tlsCurvePreferences"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("expected JSON field %q not found in output: %s", field, string(data))
		}
	}

	for _, field := range []string{"minTLSVersion", "ciphers", "curvePreferences"} {
		if _, ok := raw[field]; ok {
			t.Errorf("old JSON field %q should not be present in output: %s", field, string(data))
		}
	}
}

func TestTLSSecuritySpec_EmptyJSON(t *testing.T) {
	spec := TLSSecuritySpec{}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	if string(data) != "{}" {
		t.Errorf("empty spec should marshal to {}, got %s", string(data))
	}
}

func TestTLSSecuritySpec_DeepCopy(t *testing.T) {
	spec := TLSSecuritySpec{
		TLSVersion:          tlsVersionPtr(TLSVersionTLS13),
		TLSCipherSuites:     []string{"TLS_AES_128_GCM_SHA256"},
		TLSCurvePreferences: []string{"X25519", "P-256"},
	}

	copied := spec.DeepCopy()

	if !reflect.DeepEqual(&spec, copied) {
		t.Errorf("deep copy mismatch:\n  original: %+v\n  copy:     %+v", spec, *copied)
	}

	// Mutate the copy and verify original is unaffected
	*copied.TLSVersion = TLSVersionTLS12
	copied.TLSCipherSuites[0] = "MODIFIED"
	copied.TLSCurvePreferences = append(copied.TLSCurvePreferences, "P-384")

	if *spec.TLSVersion != TLSVersionTLS13 {
		t.Error("original TLSVersion was mutated by deep copy modification")
	}
	if spec.TLSCipherSuites[0] != "TLS_AES_128_GCM_SHA256" {
		t.Error("original TLSCipherSuites was mutated by deep copy modification")
	}
	if len(spec.TLSCurvePreferences) != 2 {
		t.Error("original TLSCurvePreferences was mutated by deep copy modification")
	}
}

func TestTLSSecuritySpec_DeepCopyNil(t *testing.T) {
	var spec *TLSSecuritySpec
	copied := spec.DeepCopy()
	if copied != nil {
		t.Error("deep copy of nil should return nil")
	}
}

func TestSecuritySpec_DeepCopyBothTLS(t *testing.T) {
	spec := SecuritySpec{
		IngressControllerSecurity: TLSSecuritySpec{
			TLSVersion:      tlsVersionPtr(TLSVersionTLS13),
			TLSCipherSuites: []string{"cipher-ingress"},
		},
		APIServerSecurity: TLSSecuritySpec{
			TLSVersion:      tlsVersionPtr(TLSVersionTLS12),
			TLSCipherSuites: []string{"cipher-api"},
		},
	}

	copied := spec.DeepCopy()

	if !reflect.DeepEqual(&spec, copied) {
		t.Errorf("deep copy mismatch:\n  original: %+v\n  copy:     %+v", spec, *copied)
	}

	// Mutate ingress in copy, verify original is unaffected
	*copied.IngressControllerSecurity.TLSVersion = TLSVersionTLS12
	if *spec.IngressControllerSecurity.TLSVersion != TLSVersionTLS13 {
		t.Error("original IngressControllerSecurity.TLSVersion was mutated")
	}

	// Mutate API server in copy, verify original is unaffected
	copied.APIServerSecurity.TLSCipherSuites[0] = "MODIFIED"
	if spec.APIServerSecurity.TLSCipherSuites[0] != "cipher-api" {
		t.Error("original APIServerSecurity.TLSCipherSuites was mutated")
	}
}

func TestSecuritySpec_JSONRoundTrip(t *testing.T) {
	spec := SecuritySpec{
		IngressControllerSecurity: TLSSecuritySpec{
			TLSVersion:          tlsVersionPtr(TLSVersionTLS13),
			TLSCipherSuites:     []string{"TLS_AES_256_GCM_SHA384"},
			TLSCurvePreferences: []string{"X25519MLKEM768"},
		},
		APIServerSecurity: TLSSecuritySpec{
			TLSVersion:      tlsVersionPtr(TLSVersionTLS12),
			TLSCipherSuites: []string{"ECDHE-RSA-AES128-GCM-SHA256"},
		},
	}

	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map failed: %v", err)
	}

	if _, ok := raw["ingressControllerSecurity"]; !ok {
		t.Error("expected JSON field ingressControllerSecurity not found")
	}
	if _, ok := raw["apiServerSecurity"]; !ok {
		t.Error("expected JSON field apiServerSecurity not found")
	}
	if _, ok := raw["endpointTLS"]; ok {
		t.Error("old JSON field endpointTLS should not be present")
	}

	var decoded SecuritySpec
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !reflect.DeepEqual(spec, decoded) {
		t.Errorf("round-trip mismatch:\n  original: %+v\n  decoded:  %+v", spec, decoded)
	}
}

func TestTLSProtocolVersion_Values(t *testing.T) {
	if TLSVersionTLS12 != "VersionTLS12" {
		t.Errorf("TLSVersionTLS12 = %q, want VersionTLS12", TLSVersionTLS12)
	}
	if TLSVersionTLS13 != "VersionTLS13" {
		t.Errorf("TLSVersionTLS13 = %q, want VersionTLS13", TLSVersionTLS13)
	}
}
