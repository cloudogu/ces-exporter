package core

import (
	"fmt"
	"os"
	"strings"
)

const (
	logLevelEnv       = "LOG_LEVEL"
	basePathEnv       = "BASE_PATH"
	NamespaceEnv      = "NAMESPACE"
	ApiKeyEnv         = "API_KEY"
	CronJobVerboseEnv = "EXPORT_CRON_VERBOSE"
	CronJobEnv        = "EXPORT_CRON"
	volumeBasePath    = "VOLUME_BASE"
)

const (
	errorFormat           = "environment variable %s is not set"
	modeClassic           = "classic"
	classicPort           = 7022
	defaultVolumeBasePath = "/data"
)

// Configuration holds all possible configurations of the exporter app
type Configuration struct {
	// LogLevel controls the granularity and amount of issued log output. Valid values are (always in
	// uppercase) `ERROR`, `WARN`, `INFO`, `DEBUG`. Defaults to `ERROR` if left empty.
	LogLevel string
	// BasePath is the context path of the application, usually "/ces-exporter"
	BasePath string
	// ApiKey is the api key which is required to access the api
	ApiKey string
	// Namespace is the namespace in the cluster where the application runs in (only used in multinode ces)
	Namespace string
	// CronExp is the cron expression for the interval to enable the export mode (only used in multinode ces)
	CronExp string
	// VerboseCron defines whether the export mode cronjob should log verbose (only used in multinode ces)
	VerboseCron bool
	// IsClassic defines if the application should start with classic ces configuration or multinode ces configuration
	IsClassic bool
	// Exporter Port in classic mode
	ClassicExportPort int
	// VolumesBasePath defines the base directory path for storing volume-related data.
	VolumesBasePath string
}

func ReadConfigFromEnv() (Configuration, error) {
	conf := Configuration{}

	mode := os.Getenv("MODE")
	conf.IsClassic = mode == modeClassic

	conf.LogLevel = os.Getenv(logLevelEnv)
	if conf.LogLevel == "" {
		conf.LogLevel = "INFO"
	}

	conf.BasePath = os.Getenv(basePathEnv)
	if conf.BasePath == "" {
		conf.BasePath = "/ces-exporter"
	}
	conf.BasePath = strings.TrimSuffix(conf.BasePath, "/")

	conf.ApiKey = os.Getenv(ApiKeyEnv)
	if conf.ApiKey == "" && !conf.IsClassic {
		return conf, fmt.Errorf(errorFormat, ApiKeyEnv)
	}

	conf.Namespace = os.Getenv(NamespaceEnv)
	if conf.Namespace == "" && !conf.IsClassic {
		return conf, fmt.Errorf(errorFormat, NamespaceEnv)
	}

	conf.CronExp = os.Getenv(CronJobEnv)

	conf.VerboseCron = os.Getenv(CronJobVerboseEnv) == "true"

	conf.ClassicExportPort = classicPort

	conf.VolumesBasePath = os.Getenv(volumeBasePath)
	if conf.VolumesBasePath == "" {
		conf.VolumesBasePath = defaultVolumeBasePath
	}

	return conf, nil
}
