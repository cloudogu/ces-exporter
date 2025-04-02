package export

import (
	"context"
	"fmt"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type configMaps interface {
	corev1.ConfigMapInterface
}

type MultinodeExportModeProvider struct {
	configMaps configMaps
	doguClient ecoSystemV2.EcoSystemV2Interface
	namespace  string
}

func NewMultinodeExportModeProvider(configMaps configMaps, doguClient ecoSystemV2.EcoSystemV2Interface, namespace string) *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{
		configMaps: configMaps,
		doguClient: doguClient,
		namespace:  namespace,
	}
}

func (m MultinodeExportModeProvider) GetExportDogu(ctx context.Context) (*DoguExport, error) {
	return nil, nil
}

func (m MultinodeExportModeProvider) SetExportDogu(doguName string, ctx context.Context) (*DoguExport, error) {
	return nil, nil
}

func (m MultinodeExportModeProvider) GetExportMode(ctx context.Context) (*ExportModeStatus, error) {

	dogus, err := m.doguClient.Dogus(m.namespace).List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("Error getting dogus", "err", err)
	}

	for _, d := range dogus.Items {
		if !d.Spec.ExportMode {
			// if just one dogu is not in export mode, the global export-mode-status is false
			return &ExportModeStatus{IsActive: false}, nil
		}
	}

	// since we did not step out until now - the global export-mode-status is true
	return &ExportModeStatus{IsActive: true}, nil

}
