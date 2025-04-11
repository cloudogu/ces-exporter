package configuration

import (
	"encoding/json"
	"fmt"
	"github.com/cloudogu/cesapp-lib/core"
	"log/slog"
	"path"
	"strings"
)

const (
	_DoguPath = "/dogu_v2"
)

type ExcludeOption func(cfg core.ConfigurationField) bool

func ExcludeGlobalConfig() ExcludeOption {
	return func(cfg core.ConfigurationField) bool {
		return cfg.Global
	}
}

func getDoguConfigKeySet(dogu string, exclude ...ExcludeOption) (map[string]bool, error) {
	doguConfigKeySet := make(map[string]bool)

	kvSlice, err := getKeyValues(path.Join(_DoguPath, dogu, "current"))
	if err != nil {
		if isKeyNotFoundError(err) {
			err = ErrDoguNotFound
		}

		return nil, fmt.Errorf("could not read key 'current' for dogu %s: %w", dogu, err)
	}

	dVersion := kvSlice[0].Value

	slog.Debug("Found current version for dogu", "dogu", dogu, "version", dVersion)

	doguJsonSlice, err := getKeyValues(path.Join(_DoguPath, dogu, dVersion))
	if err != nil {
		return nil, fmt.Errorf("could not read dogu json for dogu %s and version %s: %w", dogu, dVersion, err)
	}

	doguJson := doguJsonSlice[0].Value

	var doguSpec core.Dogu

	err = json.Unmarshal([]byte(doguJson), &doguSpec)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal dogu json for %s and version %s: %w", dogu, dVersion, err)
	}

	for _, config := range doguSpec.Configuration {
		doguConfigKeySet[config.Name] = true

		for _, e := range exclude {
			if e(config) {
				doguConfigKeySet[config.Name] = false
			}
		}
	}

	return doguConfigKeySet, nil
}

func GetAllDogus() ([]string, error) {
	kvSlice, err := getKeyValues(_DoguPath)

	if err != nil {
		return nil, fmt.Errorf("could not dogu dir %s: %w", _DoguPath, err)
	}

	dogus := make([]string, 0, len(kvSlice))

	for _, kv := range kvSlice {
		if strings.HasSuffix(kv.Key, "/current") {
			keyComponents := strings.Split(kv.Key, "/")

			dogus = append(dogus, keyComponents[1])
		}
	}

	return dogus, nil
}
