package core

import (
	"context"
	"github.com/cloudogu/ces-commons-lib/dogu"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type configmaps interface {
	v1.ConfigMapInterface
}

func GetInstalledDogus(ctx context.Context, configMaps configmaps) ([]dogu.SimpleNameVersion, error) {
	localDoguReg := libdogu.NewDoguVersionRegistry(configMaps)
	return localDoguReg.GetCurrentOfAll(ctx)
}
