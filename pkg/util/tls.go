package util

import (
	"crypto/tls"
	"fmt"
	"strings"

	ocstlsv1 "github.com/red-hat-storage/ocs-tls-profiles/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// TLSProfileAnnotation is the NooBaa CR annotation that holds the name of the
// TLSProfile CR to use for TLS configuration.
const TLSProfileAnnotation = "noobaa.io/tls-profile-name"

// TLSProfileDomain is the domain used to resolve the TLSConfig from a TLSProfile.
const TLSProfileDomain = "noobaa.io"

// LoadTLSProfile reads the TLSProfileAnnotation from annotations and fetches the
// referenced TLSProfile CR from namespace. Returns nil, nil when the annotation is absent.
// Returns an error when the annotation is present but the CR cannot be fetched.
func LoadTLSProfile(annotations map[string]string, namespace string) (*ocstlsv1.TLSProfile, error) {
	profileName, ok := annotations[TLSProfileAnnotation]
	if !ok || profileName == "" {
		return nil, nil
	}
	profile := &ocstlsv1.TLSProfile{}
	profile.Name = profileName
	profile.Namespace = namespace
	if _, _, err := KubeGet(profile); err != nil {
		if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
			// TLSProfile CRD is not installed on this cluster - skip silently.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch TLSProfile %q referenced by annotation %s: %w",
			profileName, TLSProfileAnnotation, err)
	}
	return profile, nil
}

// ApplyTLSConfigToGoTLS validates tlsCfg and copies the resulting version, cipher suites,
// and curve preferences into dst. It is a no-op when tlsCfg is nil.
func ApplyTLSConfigToGoTLS(tlsCfg *ocstlsv1.TLSConfig, dst *tls.Config) error {
	if tlsCfg == nil {
		return nil
	}
	src, err := ocstlsv1.ValidateAndGetGoTLSConfig(tlsCfg)
	if err != nil {
		return err
	}
	dst.MinVersion = src.MinVersion
	dst.CipherSuites = src.CipherSuites
	dst.CurvePreferences = src.CurvePreferences
	return nil
}

// TLSConfigFromProfile resolves the TLSConfig for "noobaa.io" from a TLSProfile CR.
// Returns nil when profile is nil or no rule matches.
func TLSConfigFromProfile(profile *ocstlsv1.TLSProfile) *ocstlsv1.TLSConfig {
	if profile == nil {
		return nil
	}
	cfg, ok := ocstlsv1.GetConfigForServer(profile, TLSProfileDomain, "")
	if !ok {
		return nil
	}
	return cfg
}

// GoCiphersAndCurvesFromTLSConfig maps a TLSConfig using ValidateAndGetGoTLSConfig.
// It returns nil slices when cfg is nil or both cipher and group lists are empty.
func GoCiphersAndCurvesFromTLSConfig(cfg *ocstlsv1.TLSConfig) ([]uint16, []tls.CurveID, error) {
	if cfg == nil || (len(cfg.Ciphers) == 0 && len(cfg.Groups) == 0) {
		return nil, nil, nil
	}
	goCfg, err := ocstlsv1.ValidateAndGetGoTLSConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	return goCfg.CipherSuites, goCfg.CurvePreferences, nil
}

// OpenSSLCipherAndGroupStringsFromTLSConfig returns colon-separated OpenSSL cipher and group
// strings for TLS_CIPHERS and TLS_GROUPS. It runs ValidateAndGetGoTLSConfig and OpenSSLConfigFrom once.
func OpenSSLCipherAndGroupStringsFromTLSConfig(cfg *ocstlsv1.TLSConfig) (ciphers, groups string, err error) {
	if cfg == nil {
		return "", "", nil
	}
	if len(cfg.Ciphers) == 0 && len(cfg.Groups) == 0 {
		return "", "", nil
	}
	goCfg, err := ocstlsv1.ValidateAndGetGoTLSConfig(cfg)
	if err != nil {
		return "", "", err
	}
	ssl := ocstlsv1.OpenSSLConfigFrom(goCfg)
	if ssl == nil {
		return "", "", fmt.Errorf("OpenSSLConfigFrom returned nil after successful TLS validation")
	}
	if len(ssl.Ciphers) > 0 {
		ciphers = strings.Join(ssl.Ciphers, ":")
	}
	if len(ssl.Groups) > 0 {
		groups = strings.Join(ssl.Groups, ":")
	}
	return ciphers, groups, nil
}

// ApplyTLSEnvVars resolves the TLS configuration from profile and sets TLS_MIN_VERSION,
// TLS_CIPHERS, and TLS_GROUPS on matching entries already present in env.
// It is a no-op for containers that do not declare those variable names.
// When profile is nil (or no rule matches "noobaa.io"), the env vars are cleared to "".
func ApplyTLSEnvVars(env *[]corev1.EnvVar, profile *ocstlsv1.TLSProfile) error {
	tlsCfg := TLSConfigFromProfile(profile)
	ciphers, groups, err := OpenSSLCipherAndGroupStringsFromTLSConfig(tlsCfg)
	if err != nil {
		return err
	}
	minVer := ""
	if tlsCfg != nil && tlsCfg.Version != "" {
		minVer = string(tlsCfg.Version)
	}
	for i := range *env {
		switch (*env)[i].Name {
		case "TLS_MIN_VERSION":
			(*env)[i].Value = minVer
		case "TLS_CIPHERS":
			(*env)[i].Value = ciphers
		case "TLS_GROUPS":
			(*env)[i].Value = groups
		}
	}
	return nil
}
