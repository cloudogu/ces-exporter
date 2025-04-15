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

type Configuration struct {
	LogLevel    string
	BasePath    string
	ApiKey      string
	Namespace   string
	CronExp     string
	VerboseCron bool
	IsClassic   bool
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
