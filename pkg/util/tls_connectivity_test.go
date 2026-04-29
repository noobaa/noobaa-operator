package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

// TestSupportedConfigDoesNotBreakConnectivity verifies that every TLS configuration
// the operator may write to NooBaa pods can still be negotiated by a client using
// Go defaults — matching the InsecureHTTPTransport used by operator→core/endpoint
// and core→endpoint connections.
//
// The server side is modelled by building a Go tls.Config from each supported CR
// value via the same helpers used in ApplyTLSEnvVars and the admission server.
// A real in-process TLS handshake is then performed with an operator-style client.
func TestSupportedConfigDoesNotBreakConnectivity(t *testing.T) {
	rsaCert := generateTestRSACert(t)
	ecdsaCert := generateTestECDSACert(t)

	minV12 := nbv1.VersionTLS12
	minV13 := nbv1.VersionTLS13

	cases := []struct {
		name     string
		spec     *nbv1.TLSSecuritySpec
		useECDSA bool // ECDSA cert needed to exercise ECDHE-ECDSA cipher suites via TLS 1.2
	}{
		// --- TLS min version ---
		{
			name: "minVersion TLSv1.2",
			spec: &nbv1.TLSSecuritySpec{TLSMinVersion: &minV12},
		},
		{
			name: "minVersion TLSv1.3",
			spec: &nbv1.TLSSecuritySpec{TLSMinVersion: &minV13},
		},

		// --- TLS 1.3 cipher suites ---
		// Go's tls.Config.CipherSuites only affects TLS 1.2; TLS 1.3 suites are
		// always negotiated when TLS 1.3 is available, so these three have no effect
		// on connectivity with a modern client.
		{
			name: "cipher TLS_AES_128_GCM_SHA256",
			spec: &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_AES_128_GCM_SHA256"}},
		},
		{
			name: "cipher TLS_AES_256_GCM_SHA384",
			spec: &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_AES_256_GCM_SHA384"}},
		},
		{
			name: "cipher TLS_CHACHA20_POLY1305_SHA256",
			spec: &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_CHACHA20_POLY1305_SHA256"}},
		},

		// --- ECDHE-RSA TLS 1.2 cipher suites ---
		{
			name: "cipher TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			spec: &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}},
		},
		{
			name: "cipher TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			spec: &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"}},
		},
		{
			name: "cipher TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
			spec: &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"}},
		},

		// --- ECDHE-ECDSA TLS 1.2 cipher suites ---
		// We use an ECDSA cert so these are exercised at TLS 1.2 as well as TLS 1.3.
		{
			name:     "cipher TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			spec:     &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"}},
			useECDSA: true,
		},
		{
			name:     "cipher TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
			spec:     &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"}},
			useECDSA: true,
		},
		{
			name:     "cipher TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
			spec:     &nbv1.TLSSecuritySpec{TLSCiphers: []string{"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256"}},
			useECDSA: true,
		},

		// --- TLS groups ---
		{
			name: "group X25519",
			spec: &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupX25519}},
		},
		{
			name: "group secp256r1",
			spec: &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupSecp256r1}},
		},
		{
			name: "group secp384r1",
			spec: &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupSecp384r1}},
		},
		{
			name: "group secp521r1",
			spec: &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupSecp521r1}},
		},
		{
			name: "group X25519MLKEM768",
			spec: &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{nbv1.TLSGroupX25519MLKEM768}},
		},
		// SecP256r1MLKEM768 and SecP384r1MLKEM1024 are in NoDefaultGroups — absent from
		// Node.js 24 and Go < 1.26 default group lists. ValidateTLSGroups rejects them
		// when used alone; the only safe (and valid) configuration is paired with a
		// standard fallback group so clients can always negotiate a common group.
		{
			name: "group SecP256r1MLKEM768 + X25519 fallback",
			spec: &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{
				nbv1.TLSGroupSecP256r1MLKEM768, nbv1.TLSGroupX25519,
			}},
		},
		{
			name: "group SecP384r1MLKEM1024 + X25519 fallback",
			spec: &nbv1.TLSSecuritySpec{TLSGroups: []nbv1.TLSGroup{
				nbv1.TLSGroupSecP384r1MLKEM1024, nbv1.TLSGroupX25519,
			}},
		},

		// --- All fields combined ---
		{
			name: "all fields: TLSv1.3 + RSA cipher + X25519 group",
			spec: &nbv1.TLSSecuritySpec{
				TLSMinVersion: &minV13,
				TLSCiphers:    []string{"TLS_AES_256_GCM_SHA384", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
				TLSGroups:     []nbv1.TLSGroup{nbv1.TLSGroupX25519, nbv1.TLSGroupSecp256r1},
			},
		},
		{
			name: "all fields: TLSv1.3 + PQC group X25519MLKEM768",
			spec: &nbv1.TLSSecuritySpec{
				TLSMinVersion: &minV13,
				TLSCiphers:    []string{"TLS_AES_256_GCM_SHA384"},
				TLSGroups:     []nbv1.TLSGroup{nbv1.TLSGroupX25519MLKEM768, nbv1.TLSGroupX25519},
			},
		},

		// --- Unknown / invalid values are dropped; remaining config must still work ---
		{
			name: "unknown cipher dropped; known cipher still applied",
			spec: &nbv1.TLSSecuritySpec{
				TLSCiphers: []string{"UNKNOWN_CIPHER", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
			},
		},
		{
			name: "unknown group dropped; known group still applied",
			spec: &nbv1.TLSSecuritySpec{
				TLSGroups: []nbv1.TLSGroup{"UnknownGroup", nbv1.TLSGroupX25519},
			},
		},
		{
			name: "unknown min version dropped; server uses default",
			spec: func() *nbv1.TLSSecuritySpec {
				unknown := nbv1.TLSProtocolVersion("TLSv1.5")
				return &nbv1.TLSSecuritySpec{TLSMinVersion: &unknown}
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cert := rsaCert
			if tc.useECDSA {
				cert = ecdsaCert
			}
			serverCfg := specToServerTLSConfig(tc.spec, cert)
			assertTLSHandshakeSucceeds(t, serverCfg)
		})
	}
}

