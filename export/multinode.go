package export

import (
	"context"
	"fmt"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type configMaps interface {
	corev1.ConfigMapInterface
}
type kubernetesClient interface {
	kubernetes.Interface
}

type MultinodeExportModeProvider struct {
	client     kubernetesClient
	namespace  string
	configMaps configMaps
	doguclient ecoSystemV2.EcoSystemV2Interface
}

func NewMultinodeExportModeProvider(client kubernetesClient, namespace string, configMaps configMaps, doguclient ecoSystemV2.EcoSystemV2Interface) *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{
		client:     client,
		namespace:  namespace,
		configMaps: configMaps,
		doguclient: doguclient,
	}
}

/*
GetExportDogu gets the dogu.name currently set in the ces-exporter-dogu-exporter service
*/
func (m MultinodeExportModeProvider) GetExportDogu(ctx context.Context) (*DoguExport, error) {
	service, _ := m.client.CoreV1().Services(m.namespace).Get(ctx, "ces-exporter-dogu-exporter", metav1.GetOptions{})
	doguName := service.Spec.Selector["dogu.name"]
	port := service.Spec.Ports[0]

	dogu, err := m.doguclient.Dogus(m.namespace).Get(ctx, doguName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get dogu resource for current export dogu: %s", err)
	}

	doguExport := DoguExport{
		Dogu:         doguName,
		VolumePath:   dogu.GetDataVolumeName(),
		ExporterPort: int(port.Port),
	}
	return &doguExport, nil
}

/*
SetExportDogu sets the given dogu as dogu.name in the ces-exporter-dogu-exporter service
*/
func (m MultinodeExportModeProvider) SetExportDogu(doguName string, ctx context.Context) (*DoguExport, error) {
	service, _ := m.client.CoreV1().Services(m.namespace).Get(ctx, "ces-exporter-dogu-exporter", metav1.GetOptions{})
	service.Spec.Selector["dogu.name"] = doguName
	port := service.Spec.Ports[0]

	dogu, err := m.doguclient.Dogus(m.namespace).Get(ctx, doguName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get dogu resource for current export dogu: %s", err)
	}

	_, err = m.client.CoreV1().Services(m.namespace).Update(ctx, service, metav1.UpdateOptions{})
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

	dogus, err := m.doguclient.Dogus(m.namespace).List(ctx, metav1.ListOptions{})

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
