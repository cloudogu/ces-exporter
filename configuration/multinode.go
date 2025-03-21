package configuration

import v1 "k8s.io/client-go/kubernetes/typed/core/v1"

type MultinodeSystemInfoProvider struct {
	namespace  string
	configMaps v1.ConfigMapInterface
	secrets    v1.SecretInterface
}

func NewMultinodeSystemInfoProvider(namespace string, configMaps v1.ConfigMapInterface, secrets v1.SecretInterface) *MultinodeSystemInfoProvider {
	return &MultinodeSystemInfoProvider{
		namespace:  namespace,
		configMaps: configMaps,
		secrets:    secrets,
	}
}
