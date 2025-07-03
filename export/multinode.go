package export

import (
	"context"
	"fmt"
	"path"

	doguv2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	apiCorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	cesDoguExporter = "ces-exporter-dogu-exporter"
)

type configMaps interface {
	corev1.ConfigMapInterface
}

type doguClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*doguv2.Dogu, error)
	List(ctx context.Context, opts metav1.ListOptions) (*doguv2.DoguList, error)
	Update(ctx context.Context, dogu *doguv2.Dogu, opts metav1.UpdateOptions) (*doguv2.Dogu, error)
}

type serviceClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiCorev1.Service, error)
	Update(ctx context.Context, service *apiCorev1.Service, opts metav1.UpdateOptions) (*apiCorev1.Service, error)
}

type MultinodeExportModeProvider struct {
	namespace     string
	configMaps    configMaps
	doguclient    doguClient
	serviceclient serviceClient
}

func NewMultinodeExportModeProvider(namespace string, configMaps configMaps, ecosystemDoguClient doguClient, serviceClient serviceClient) *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{
		namespace:     namespace,
		configMaps:    configMaps,
		doguclient:    ecosystemDoguClient,
		serviceclient: serviceClient,
	}
}

// GetExportDogu gets the dogu.name currently set in the ces-exporter-dogu-exporter service
func (m MultinodeExportModeProvider) GetExportDogu(ctx context.Context) (*doguExport, error) {
	service, _ := m.serviceclient.Get(ctx, cesDoguExporter, metav1.GetOptions{})
	doguName := service.Spec.Selector[doguv2.DoguLabelName]
	port := service.Spec.Ports[0]

	dE := &doguExport{
		Dogu:         doguName,
		VolumePath:   path.Join(dataVolumePath, doguName),
		ExporterPort: int(port.Port),
	}
	return dE, nil
}

// SetExportDogu sets the given dogu as dogu.name in the ces-exporter-dogu-exporter service
func (m MultinodeExportModeProvider) SetExportDogu(ctx context.Context, doguName string) (*doguExport, error) {
	service, _ := m.serviceclient.Get(ctx, cesDoguExporter, metav1.GetOptions{})
	service.Spec.Selector[doguv2.DoguLabelName] = doguName
	port := service.Spec.Ports[0]

	_, err := m.doguclient.Get(ctx, doguName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not get dogu resource for current export dogu: %s", err)
	}

	_, err = m.serviceclient.Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update exporter service: %w", err)
	}

	dE := &doguExport{
		Dogu:         doguName,
		VolumePath:   path.Join(dataVolumePath, doguName),
		ExporterPort: int(port.Port),
	}
	return dE, nil
}

func (m MultinodeExportModeProvider) GetExportMode(ctx context.Context) (*exportModeStatus, error) {
	dogus, err := m.doguclient.List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("error getting dogu list: %s", err)
	}

	for _, d := range dogus.Items {
		if !d.Status.ExportMode || d.Status.Health != doguv2.AvailableHealthStatus {
			// if just one dogu is not in export mode, the global export-mode-status is false
			return &exportModeStatus{IsActive: false}, nil
		}
	}

	// since we did not step out until now - the global export-mode-status is true
	return &exportModeStatus{IsActive: true}, nil
}
