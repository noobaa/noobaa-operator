package util

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// DisableTLSSecurityConfigEnv is the name of the environment variable that disables
// applying TLS security configuration (min version, ciphers, groups) from the NooBaa
// CR to endpoint, core, and admission server. Set to "true" to disable. Default: enabled.
const DisableTLSSecurityConfigEnv = "DISABLE_TLS_SECURITY_CONFIG"

// IsTLSConfigDisabled returns true when the DISABLE_TLS_SECURITY_CONFIG env var is set
// to "true", allowing operators to opt out of the TLS security configuration feature.
func IsTLSConfigDisabled() bool {
	return os.Getenv(DisableTLSSecurityConfigEnv) == "true"
}

// IanaCipherEntry holds the Go numeric ID and OpenSSL-format name for a single
// IANA/Go cipher suite. Node.js endpoints need the OpenSSL name while Go's
// tls.Config.CipherSuites needs the numeric ID.
type IanaCipherEntry struct {
	CipherGoID        uint16
	CipherOpenSSLName string
}

// IanaCipherMap maps supported IANA/Go cipher suite names to their Go ID and
// OpenSSL equivalents. ODF propagates IANA-format names in the NooBaa CR
// APIServerSecurity TLS settings — see ODF's supported ciphers at
// https://github.com/red-hat-storage/ocs-tls-profiles/
var IanaCipherMap = map[string]IanaCipherEntry{
	"TLS_AES_128_GCM_SHA256":                        {CipherGoID: 4865, CipherOpenSSLName: "TLS_AES_128_GCM_SHA256"},
	"TLS_AES_256_GCM_SHA384":                        {CipherGoID: 4866, CipherOpenSSLName: "TLS_AES_256_GCM_SHA384"},
	"TLS_CHACHA20_POLY1305_SHA256":                  {CipherGoID: 4867, CipherOpenSSLName: "TLS_CHACHA20_POLY1305_SHA256"},
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":       {CipherGoID: 49195, CipherOpenSSLName: "ECDHE-ECDSA-AES128-GCM-SHA256"},
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":       {CipherGoID: 49196, CipherOpenSSLName: "ECDHE-ECDSA-AES256-GCM-SHA384"},
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256": {CipherGoID: 52393, CipherOpenSSLName: "ECDHE-ECDSA-CHACHA20-POLY1305"},
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":         {CipherGoID: 49199, CipherOpenSSLName: "ECDHE-RSA-AES128-GCM-SHA256"},
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":         {CipherGoID: 49200, CipherOpenSSLName: "ECDHE-RSA-AES256-GCM-SHA384"},
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":   {CipherGoID: 52392, CipherOpenSSLName: "ECDHE-RSA-CHACHA20-POLY1305"},
}

// TLSGroupEntry holds the Go CurveID for a TLS key exchange group and records
// whether the group is included in the default group lists of both Node.js 24
// and Go 1.24+ clients. Groups where InDefaultClients is false must not be used
// exclusively — a standard group must accompany them as a fallback, otherwise
// clients that don't advertise the group cannot complete the TLS handshake.
type TLSGroupEntry struct {
	CurveID          tls.CurveID
	InDefaultClients bool
}

// TLSGroupMap maps NooBaa TLSGroup constants to their Go CurveID and default-client
// membership. The TLSGroup string values are already valid Node.js/OpenSSL ecdhCurve
// names and can be used as-is for TLS_GROUPS.
var TLSGroupMap = map[nbv1.TLSGroup]TLSGroupEntry{
	nbv1.TLSGroupX25519:             {CurveID: tls.X25519, InDefaultClients: true},
	nbv1.TLSGroupSecp256r1:          {CurveID: tls.CurveP256, InDefaultClients: true},
	nbv1.TLSGroupSecp384r1:          {CurveID: tls.CurveP384, InDefaultClients: true},
	nbv1.TLSGroupSecp521r1:          {CurveID: tls.CurveP521, InDefaultClients: true},
	nbv1.TLSGroupX25519MLKEM768:     {CurveID: tls.X25519MLKEM768, InDefaultClients: true},
	nbv1.TLSGroupSecP256r1MLKEM768:  {CurveID: tls.CurveID(4587), InDefaultClients: false}, // tls.SecP256r1MLKEM768 added in Go 1.26
	nbv1.TLSGroupSecP384r1MLKEM1024: {CurveID: tls.CurveID(4589), InDefaultClients: false}, // tls.SecP384r1MLKEM1024 added in Go 1.26
}

// MapCiphersToOpenSSL converts IANA cipher suite names to OpenSSL format for
// Node.js endpoints. The caller is expected to have validated names via ValidateTLSSpec;
// unrecognized entries are silently skipped.
func MapCiphersToOpenSSL(names []string) string {
	var result []string
	for _, name := range names {
		if entry, ok := IanaCipherMap[name]; ok {
			result = append(result, entry.CipherOpenSSLName)
		}
	}
	return strings.Join(result, ":")
}

// MapCipherSuites converts IANA cipher suite names to uint16 IDs for tls.Config.CipherSuites.
//
// Note: TLS 1.3 cipher suites (e.g. TLS_AES_128_GCM_SHA256) are present in IanaCipherMap and
// their IDs will be included in the returned slice, but Go always enables TLS 1.3 suites
// unconditionally — tls.Config.CipherSuites only affects TLS 1.2 and below.
func MapCipherSuites(names []string) []uint16 {
	var ids []uint16
	var applied []string
	for _, name := range names {
		if entry, ok := IanaCipherMap[name]; ok {
			ids = append(ids, entry.CipherGoID)
			applied = append(applied, name)
		} else {
			log.Warnf("MapCipherSuites: ignoring unrecognized/unsupported cipher suite %q", name)
		}
	}
	if len(applied) > 0 {
		log.Infof("MapCipherSuites: TLS config cipher suites %s", strings.Join(applied, ":"))
	}
	return ids
}

