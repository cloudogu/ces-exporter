package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/ces-exporter/configuration"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/export"
	"github.com/cloudogu/ces-exporter/maintenance"
	"github.com/cloudogu/ces-exporter/systeminfo"
	bup "github.com/cloudogu/k8s-backup-operator/pkg/api/v1"
	componentEcoClient "github.com/cloudogu/k8s-component-operator/pkg/api/ecosystem"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	"io/fs"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	regKeyApi      = "/config/ces-exporter/authentication/api_key"
	regKeySsh      = "/config/ces-exporter/authentication/public_key"
	sshKeyFileMode = fs.FileMode(0600)
	authorizedKeys = "/root/.ssh/authorized_keys"
)

type ecosystemComponentClient interface {
	componentEcoClient.ComponentV1Alpha1Interface
}

type ecosystemDogusClient interface {
	ecoSystemV2.EcoSystemV2Interface
}

type kubernetesClient interface {
	kubernetes.Interface
}

type controllerProvider interface {
	createControllers(ctx context.Context) (*systeminfo.Controller, *configuration.Controller, *maintenance.Controller, *export.Controller)
}

type server struct {
	config             *core.Configuration
	controllerProvider controllerProvider
}

func newServer(config core.Configuration) (*server, error) {
	var provider controllerProvider
	if config.IsClassic {
		var err error
		provider, err = newClassicControllerProvider(&config)
		if err != nil {
			return nil, fmt.Errorf("failed to create classic server: %w", err)
		}
	} else {
		err := bup.AddToScheme(scheme.Scheme)
		if err != nil {
			return nil, fmt.Errorf("failed to register BackupSchedule scheme: %w", err)
		}

		provider, err = newMultinodeControllerProvider(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create multinode controller provider: %w", err)
		}
	}
	return &server{
		controllerProvider: provider,
		config:             &config,
	}, nil
}

func (s *server) run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	srv := s.createEndpoints(ctx)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv,
	}
	go func() {
		slog.Info(fmt.Sprintf("listening on %s", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error start serving.", "err", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		slog.Info("shutting down...")
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
			slog.Error("error shutting down http server.", "err", err)
		}
	}()
	wg.Wait()

	return nil
}

func (s *server) createEndpoints(ctx context.Context) http.Handler {
	systemInfoController, configController, maintenanceModeController, exportModeController := s.controllerProvider.createControllers(ctx)

	authMiddleware := core.NewAuthMiddleware(s.config)

	rootHandler := http.NewServeMux()
	rootHandler.HandleFunc("GET /health", core.Health)

	rootHandler.HandleFunc("GET /system-info", authMiddleware(systemInfoController.GetSystemInfo))

	rootHandler.HandleFunc("GET /configuration", authMiddleware(configController.GetConfig))

	rootHandler.HandleFunc("GET /export/dogu", authMiddleware(exportModeController.GetExportDogu))
	rootHandler.HandleFunc("POST /export/dogu/{doguName}", authMiddleware(exportModeController.SetExportDogu))
	rootHandler.HandleFunc("GET /export/mode", authMiddleware(exportModeController.GetExportMode))

	rootHandler.HandleFunc("GET /maintenance/mode", authMiddleware(maintenanceModeController.GetMaintenanceMode))
	rootHandler.HandleFunc("POST /maintenance/mode", authMiddleware(maintenanceModeController.SetMaintenanceMode))

	router := http.NewServeMux()
	router.Handle(fmt.Sprintf("%s/", s.config.BasePath), http.StripPrefix(s.config.BasePath, rootHandler))

	return router
}
