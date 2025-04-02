package core

import (
	"fmt"
	"os"
	"strings"
)

const (
	modeClassic = "classic"
)

const logLevelEnv = "LOG_LEVEL"
const basePathEnv = "BASE_PATH"
const apiKeyEnv = "API_KEY"
const namespaceEnv = "NAMESPACE"

const errorFormat = "environment variable %s is not set"

type Configuration struct {
	LogLevel  string
	BasePath  string
	ApiKey    string
	Namespace string
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

	conf.ApiKey = os.Getenv(apiKeyEnv)
	if conf.ApiKey == "" {
		return conf, fmt.Errorf(errorFormat, apiKeyEnv)
	}

	conf.Namespace = os.Getenv(namespaceEnv)
	if conf.Namespace == "" && !conf.IsClassic {
		return conf, fmt.Errorf(errorFormat, namespaceEnv)
	}

	return conf, nil
}
