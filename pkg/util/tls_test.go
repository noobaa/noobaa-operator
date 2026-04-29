package util

import (
	"strings"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// tlsEnvWithValues returns an env slice that simulates a container that had TLS
// config previously applied — i.e. the TLS vars already carry non-empty values.
func tlsEnvWithValues() []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: "TLS_MIN_VERSION", Value: "TLSv1.2"},
		{Name: "TLS_CIPHERS", Value: "ECDHE-RSA-AES128-GCM-SHA256"},
		{Name: "TLS_GROUPS", Value: "X25519"},
	}
}

func TestValidateTLSSpec(t *testing.T) {
	minV12 := nbv1.VersionTLS12
	minV13 := nbv1.VersionTLS13
	minBad := nbv1.TLSProtocolVersion("TLSv9")

	cases := []struct {
		name    string
		spec    *nbv1.TLSSecuritySpec
		wantErr bool
		errSubs []string // substrings expected in the error message
	}{
		{name: "nil spec", spec: nil, wantErr: false},
		{name: "empty spec", spec: &nbv1.TLSSecuritySpec{}, wantErr: false},
		// --- tlsMinVersion ---
		{name: "valid minVersion TLSv1.2", spec: &nbv1.TLSSecuritySpec{TLSMinVersion: &minV12}, wantErr: false},
		{name: "valid minVersion TLSv1.3", spec: &nbv1.TLSSecuritySpec{TLSMinVersion: &minV13}, wantErr: false},
		{name: "unknown minVersion", spec: &nbv1.TLSSecuritySpec{TLSMinVersion: &minBad},
			wantErr: true, errSubs: []string{"tlsMinVersion", "TLSv9"}},
		// --- tlsCiphers ---
		{name: "valid cipher", spec: &nbv1.TLSSecuritySpec{
			TLSCiphers: []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}}, wantErr: false},
		{name: "unknown cipher", spec: &nbv1.TLSSecuritySpec{TLSCiphers: []string{"FAKE_CIPHER"}},
			wantErr: true, errSubs: []string{"tlsCiphers", "FAKE_CIPHER"}},
		{name: "mixed ciphers — unknown causes error", spec: &nbv1.TLSSecuritySpec{
			TLSCiphers: []string{"TLS_AES_256_GCM_SHA384", "NOT_A_CIPHER"}},
			wantErr: true, errSubs: []string{"NOT_A_CIPHER"}},
		// --- tlsGroups ---
		{name: "valid group", spec: &nbv1.TLSSecuritySpec{
			TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupX25519}}, wantErr: false},
		{name: "unknown group", spec: &nbv1.TLSSecuritySpec{
			TLSGroups: []nbv1.TLSGroup{"UnknownGroup"}},
			wantErr: true, errSubs: []string{"tlsGroups", "UnknownGroup"}},
		{name: "PQC group alone — no InDefaultClients=true group", spec: &nbv1.TLSSecuritySpec{
			TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupSecP256r1MLKEM768}},
			wantErr: true, errSubs: []string{"InDefaultClients=true"}},
		{name: "PQC group with standard fallback", spec: &nbv1.TLSSecuritySpec{
			TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupSecP256r1MLKEM768, nbv1.TLSGroupX25519}},
			wantErr: false},
		// --- multiple errors reported together ---
		{name: "unknown cipher + unknown group", spec: &nbv1.TLSSecuritySpec{
			TLSCiphers: []string{"BAD_CIPHER"},
			TLSGroups:  []nbv1.TLSGroup{"BadGroup"}},
			wantErr: true, errSubs: []string{"BAD_CIPHER", "BadGroup"}},
		{name: "unknown minVersion + unknown cipher", spec: &nbv1.TLSSecuritySpec{
			TLSMinVersion: &minBad,
			TLSCiphers:    []string{"BAD_CIPHER"}},
			wantErr: true, errSubs: []string{"TLSv9", "BAD_CIPHER"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTLSSpec(tc.spec)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			for _, sub := range tc.errSubs {
				if err == nil || !strings.Contains(err.Error(), sub) {
					t.Errorf("expected error to contain %q, got: %v", sub, err)
				}
			}
		})
	}
}


