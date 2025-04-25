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
	"go.etcd.io/etcd/client/v2"
	"log/slog"
	"os"
	"strconv"
)

const (
	fqdnEnv = "FQDN"
	// defaultVolumeIncreaseFactor represents the default percentage by which a Dogu's volume size should be increased
	// when calculating the target volume size.
	defaultVolumeIncreaseFactor = 0.3
)

type watchConfigurationContext interface {
	Watch(ctx context.Context, key string, recursive bool, eventChannel chan *client.Response)
	Get(key string) (string, error)
}

type writeFileFunc func(name string, data []byte, perm os.FileMode) error

type classicControllerProvider struct {
	config *core.Configuration
	write  writeFileFunc
	reg    watchConfigurationContext
}

func (c *classicControllerProvider) createControllers(ctx context.Context) (*systeminfo.Controller, *configuration.Controller, *maintenance.Controller, *export.Controller) {
	watchApiKeyConfig(ctx, c.reg, c.config)
	watchSshKeyConfig(ctx, c.reg, c.write)

	exportModeProvider := export.NewClassicExportModeProvider(*c.config)
	exportModeController := export.NewController(exportModeProvider)

	systemInfoProvider := systeminfo.NewSingleNodeSystemInfoProvider(c.config.VolumesBasePath, getVolumeIncreaseFactor(c.reg))
	systemInfoController := systeminfo.NewController(systemInfoProvider)

	maintenanceModeProvider := maintenance.NewClassicProvider()
	maintenanceModeController := maintenance.NewController(maintenanceModeProvider)

	configurationProvider := configuration.NewClassicProvider()
	configurationController := configuration.NewController(configurationProvider)

	return systemInfoController,
		configurationController,
		maintenanceModeController,
		exportModeController
}

func newClassicControllerProvider(conf *core.Configuration) (*classicControllerProvider, error) {
	fqdn := os.Getenv(fqdnEnv)
	if fqdn == "" {
		return nil, fmt.Errorf("FQDN environment variable is unset")
	}

	reg := createRegistry(fqdn)

	return &classicControllerProvider{
		config: conf,
		write:  os.WriteFile,
		reg:    reg,
	}, nil
}

func createRegistry(fqdn string) watchConfigurationContext {
	reg, err := registry.New(libcore.Registry{
		Type:      "etcd",
		Endpoints: []string{fmt.Sprintf("http://%s:4001", fqdn)},
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

func watchApiKeyConfig(ctx context.Context, reg watchConfigurationContext, config *core.Configuration) {
	apiKeyWatcher := make(chan *client.Response)

	go func() {
		v, err := reg.Get(regKeyApi)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to read API key %s from etcd: %s", regKeyApi, err.Error()))
		} else {
			config.ApiKey = v
		}

		go func() {
			for event := range apiKeyWatcher {
				slog.Info("Updating api-key because registry config has changed...")
				config.ApiKey = event.Node.Value
			}
		}()

		reg.Watch(ctx, regKeyApi, false, apiKeyWatcher)
	}()
}

func writeAuthorizedKey(v string, write writeFileFunc) {
	slog.Info(fmt.Sprintf("The authorized ssh public key has changed to %s", v))
	err := write(authorizedKeys, []byte(v), sshKeyFileMode)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not write changed ssh key to file: %s", err.Error()))
	} else {
		slog.Info("Successfully wrote new ssh key to authorized_keys file...")
	}
}

func watchSshKeyConfig(ctx context.Context, reg watchConfigurationContext, write writeFileFunc) {
	sshKeyWatcher := make(chan *client.Response)

	go func() {
		v, err := reg.Get(regKeySsh)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to read public key %s from etcd: %s", regKeySsh, err.Error()))
		} else {
			writeAuthorizedKey(v, write)
		}

		go func() {
			for event := range sshKeyWatcher {
				writeAuthorizedKey(event.Node.Value, write)
			}
		}()

		reg.Watch(ctx, regKeySsh, false, sshKeyWatcher)
	}()
}

func getVolumeIncreaseFactor(reg watchConfigurationContext) float32 {
	volumeIncreaseFactorString, err := reg.Get(regKeyVolumeIncreaseFactor)
	if err != nil || volumeIncreaseFactorString == "" {
		slog.Warn("Could not read volume increase factor from registry. Using default value of 0.3.")
		return defaultVolumeIncreaseFactor
	}

	volumeIncreaseFactorF64, err := strconv.ParseFloat(volumeIncreaseFactorString, 32)
	if err != nil {
		slog.Warn("Could not parse volume increase factor from registry. Using default value of 0.3.")
		return defaultVolumeIncreaseFactor
	}

	return float32(volumeIncreaseFactorF64)
}
