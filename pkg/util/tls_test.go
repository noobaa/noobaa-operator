package util

import (
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
			name:            "enabled skips unrecognized cipher without error",
			disabled:        "false",
			spec:            &nbv1.TLSSecuritySpec{TLSCiphers: []string{"UNKNOWN_CIPHER"}},
			expectedCiphers: "",
		},
		{
			name:           "enabled skips unknown TLS group without error",
			disabled:       "false",
			spec:           &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{"UnknownGroup"}},
			expectedGroups: "",
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
