package util

import (
	"crypto/tls"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

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

// TLSGroupToID maps NooBaa TLSGroup constants to Go tls.CurveID values.
// TODO: When ODF is updated to include the new TLSGroups, we should switch to using the
// ODF supported tls groups
var TLSGroupToID = map[nbv1.TLSGroup]tls.CurveID{
	nbv1.TLSGroupX25519:             tls.X25519,
	nbv1.TLSGroupSecp256r1:          tls.CurveP256,
	nbv1.TLSGroupSecp384r1:          tls.CurveP384,
	nbv1.TLSGroupSecp521r1:          tls.CurveP521,
	nbv1.TLSGroupX25519MLKEM768:     tls.X25519MLKEM768,
	nbv1.TLSGroupSecP256r1MLKEM768:  4587, // tls.SecP256r1MLKEM768 in Go 1.26+
	nbv1.TLSGroupSecP384r1MLKEM1024: 4589, // tls.SecP384r1MLKEM1024 in Go 1.26+
}

// MapCiphersToOpenSSL converts IANA cipher suite names to OpenSSL format for
// Node.js endpoints. Unrecognized/unsupported IANA names are skipped with a warning.
func MapCiphersToOpenSSL(names []string) string {
	var result []string
	for _, name := range names {
		if entry, ok := IanaCipherMap[name]; ok {
			result = append(result, entry.CipherOpenSSLName)
		} else {
			log.Warnf("MapCiphersToOpenSSL: skipping unrecognized/unsupported IANA cipher suite %q", name)
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

// MapGroupPreferences converts a list of NooBaa TLSGroup values to the corresponding
// tls.CurveID slice for use in tls.Config.CurvePreferences.
func MapGroupPreferences(groups []nbv1.TLSGroup) []tls.CurveID {
	var ids []tls.CurveID
	var applied []string
	for _, g := range groups {
		if id, ok := TLSGroupToID[g]; ok {
			ids = append(ids, id)
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
