package core

import (
	"fmt"
	"os"
	"strings"
)

const logLevelEnv = "LOG_LEVEL"
const basePathEnv = "BASE_PATH"
const ApiKeyEnv = "API_KEY"
const NamespaceEnv = "NAMESPACE"
const CronJobEnv = "EXPORT_CRON"
const CronJobVerboseEnv = "EXPORT_CRON_VERBOSE"

const errorFormat = "environment variable %s is not set"

type Configuration struct {
	LogLevel    string
	BasePath    string
	ApiKey      string
	Namespace   string
	CronExp     string
	VerboseCron bool
}

func ReadConfigFromEnv() (Configuration, error) {
	conf := Configuration{}

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
	if conf.ApiKey == "" {
		return conf, fmt.Errorf(errorFormat, ApiKeyEnv)
	}

	conf.Namespace = os.Getenv(NamespaceEnv)
	if conf.Namespace == "" {
		return conf, fmt.Errorf(errorFormat, NamespaceEnv)
	}

	conf.CronExp = os.Getenv(CronJobEnv)

	conf.VerboseCron = os.Getenv(CronJobVerboseEnv) == "true"

	return conf, nil
}
