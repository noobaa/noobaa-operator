package admission

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/sirupsen/logrus"
)

var (
	certs tls.Certificate
	err   error
)

const (
	port       = "8080"
	tlscert    = "/etc/certs/tls.cert"
	tlskey     = "/etc/certs/tls.key"
	tlscertolm = "/tmp/k8s-webhook-server/serving-certs/tls.crt"
	tlskeyolm  = "/tmp/k8s-webhook-server/serving-certs/tls.key"
)

// RunAdmissionServer starts the admission https server
func RunAdmissionServer() {
	namespace := options.Namespace
	log := logrus.WithField("admission server", namespace)

	_, ok := os.LookupEnv("NOOBAA_CLI_DEPLOYMENT")
	if !ok {
		certs, err = tls.LoadX509KeyPair(tlscertolm, tlskeyolm)
		if err != nil {
			log.Errorf("Filed to load olm key pair: %v", err)
			return
		}
	} else {
		certs, err = tls.LoadX509KeyPair(tlscert, tlskey)
		if err != nil {
			log.Errorf("Filed to load key pair: %v", err)
			return
		}
	}

	server := &http.Server{
		Addr:      fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{certs}},
	}

	// define http server and server handler
	sh := ServerHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", sh.serve)
	server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := server.ListenAndServeTLS("", ""); err != nil {
			log.Errorf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	log.Infof("Server running and listening in port: %s", port)

	// listening shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Info("Got shutdown signal, shutting down webhook server gracefully...")
	err = server.Shutdown(context.Background())
	if err != nil {
		log.Info("Failed to Shutdown admission server")
	}
}
