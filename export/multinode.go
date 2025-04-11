package export

import (
	"context"
	"fmt"
	doguv2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	cesDoguExporter = "ces-exporter-dogu-exporter"
)

type configMaps interface {
	corev1.ConfigMapInterface
}

type MultinodeExportModeProvider struct {
	namespace     string
	configMaps    configMaps
	doguclient    DoguClientInterface
	serviceclient ServiceClientInterface
}

func NewMultinodeExportModeProvider(namespace string, configMaps configMaps, ecosystemDoguClient DoguClientInterface, serviceClient ServiceClientInterface) *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{
		namespace:     namespace,
		configMaps:    configMaps,
		doguclient:    ecosystemDoguClient,
		serviceclient: serviceClient,
	}
}

/*
GetExportDogu gets the dogu.name currently set in the ces-exporter-dogu-exporter service
*/
func (m MultinodeExportModeProvider) GetExportDogu(ctx context.Context) (*DoguExport, error) {
	service, _ := m.serviceclient.Get(ctx, cesDoguExporter, metav1.GetOptions{})
	doguName := service.Spec.Selector[doguv2.DoguLabelName]
	port := service.Spec.Ports[0]

	dogu, err := m.doguclient.Get(ctx, doguName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get dogu resource for current export dogu: %s", err)
	}

	doguExport := DoguExport{
		Dogu:         doguName,
		VolumePath:   "/data/" + dogu.GetDataVolumeName(),
		ExporterPort: int(port.Port),
	}
	return &doguExport, nil
}

/*
SetExportDogu sets the given dogu as dogu.name in the ces-exporter-dogu-exporter service
*/
func (m MultinodeExportModeProvider) SetExportDogu(doguName string, ctx context.Context) (*DoguExport, error) {
	service, _ := m.serviceclient.Get(ctx, cesDoguExporter, metav1.GetOptions{})
	service.Spec.Selector[doguv2.DoguLabelName] = doguName
	port := service.Spec.Ports[0]

	dogu, err := m.doguclient.Get(ctx, doguName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get dogu resource for current export dogu: %s", err)
	}

	_, err = m.serviceclient.Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	doguExport := DoguExport{
		Dogu:         doguName,
		VolumePath:   dogu.GetDataVolumeName(),
		ExporterPort: int(port.Port),
	}
	return &doguExport, nil
}

func (m MultinodeExportModeProvider) GetExportMode(ctx context.Context) (*ExportModeStatus, error) {

	dogus, err := m.doguclient.List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("error getting dogu list: %s", err)
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
