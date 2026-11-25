// This file wires Kubernetes clients, TLS serving and the controller lifecycle.

package exportpoc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Config provides namespace, image, key, TLS and timing settings for the optional PoC.
type Config struct {
	Namespace, SidecarImage, PublicKeyConfigMap, PublicKeyDataKey, CertDir, WebhookName string
	ReplacementTimeout                                                                  time.Duration
}

// Runtime hosts the webhook and controller using their shared Kubernetes discovery.
type Runtime struct {
	config    Config
	discovery *discovery
}

// New creates independent Kubernetes clients with bounded API requests and applies the default pod replacement timeout.
func New(config Config, restConfig *rest.Config) (*Runtime, error) {
	restConfig = rest.CopyConfig(restConfig)
	restConfig.Timeout = 15 * time.Second
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	if config.ReplacementTimeout == 0 {
		config.ReplacementTimeout = 10 * time.Minute
	}
	return &Runtime{config: config, discovery: &discovery{kube: kube, dynamic: dyn, namespace: config.Namespace}}, nil
}

// Run serves TLS before reconciling and stops both components with the exporter.
func (r *Runtime) Run(ctx context.Context) error {
	var certificate tls.Certificate
	for {
		var err error
		certificate, err = tls.LoadX509KeyPair(r.config.CertDir+"/tls.crt", r.config.CertDir+"/tls.key")
		if err == nil {
			break
		}
		slog.Info("waiting for PoC webhook certificate", "error", err)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}
	listener, err := net.Listen("tcp", ":9443")
	if err != nil {
		return fmt.Errorf("listen for PoC webhook: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate-pods", r.admission)
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12, GetCertificate: func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
		// Reload mounted certificates to support cert-manager renewal.
		cert, err := tls.LoadX509KeyPair(r.config.CertDir+"/tls.crt", r.config.CertDir+"/tls.key")
		if err != nil {
			return &certificate, nil
		}
		return &cert, nil
	}}}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go r.control(runCtx)
	go func() {
		<-runCtx.Done()
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(stopCtx)
	}()
	err = server.Serve(tls.NewListener(listener, server.TLSConfig))
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
