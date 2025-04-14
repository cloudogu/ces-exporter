package systeminfo

import (
	"errors"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	cesLibCore "github.com/cloudogu/cesapp-lib/core"
)

var (
	errFqdnGlobalConfigKeyNotFound = errors.New("global config key for fqdn not found")
)

type getGlobalConfigFunc func(ignoreKeys []string) (core.GlobalConfig, error)

type doguGetter interface {
	GetAllDogus() ([]string, error)
	GetDoguSpec(dogu string) (cesLibCore.Dogu, error)
}

type SingleNodeSystemInfoProvider struct {
	getGlobalConfig getGlobalConfigFunc
	doguGetter
}

func NewSingleNodeSystemInfoProvider() *SingleNodeSystemInfoProvider {
	return &SingleNodeSystemInfoProvider{}
}

func (sp SingleNodeSystemInfoProvider) getFqdn() (string, error) {
	globalCfg, err := sp.getGlobalConfig(nil)
	if err != nil {
		return "", fmt.Errorf("error getting global config: %w", err)
	}

	for _, kvPair := range globalCfg {
		if kvPair.Key == globalConfigKeyFqdn {
			return kvPair.Value, nil
		}
	}

	return "", errFqdnGlobalConfigKeyNotFound
}

func (sp SingleNodeSystemInfoProvider) getDogus() ([]dogu, error) {
	doguList, err := sp.GetAllDogus()
	if err != nil {
		return nil, fmt.Errorf("error getting currently installed dogus: %w", err)
	}

	doguSystemInfoList := make([]dogu, 0, len(doguList))

	for _, d := range doguList {
		doguSpec, lErr := sp.GetDoguSpec(d)
		if lErr != nil {
			return nil, fmt.Errorf("error getting doguSpec for dogu %s: %w", d, lErr)
		}

		doguSystemInfoList = append(doguSystemInfoList, dogu{
			Name:    doguSpec.Name,
			Version: doguSpec.Version,
			Volume:  volume{SizeInBytes: 0},
		})
	}

	return doguSystemInfoList, nil
}