func TestApplyTLSEnvVars(t *testing.T) {
	minV12 := nbv1.VersionTLS12
	minV13 := nbv1.VersionTLS13

	cases := []struct {
		name            string
		disabled        string // value of DISABLE_TLS_SECURITY_CONFIG
		spec            *nbv1.TLSSecuritySpec
		expectedMin     string
		expectedCiphers string
		expectedGroups  string
	}{
		// --- DISABLE_TLS_SECURITY_CONFIG=true -----------------------------------------
		// Previously-set values must be cleared even when a non-nil spec is provided,
		// because the operator must be able to roll back TLS config by setting the flag.
		{
			name:     "disabled clears TLS vars regardless of non-nil spec",
			disabled: "true",
			spec: &nbv1.TLSSecuritySpec{
				TLSMinVersion: &minV12,
				TLSCiphers:    []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
				TLSGroups:     []nbv1.TLSGroup{nbv1.TLSGroupX25519},
			},
			expectedMin:     "",
			expectedCiphers: "",
			expectedGroups:  "",
		},
		{
			name:            "disabled clears TLS vars when spec is nil",
			disabled:        "true",
			spec:            nil,
			expectedMin:     "",
			expectedCiphers: "",
			expectedGroups:  "",
		},
		// --- DISABLE_TLS_SECURITY_CONFIG not set / false --------------------------------
		{
			name:        "enabled applies TLS min version TLSv1.2",
			disabled:    "false",
			spec:        &nbv1.TLSSecuritySpec{TLSMinVersion: &minV12},
			expectedMin: "TLSv1.2",
		},
		{
			name:        "enabled applies TLS min version TLSv1.3",
			disabled:    "false",
			spec:        &nbv1.TLSSecuritySpec{TLSMinVersion: &minV13},
			expectedMin: "TLSv1.3",
		},
		{
			name:            "enabled applies ciphers in OpenSSL format",
			disabled:        "false",
			spec:            &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}},
			expectedCiphers: "ECDHE-RSA-AES128-GCM-SHA256",
		},
		{
			name:     "enabled applies multiple ciphers colon-separated",
			disabled: "false",
			spec: &nbv1.TLSSecuritySpec{
				TLSCiphers: []string{
					"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
					"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
				},
			},
			expectedCiphers: "ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384",
		},
		{
			name:           "enabled applies groups colon-separated",
			disabled:       "false",
			spec:           &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupX25519, nbv1.TLSGroupSecp256r1}},
			expectedGroups: "X25519:secp256r1",
		},
		{
			name:        "enabled with nil spec clears previously set TLS vars",
			disabled:    "false",
			spec:        nil,
			expectedMin: "",
		},
		{
			// ValidateTLSSpec (called in phase 1) would reject this before ApplyTLSEnvVars is reached.
			// MapCiphersToOpenSSL skips unrecognized IANA names, so the result is empty.
			name:            "unrecognized cipher is dropped by MapCiphersToOpenSSL",
			disabled:        "false",
			spec:            &nbv1.TLSSecuritySpec{TLSCiphers: []string{"UNKNOWN_CIPHER"}},
			expectedCiphers: "",
		},
		{
			// ValidateTLSSpec (called in phase 1) would reject this before ApplyTLSEnvVars is reached.
			// ApplyTLSEnvVars passes group names through as-is without re-checking the map.
			name:           "unknown TLS group is passed through as-is",
			disabled:       "false",
			spec:           &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{"UnknownGroup"}},
			expectedGroups: "UnknownGroup",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(DisableTLSSecurityConfigEnv, tc.disabled)
			env := tlsEnvWithValues()
			ApplyTLSEnvVars(&env, tc.spec)

			actual := map[string]string{}
			for _, e := range env {
				actual[e.Name] = e.Value
			}

			if actual["TLS_MIN_VERSION"] != tc.expectedMin {
				t.Errorf("TLS_MIN_VERSION = %q, expected %q", actual["TLS_MIN_VERSION"], tc.expectedMin)
			}
			if actual["TLS_CIPHERS"] != tc.expectedCiphers {
				t.Errorf("TLS_CIPHERS = %q, expected %q", actual["TLS_CIPHERS"], tc.expectedCiphers)
			}
			if actual["TLS_GROUPS"] != tc.expectedGroups {
				t.Errorf("TLS_GROUPS = %q, expected %q", actual["TLS_GROUPS"], tc.expectedGroups)
			}
		})
	}
}
