package export

import (
	"context"
	"fmt"
	doguv2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	apiCorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log/slog"
	"time"
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

type endpointsClient interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type MultinodeExportModeProvider struct {
	namespace       string
	configMaps      configMaps
	doguclient      doguClient
	serviceclient   serviceClient
	endpointsclient endpointsClient
}

func NewMultinodeExportModeProvider(namespace string, configMaps configMaps, ecosystemDoguClient doguClient, serviceClient serviceClient, endpointsClient endpointsClient) *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{
		namespace:       namespace,
		configMaps:      configMaps,
		doguclient:      ecosystemDoguClient,
		serviceclient:   serviceClient,
		endpointsclient: endpointsClient,
	}
}

// GetExportDogu gets the dogu.name currently set in the ces-exporter-dogu-exporter service
func (m MultinodeExportModeProvider) GetExportDogu(ctx context.Context) (*doguExport, error) {
	service, _ := m.serviceclient.Get(ctx, cesDoguExporter, metav1.GetOptions{})
	doguName := service.Spec.Selector[doguv2.DoguLabelName]
	port := service.Spec.Ports[0]

	dE := &doguExport{
		Dogu:         doguName,
		VolumePath:   dataVolumePath,
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

	if err := m.waitForServiceEndpoints(ctx, service); err != nil {
		return nil, fmt.Errorf("failed to wait for endpoints to update: %w", err)
	}

	dE := &doguExport{
		Dogu:         doguName,
		VolumePath:   dataVolumePath,
		ExporterPort: int(port.Port),
	}
	return dE, nil
}

func (m MultinodeExportModeProvider) waitForServiceEndpoints(ctx context.Context, service *apiCorev1.Service) error {
	slog.Debug("start waiting for endpoints to update", "service", service.Name)
	endpointsWatcher, err := m.endpointsclient.Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", service.Name),
	})
	if err != nil {
		return fmt.Errorf("failed to start watch on endpoints: %w", err)
	}
	defer endpointsWatcher.Stop()

	timeoutCh := time.After(time.Second * 10)
	for {
		select {
		case event := <-endpointsWatcher.ResultChan():
			if event.Type == watch.Error {
				return fmt.Errorf("watch error occurred watching endpoints: %s", event.Object)
			}
			if ep, ok := event.Object.(*apiCorev1.Endpoints); ok {
				// Check if any address is available
				for _, subset := range ep.Subsets {
					if len(subset.Addresses) > 0 {
						slog.Debug("endpoints available", "service", service.Name, "subset", subset.Addresses[0].TargetRef.Name)
						return nil
					}
				}
			}
		case <-timeoutCh:
			return fmt.Errorf("timeout waiting for endpoints to update for service %s", service.Name)
		}
	}
}

func (m MultinodeExportModeProvider) GetExportMode(ctx context.Context) (*exportModeStatus, error) {

	dogus, err := m.doguclient.List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("error getting dogu list: %s", err)
	}

	for _, d := range dogus.Items {
		if !d.Spec.ExportMode {
			// if just one dogu is not in export mode, the global export-mode-status is false
			return &exportModeStatus{IsActive: false}, nil
		}
	}

	// since we did not step out until now - the global export-mode-status is true
	return &exportModeStatus{IsActive: true}, nil

}
