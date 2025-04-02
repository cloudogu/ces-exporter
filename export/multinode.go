package export

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type kubernetesClient interface {
	kubernetes.Interface
}

type MultinodeExportModeProvider struct {
	client    kubernetesClient
	namespace string
}

func NewMultinodeExportModeProvider(client kubernetesClient, namespace string) *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{
		client:    client,
		namespace: namespace,
	}
}

func (m MultinodeExportModeProvider) GetExportDogu(ctx context.Context) (*DoguExport, error) {
	service, _ := m.client.CoreV1().Services(m.namespace).Get(ctx, "ces-exporter-dogu-exporter", metav1.GetOptions{})
	dogu := service.Spec.Selector["dogu.name"]
	port := service.Spec.Ports[0]

	// TODO Volumepath aus Dogu auslesen
	doguExport := DoguExport{
		Dogu:         dogu,
		VolumePath:   "",
		ExporterPort: int(port.Port),
	}
	return &doguExport, nil
}

func (m MultinodeExportModeProvider) SetExportDogu(doguName string, ctx context.Context) (*DoguExport, error) {
	service, _ := m.client.CoreV1().Services(m.namespace).Get(ctx, "ces-exporter-dogu-exporter", metav1.GetOptions{})
	service.Spec.Selector["dogu.name"] = doguName

	service, err := m.client.CoreV1().Services(m.namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	dogu := service.Spec.Selector["dogu.name"]
	port := service.Spec.Ports[0]
	doguExport := DoguExport{
		Dogu:         dogu,
		VolumePath:   "",
		ExporterPort: int(port.Port),
	}
	return &doguExport, nil
}

func (m MultinodeExportModeProvider) GetExportMode(ctx context.Context) (*ExportModeStatus, error) {
	return nil, nil
}
