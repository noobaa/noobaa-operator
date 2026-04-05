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

// ReloadTLSConfig rebuilds the TLS configuration by loading certificates
// from disk and reading the NooBaa CR's APIServerSecurity settings, then
// swaps it atomically. New TLS connections will use the updated config.
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
	applyAPIServerTLS(cfg, log)
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

// applyAPIServerTLS fetches the NooBaa CR and applies APIServerSecurity TLS
// properties to the given tls.Config when they are set.
func applyAPIServerTLS(tlsConfig *tls.Config, log *logrus.Entry) {
	noobaa := &nbv1.NooBaa{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "noobaa",
			Namespace: options.Namespace,
		},
	}
	if !util.KubeCheckQuiet(noobaa) {
		log.Info("NooBaa CR not found, using default TLS config for admission server")
		return
	}

	spec := noobaa.Spec.Security.APIServerSecurity
	if spec == nil {
		log.Info("APIServerSecurity not configured, using default TLS config for admission server")
		return
	}

	if spec.TLSMinVersion != nil {
		switch *spec.TLSMinVersion {
		case nbv1.VersionTLS12:
			tlsConfig.MinVersion = tls.VersionTLS12
			log.Info("Admission server TLS min version set to TLSv1.2")
		case nbv1.VersionTLS13:
			tlsConfig.MinVersion = tls.VersionTLS13
			log.Info("Admission server TLS min version set to TLSv1.3")
		}
	}

	if len(spec.TLSCiphers) > 0 {
		tlsConfig.CipherSuites = util.MapCipherSuites(spec.TLSCiphers)
	}

	if len(spec.TLSGroups) > 0 {
		tlsConfig.CurvePreferences = util.MapGroupPreferences(spec.TLSGroups)
	}
}
