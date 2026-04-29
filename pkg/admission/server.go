package admission

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"syscall"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


const (
	port       = "8080"
	tlscert    = "/etc/certs/tls.cert"
	tlskey     = "/etc/certs/tls.key"
	tlscertolm = "/tmp/k8s-webhook-server/serving-certs/tls.crt"
	tlskeyolm  = "/tmp/k8s-webhook-server/serving-certs/tls.key"
)

var currentTLSConfig atomic.Pointer[tls.Config]

// ReloadTLSConfig rebuilds the TLS configuration by loading certificates from
// disk and applying TLS settings from the TLSProfile referenced by the
// noobaa.io/tls-profile-name annotation on the NooBaa CR, then swaps it
// atomically so new connections use the updated config.
func ReloadTLSConfig() error {
	log := logrus.WithField("admission server", options.Namespace)

	var certPath, keyPath string
	if _, ok := os.LookupEnv("NOOBAA_CLI_DEPLOYMENT"); !ok {
		certPath, keyPath = tlscertolm, tlskeyolm
	} else {
		certPath, keyPath = tlscert, tlskey
	}

	certs, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Errorf("Failed to reload TLS key pair: %v", err)
		return err
	}

	cfg := &tls.Config{Certificates: []tls.Certificate{certs}}
	if err := applyTLSPConfig(cfg, log); err != nil {
		log.Errorf("TLS config not reloaded: %v", err)
		return err
	}
	currentTLSConfig.Store(cfg)
	log.Info("Admission server TLS configuration reloaded")
	return nil
}

// RunAdmissionServer starts the admission HTTPS server.
func RunAdmissionServer() {
	log := logrus.WithField("admission server", options.Namespace)

	if err := ReloadTLSConfig(); err != nil {
		log.Errorf("Failed to load initial TLS config, admission server not started: %v", err)
		return
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{
			GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
				return currentTLSConfig.Load(), nil
			},
		},
	}

	sh := ServerHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", sh.serve)
	server.Handler = mux

	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	log.Infof("Admission server start running and listening on port: %s", port)

	util.OnSignal(func() {
		log.Info("Got shutdown signal, shutting down webhook server gracefully...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shut down the admission server: %v", err)
		}
	}, syscall.SIGINT, syscall.SIGTERM)
}

// applyTLSPConfig fetches the NooBaa CR, reads the noobaa.io/tls-profile-name
// annotation, fetches the referenced TLSProfile CR, and applies its TLS settings
// for "noobaa.io" to the given tls.Config.
func applyTLSPConfig(tlsConfig *tls.Config, log *logrus.Entry) error {
	noobaa := &nbv1.NooBaa{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "noobaa",
			Namespace: options.Namespace,
		},
	}
	if !util.KubeCheckQuiet(noobaa) {
		log.Info("NooBaa CR not found, using default TLS config for admission server")
		return nil
	}

	profile, err := util.LoadTLSProfile(noobaa.GetAnnotations(), options.Namespace)
	if err != nil {
		return err
	}
	if profile == nil {
		log.Info("noobaa.io/tls-profile-name annotation not set, using default TLS config for admission server")
		return nil
	}

	tlsCfg := util.TLSConfigFromProfile(profile)
	if tlsCfg == nil {
		log.Infof("TLSProfile %q has no rule matching %s, using default TLS config for admission server",
			profile.Name, util.TLSProfileDomain)
		return nil
	}

	if err := util.ApplyTLSConfigToGoTLS(tlsCfg, tlsConfig); err != nil {
		return fmt.Errorf("TLSProfile %q: %w", profile.Name, err)
	}
	log.Infof("Admission server TLS configured from TLSProfile %q (version=%s, ciphers=%v, groups=%v)",
		profile.Name, tlsCfg.Version, tlsCfg.Ciphers, tlsCfg.Groups)
	return nil
}
