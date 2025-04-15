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
)

const (
	errorFormat = "environment variable %s is not set"
	modeClassic = "classic"
)

// Configuration holds all possible configurations of the exporter app
type Configuration struct {
	// LogLevel controls the granularity and amount of issued log output. Valid values are (always in
	// uppercase) `ERROR`, `WARN`, `INFO`, `DEBUG`. Defaults to `ERROR` if left empty.
	LogLevel string
	// BasePath is the first part of path on which the api can be reached
	BasePath string
	// ApiKey is the api key which is required to access the api
	ApiKey string
	// Namespace is the namespace in the cluster where the application runs in
	Namespace string
	// CronExp is the cron expression for the interval to enable the export mode
	CronExp string
	// VerboseCron defines whether the export mode cronjob should log verbose
	VerboseCron bool
	// IsClassic defines if the application should start with classic ces configuration or multinode ces configuration
	IsClassic bool
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

	return conf, nil
}
