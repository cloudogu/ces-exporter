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
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"go.etcd.io/etcd/client/v2"
	"io/fs"
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
	regKeyApi      = "/config/ces-exporter/authentication/api_key"
	regKeySsh      = "/config/ces-exporter/authentication/public_key"
	sshKeyFileMode = fs.FileMode(0600)
	authorizedKeys = "/root/.ssh/authorized_keys"
)

type watchConfigurationContext interface {
	Watch(ctx context.Context, key string, recursive bool, eventChannel chan *client.Response)
	Get(key string) (string, error)
}

type ecosystemComponentClient interface {
	componentEcoClient.ComponentV1Alpha1Interface
}

type ecosystemDogusClient interface {
	ecoSystemV2.EcoSystemV2Interface
}

type kubernetesClient interface {
	kubernetes.Interface
}

type server struct {
	componentClient ecosystemComponentClient
	doguClient      ecosystemDogusClient
	client          kubernetesClient
	config          *core.Configuration
	bclient         *core.BackupScheduleRuntimeClient
}

// startCronJob starts an asynchronous task - it can not return anything but will log error if the job fails
func (s *server) startCronJob(dogus ecoSystemV2.DoguInterface) {
	if s.config.CronExp == "" {
		return
	}
	cj := export.NewCronJob(s.config.CronExp, dogus, s.config.Namespace, s.config.VerboseCron)
	err := cj.Run()

	if err != nil {
		slog.Error("Failed to start cronjob:", "err", err)
	}
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

	componentClient, err := componentEcoClient.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create component client: %w", err)
	}

	doguClient, err := ecoSystemV2.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dogu client: %w", err)
	}

	cl, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &server{
		componentClient,
		doguClient,
		cl,
		&config,
		bclient,
	}, nil
}

func initServerForClassic(config core.Configuration) *server {
	watchApiKeyConfig(config.ClassicOnlyConfiguration.Registry, &config)
	watchSshKeyConfig(config.ClassicOnlyConfiguration.Registry, config.ClassicOnlyConfiguration.WriteFile)

	return &server{
		config: &config,
	}
}

func createRegistry(config core.Configuration) watchConfigurationContext {
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
		// This error can only occur if type is not equal etcd - as this is hardcoded to etcd, the error will not occur
		panic(err.Error())
	}

	return reg.RootConfig()
}

func watchApiKeyConfig(reg watchConfigurationContext, config *core.Configuration) {
	apiKeyWatcher := make(chan *client.Response)

	go func() {
		v, err := reg.Get(regKeyApi)
		if err != nil {
			slog.Error(err.Error())
		} else {
			config.ApiKey = v
		}

		go func() {
			for event := range apiKeyWatcher {
				slog.Info("Updating api-key because registry config has changed...")
				config.ApiKey = event.Node.Value
			}
		}()

		reg.Watch(context.Background(), regKeyApi, false, apiKeyWatcher)
	}()
}

func writeAuthorizedKey(v string, write core.WriteFileFunc) {
	slog.Info(fmt.Sprintf("The authroized ssz public key has changed to %s", v))
	err := write(authorizedKeys, []byte(v), sshKeyFileMode)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not write changed ssh key to file: %s", err.Error()))
	} else {
		slog.Info("Successfully wrote new ssh key to authorized_keys file...")
	}
}

func watchSshKeyConfig(reg watchConfigurationContext, write core.WriteFileFunc) {
	sshKeyWatcher := make(chan *client.Response)

	go func() {
		v, err := reg.Get(regKeySsh)
		if err != nil {
			slog.Error(err.Error())
		} else {
			writeAuthorizedKey(v, write)
		}

		go func() {
			for event := range sshKeyWatcher {
				writeAuthorizedKey(event.Node.Value, write)
			}
		}()

		reg.Watch(context.Background(), regKeySsh, false, sshKeyWatcher)
	}()
}

