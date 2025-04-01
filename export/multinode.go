package export

import (
	"context"
	"fmt"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log/slog"
)

type configMaps interface {
	corev1.ConfigMapInterface
}

type MultinodeExportModeProvider struct {
	configMaps configMaps
}

func NewMultinodeExportModeProvider(configMaps configMaps) *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{
		configMaps: configMaps,
	}
}

func (m MultinodeExportModeProvider) GetExportDogu(ctx context.Context) (*DoguExport, error) {
	return nil, nil
}

func (m MultinodeExportModeProvider) SetExportDogu(doguName string, ctx context.Context) (*DoguExport, error) {
	return nil, nil
}

func (m MultinodeExportModeProvider) GetExportMode(ctx context.Context) (*ExportModeStatus, error) {

	localDoguReg := libdogu.NewDoguVersionRegistry(m.configMaps)

	localDogus, err := localDoguReg.GetCurrentOfAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error getting dogu list: %s", err)
	}
	for _, element := range localDogus {
		slog.Info(element.Name.String())
	}
	return nil, nil

}
