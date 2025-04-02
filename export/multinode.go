package export

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"fmt"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log/slog"
)

type configMaps interface {
	corev1.ConfigMapInterface
}
type kubernetesClient interface {
	kubernetes.Interface
}

type MultinodeExportModeProvider struct {
	client    kubernetesClient
	namespace string
	configMaps configMaps
}

func NewMultinodeExportModeProvider(client kubernetesClient, namespace string, configMaps configMaps) *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{
		client:    client,
		namespace: namespace,
		configMaps: configMaps,
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