func newServer(config core.Configuration) (*server, error) {
	if config.IsClassic {
		config.ClassicOnlyConfiguration.Registry = createRegistry(config)
		config.ClassicOnlyConfiguration.WriteFile = os.WriteFile
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

func (s *server) createMultinodeControllers() (*systeminfo.Controller, *configuration.Controller, *maintenance.Controller, *export.Controller) {
	configMaps := s.client.CoreV1().ConfigMaps(s.config.Namespace)
	secrets := s.client.CoreV1().Secrets(s.config.Namespace)
	globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
	sensitiveRepo := repository.NewSensitiveDoguConfigRepository(secrets)
	doguRepo := repository.NewDoguConfigRepository(configMaps)
	doguVersionReg := libdogu.NewDoguVersionRegistry(configMaps)
	services := s.client.CoreV1().Services(s.config.Namespace)
	dogus := s.doguClient.Dogus(s.config.Namespace)

	systemInfoProvider := systeminfo.NewMultinodeSystemInfoProvider(
		configMaps,
		s.client.CoreV1().PersistentVolumeClaims(s.config.Namespace),
		s.config.Namespace,
		s.componentClient.Components(s.config.Namespace),
	)
	systemInfoController := systeminfo.NewController(systemInfoProvider)

	exportModeProvider := export.NewMultinodeExportModeProvider(s.config.Namespace, configMaps, dogus, services)
	exportModeController := export.NewController(exportModeProvider)

	// start cron job for setting export mode. See env variable "EXPORT_CRON" for timetable
	go s.startCronJob(dogus)

	configurationProvider := configuration.NewMultinodeConfigurationProvider(
		s.config.Namespace,
		sensitiveRepo,
		doguRepo,
		globalConfigRepo,
		doguVersionReg,
		s.bclient,
	)
	configController := configuration.NewController(configurationProvider)

	maintenanceModeProvider := maintenance.NewMultinodeProvider(globalConfigRepo)
	maintenanceModeController := maintenance.NewController(maintenanceModeProvider)

	return systemInfoController, configController, maintenanceModeController, exportModeController
}

func (s *server) createEndpoints() http.Handler {
	var systemInfoController *systeminfo.Controller
	var configController *configuration.Controller
	var maintenanceModeController *maintenance.Controller
	var exportModeController *export.Controller

	if !s.config.IsClassic {
		systemInfoController, configController, maintenanceModeController, exportModeController = s.createMultinodeControllers()
	} else {
		systemInfoController = &systeminfo.Controller{}
		configController = &configuration.Controller{}
		maintenanceModeProvider := maintenance.NewClassicProvider()
		maintenanceModeController = maintenance.NewController(maintenanceModeProvider)
		slog.Error("TODO: Implement classic ces controllers")
	}

	authMiddleware := core.NewAuthMiddleware(s.config)

	rootHandler := http.NewServeMux()
	rootHandler.HandleFunc("GET /health", core.Health)

	rootHandler.HandleFunc("GET /system-info", authMiddleware(systemInfoController.GetSystemInfo))

	rootHandler.HandleFunc("GET /configuration", authMiddleware(configController.GetConfig))

	rootHandler.HandleFunc("GET /export/dogu/{doguName}", authMiddleware(exportModeController.GetExportDogu))
	rootHandler.HandleFunc("POST /export/dogu/{doguName}", authMiddleware(exportModeController.SetExportDogu))
	rootHandler.HandleFunc("GET /export/mode", authMiddleware(exportModeController.GetExportMode))

	rootHandler.HandleFunc("GET /maintenance/mode", authMiddleware(maintenanceModeController.GetMaintenanceMode))
	rootHandler.HandleFunc("POST /maintenance/mode", authMiddleware(maintenanceModeController.SetMaintenanceMode))

	router := http.NewServeMux()
	router.Handle(fmt.Sprintf("%s/", s.config.BasePath), http.StripPrefix(s.config.BasePath, rootHandler))

	return router
}