// ApplyTLSEnvVars sets TLS_MIN_VERSION, TLS_CIPHERS, and TLS_GROUPS on entries already
// present in env based on the given TLSSecuritySpec. It is a no-op for containers that
// do not declare those variable names. When tlsSec is nil, or when DISABLE_TLS_SECURITY_CONFIG=true,
// all three are cleared to "" so any previously applied values are removed.
// The caller is responsible for validating tlsSec before calling this function
// (e.g. via ValidateTLSSpec in phase 1 verifying).
func ApplyTLSEnvVars(env *[]corev1.EnvVar, tlsSec *nbv1.TLSSecuritySpec) {
	if IsTLSConfigDisabled() {
		log.Infof("TLS security config disabled via %s, skipping TLS env var updates", DisableTLSSecurityConfigEnv)
		tlsSec = nil
	}
	if tlsSec == nil {
		tlsSec = &nbv1.TLSSecuritySpec{}
	}
	// TLSGroup string values (e.g. "X25519", "secp256r1") are already valid
	// Node.js/OpenSSL ecdhCurve names and can be used as-is for TLS_GROUPS.
	var groupNames []string
	for _, g := range tlsSec.TLSGroups {
		groupNames = append(groupNames, string(g))
	}
	for i := range *env {
		switch (*env)[i].Name {
		case "TLS_MIN_VERSION":
			if tlsSec.TLSMinVersion != nil {
				(*env)[i].Value = string(*tlsSec.TLSMinVersion)
			} else {
				(*env)[i].Value = ""
			}
		case "TLS_CIPHERS":
			(*env)[i].Value = MapCiphersToOpenSSL(tlsSec.TLSCiphers)
		case "TLS_GROUPS":
			(*env)[i].Value = strings.Join(groupNames, ":")
		}
	}
}

// ValidateTLSSpec validates all fields in a TLSSecuritySpec against the values
// defined in this package (IanaCipherMap, TLSGroupMap, known TLSProtocolVersion
// constants). All problems are collected and returned as a single error so the
// caller can surface every issue at once.
//
// Rules enforced:
//   - tlsMinVersion, if set, must be VersionTLS12 or VersionTLS13.
//   - Every tlsCiphers entry must be a key in IanaCipherMap.
//   - Every tlsGroups entry must be a key in TLSGroupMap.
//   - tlsGroups must contain at least one entry with InDefaultClients=true;
//     groups where InDefaultClients is false (SecP256r1MLKEM768,
//     SecP384r1MLKEM1024) cannot be used exclusively because they are absent
//     from the Node.js 24 and Go < 1.26 default group lists and will break
//     all TLS connections.
func ValidateTLSSpec(spec *nbv1.TLSSecuritySpec) error {
	if spec == nil {
		return nil
	}
	if IsTLSConfigDisabled() {
		log.Infof("TLS security config disabled via %s, using default TLS config", DisableTLSSecurityConfigEnv)
		return nil
	}
	var errs []string

	if spec.TLSMinVersion != nil {
		switch *spec.TLSMinVersion {
		case nbv1.VersionTLS12, nbv1.VersionTLS13:
		default:
			errs = append(errs, fmt.Sprintf(
				"tlsMinVersion %q is not valid; allowed values: %q, %q",
				*spec.TLSMinVersion, nbv1.VersionTLS12, nbv1.VersionTLS13))
		}
	}

	for _, c := range spec.TLSCiphers {
		if _, ok := IanaCipherMap[c]; !ok {
			errs = append(errs, fmt.Sprintf("tlsCiphers: %q is not a recognized cipher suite", c))
		}
	}

	hasDefaultClientGroup := false
	for _, g := range spec.TLSGroups {
		entry, ok := TLSGroupMap[g]
		if !ok {
			errs = append(errs, fmt.Sprintf("tlsGroups: %q is not a recognized group", g))
			continue
		}
		if entry.InDefaultClients {
			hasDefaultClientGroup = true
		}
	}
	if len(spec.TLSGroups) > 0 && !hasDefaultClientGroup {
		errs = append(errs, fmt.Sprintf(
			"tlsGroups must include at least one group with InDefaultClients=true "+
				"(e.g. %q, %q, %q, %q, %q) — SecP256r1MLKEM768 and SecP384r1MLKEM1024 "+
				"are absent from the Node.js and Go current default group lists, so "+
				"using them exclusively breaks all TLS connections",
			nbv1.TLSGroupX25519, nbv1.TLSGroupSecp256r1,
			nbv1.TLSGroupSecp384r1, nbv1.TLSGroupSecp521r1, nbv1.TLSGroupX25519MLKEM768))
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(errs, "; "))
}

// MapGroupPreferences converts a list of NooBaa TLSGroup values to the corresponding
// tls.CurveID slice for use in tls.Config.CurvePreferences.
func MapGroupPreferences(groups []nbv1.TLSGroup) []tls.CurveID {
	var ids []tls.CurveID
	var applied []string
	for _, g := range groups {
		if entry, ok := TLSGroupMap[g]; ok {
			ids = append(ids, entry.CurveID)
			applied = append(applied, string(g))
		} else {
			log.Warnf("MapGroupPreferences: ignoring unsupported TLS group %q", g)
		}
	}
	if len(applied) > 0 {
		log.Infof("MapGroupPreferences: TLS group preferences set to %s", strings.Join(applied, ":"))
	}
	return ids
}
