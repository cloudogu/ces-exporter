package export

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	client     kubernetesClient
	namespace  string
	configMaps configMaps
	doguclient ecoSystem.EcoSystemV2Interface
}

func NewMultinodeExportModeProvider(client kubernetesClient, namespace string, configMaps configMaps, doguclient ecoSystem.EcoSystemV2Interface) *MultinodeExportModeProvider {
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

	service, err := m.client.CoreV1().Services(m.namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

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

func (m MultinodeExportModeProvider) GetExportMode(ctx context.Context) (*ExportModeStatus, error) {

	localDoguReg := libdogu.NewDoguVersionRegistry(m.configMaps)

	localDogus, err := localDoguReg.GetCurrentOfAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting dogu list: %s", err)
	}
	for _, element := range localDogus {
		slog.Info(element.Name.String())
	}
	return nil, nil

}
