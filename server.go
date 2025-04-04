package main

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/configuration"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/export"
	"github.com/cloudogu/ces-exporter/maintenance"
	"github.com/cloudogu/ces-exporter/systeminfo"
	libcore "github.com/cloudogu/cesapp-lib/core"
	"github.com/cloudogu/cesapp-lib/registry"
	bup "github.com/cloudogu/k8s-backup-operator/pkg/api/v1"
	componentEcoClient "github.com/cloudogu/k8s-component-operator/pkg/api/ecosystem"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"go.etcd.io/etcd/client/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

const (
	regKeyApi = "/config/ces-exporter/authentication/api_key"
	regKeySsh = "/config/ces-exporter/authentication/public_key"
)

type v1AlphaClientInterface interface {
	componentEcoClient.ComponentV1Alpha1Interface
}

type kubernetesClient interface {
	kubernetes.Interface
}

type server struct {
	ecosystemClient v1AlphaClientInterface
	client          kubernetesClient
	config          *core.Configuration
	bclient         *core.BackupScheduleRuntimeClient
}

func initServerForMultinode(config core.Configuration) (*server, error) {
	clusterConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s cluster config: %w", err)
	}

	rtclient, err := rclient.New(clusterConfig, rclient.Options{})
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	bclient := core.NewBackupScheduleRuntimeClient(rtclient, config.Namespace)

	ecosystemClient, err := componentEcoClient.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	client, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &server{
		ecosystemClient,
		client,
		&config,
		bclient,
	}, nil
}

func initServerForClassic(config core.Configuration) *server {
	reg, err := registry.New(libcore.Registry{
		Type:      "etcd",
		Endpoints: []string{fmt.Sprintf("http://%s:4001", config.ClassicOnlyConfiguration.Fqdn)},
		RetryPolicy: libcore.RetryPolicy{
			Type:          "constant",
			Interval:      5,
			MaxRetryCount: 3,
		},
	})
	if err != nil {
		panic(err.Error())
	}
	apiKeyWatcher := make(chan *client.Response)
	sshKeyWatcher := make(chan *client.Response)

	go func() {
		go func() {
			for event := range apiKeyWatcher {
				slog.Info("Updating api-key because registry config has changed...")
				config.ApiKey = event.Node.Value
			}
		}()

		reg.RootConfig().Watch(context.Background(), regKeyApi, false, apiKeyWatcher)
	}()

	go func() {
		go func() {
			for event := range sshKeyWatcher {
				fmt.Println(event.Node.Value)
			}
		}()

		reg.RootConfig().Watch(context.Background(), regKeySsh, false, sshKeyWatcher)
	}()

	return &server{
		config: &config,
	}
}

func newServer(config core.Configuration) (*server, error) {
	if config.IsClassic {
		return initServerForClassic(config), nil
	} else {
		return initServerForMultinode(config)
	}
}

func (s *server) run(ctx context.Context) error {
	err := bup.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("Failed to register BackupSchedule scheme: %v", err)
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	srv := s.createEndpoints()

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv,
	}
	go func() {
		slog.Info(fmt.Sprintf("listening on %s", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

func (s *server) createMultinodeControllers() (*systeminfo.Controller, *configuration.Controller) {
	configMaps := s.client.CoreV1().ConfigMaps(s.config.Namespace)
	secrets := s.client.CoreV1().Secrets(s.config.Namespace)
	globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
	sensitiveRepo := repository.NewSensitiveDoguConfigRepository(secrets)
	doguRepo := repository.NewDoguConfigRepository(configMaps)
	doguVersionReg := libdogu.NewDoguVersionRegistry(configMaps)

	systemInfoProvider := systeminfo.NewMultinodeSystemInfoProvider(
		configMaps,
		s.client.CoreV1().PersistentVolumeClaims(s.config.Namespace),
		s.config.Namespace,
		s.ecosystemClient.Components(s.config.Namespace),
	)
	systemInfoController := systeminfo.NewController(systemInfoProvider)

	configurationProvider := configuration.NewMultinodeConfigurationProvider(
		s.config.Namespace,
		sensitiveRepo,
		doguRepo,
		globalConfigRepo,
		doguVersionReg,
		s.bclient,
	)
	configController := configuration.NewController(configurationProvider)

	return systemInfoController, configController
}

func (s *server) createEndpoints() http.Handler {
	var systemInfoController *systeminfo.Controller
	var configController *configuration.Controller

	if !s.config.IsClassic {
		systemInfoController, configController = s.createMultinodeControllers()
	} else {
		systemInfoController = &systeminfo.Controller{}
		configController = &configuration.Controller{}
		slog.Error("TODO: Implement classic ces controllers")
	}

	authMiddleware := core.NewAuthMiddleware(s.config)

	rootHandler := http.NewServeMux()
	rootHandler.HandleFunc("GET /health", core.Health)

	rootHandler.HandleFunc("GET /system-info", authMiddleware(systemInfoController.GetSystemInfo))

	rootHandler.HandleFunc("GET /configuration", authMiddleware(configController.GetConfig))

	rootHandler.HandleFunc("GET /export/dogu/{doguName}", authMiddleware(export.GetExportDogu))
	rootHandler.HandleFunc("POST /export/dogu/{doguName}", authMiddleware(export.SetExportDogu))
	rootHandler.HandleFunc("GET /export/mode", authMiddleware(export.GetExportMode))

	rootHandler.HandleFunc("GET /maintenance/mode", authMiddleware(maintenance.GetMaintenanceMode))
	rootHandler.HandleFunc("POST /maintenance/mode", authMiddleware(maintenance.SetMaintenanceMode))

	router := http.NewServeMux()
	router.Handle(fmt.Sprintf("%s/", s.config.BasePath), http.StripPrefix(s.config.BasePath, rootHandler))

	return router
}