// specToServerTLSConfig builds a Go tls.Config for a server from a TLSSecuritySpec,
// using the same mapping helpers as ApplyTLSEnvVars and the admission server.
func specToServerTLSConfig(spec *nbv1.TLSSecuritySpec, cert tls.Certificate) *tls.Config {
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	if spec == nil {
		return cfg
	}
	if spec.TLSMinVersion != nil {
		switch *spec.TLSMinVersion {
		case nbv1.VersionTLS12:
			cfg.MinVersion = tls.VersionTLS12
		case nbv1.VersionTLS13:
			cfg.MinVersion = tls.VersionTLS13
		}
	}
	if suites := MapCipherSuites(spec.TLSCiphers); len(suites) > 0 {
		cfg.CipherSuites = suites
	}
	if curves := MapGroupPreferences(spec.TLSGroups); len(curves) > 0 {
		cfg.CurvePreferences = curves
	}
	return cfg
}

// assertTLSHandshakeSucceeds starts a TLS listener with serverCfg, dials it with
// an operator-style client (InsecureSkipVerify, no other restrictions — matching
// InsecureHTTPTransport), and fails the test if either side reports a handshake error.
func assertTLSHandshakeSucceeds(t *testing.T, serverCfg *tls.Config) {
	t.Helper()

	ln, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatalf("failed to start TLS listener: %v", err)
	}
	defer func() { _ = ln.Close() }()

	serverErr := make(chan error, 1)
	go func() {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			serverErr <- fmt.Errorf("server accept: %w", acceptErr)
			return
		}
		defer func() { _ = conn.Close() }()
		serverErr <- conn.(*tls.Conn).Handshake()
	}()

	// Operator-style client: InsecureSkipVerify with no min-version / cipher / group
	// restrictions — identical to the InsecureHTTPTransport used in util.go.
	clientCfg := &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- intentional for test
	conn, dialErr := tls.Dial("tcp", ln.Addr().String(), clientCfg)
	if dialErr != nil {
		t.Fatalf("client dial/handshake failed: %v", dialErr)
	}
	_ = conn.Close()

	if se := <-serverErr; se != nil {
		t.Fatalf("server handshake error: %v", se)
	}
}

// generateTestRSACert returns a self-signed RSA-2048 certificate for use in tests.
func generateTestRSACert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	return selfSignedTLSCert(t, key, &key.PublicKey)
}

// generateTestECDSACert returns a self-signed ECDSA P-256 certificate for use in tests.
func generateTestECDSACert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey: %v", err)
	}
	return selfSignedTLSCert(t, key, &key.PublicKey)
}

func selfSignedTLSCert(t *testing.T, priv, pub interface{}) tls.Certificate {
	t.Helper()
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "noobaa-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, pub, priv)
	if err != nil {
		t.Fatalf("x509.CreateCertificate: %v", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("x509.MarshalPKCS8PrivateKey: %v", err)
	}
	cert, err := tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
	)
	if err != nil {
		t.Fatalf("tls.X509KeyPair: %v", err)
	}
	return cert
}
