package systeminfo

import (
	"context"
	"fmt"
	componentEcoClient "github.com/cloudogu/k8s-component-operator/pkg/api/ecosystem"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/repository"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log/slog"
)

type MultinodeSystemInfoProvider struct {
	client    *kubernetes.Clientset
	namespace string
	ecoClient *componentEcoClient.V1Alpha1Client
}

func (m *MultinodeSystemInfoProvider) isMultinode() bool {
	return true
}

func (m *MultinodeSystemInfoProvider) getComponents(ctx context.Context) ([]component, error) {
	slog.Debug("collect components...")
	var components []component
	componentsList, err := m.ecoClient.Components(m.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to get installed components: %w", err)
		}
	}

	for _, c := range componentsList.Items {
		slog.Debug(fmt.Sprintf("found component %s in version %s installed", c.Spec.Name, c.Spec.Version))
		components = append(components, component{
			Name:    c.Spec.Name,
			Version: c.Spec.Version,
		})
	}

	return components, nil
}

func (m *MultinodeSystemInfoProvider) getDogus(ctx context.Context) ([]dogu, error) {
	slog.Debug("collect dogus from local dogu registry...")
	localDoguReg := libdogu.NewDoguVersionRegistry(m.client.CoreV1().ConfigMaps(m.namespace))

	var dogus []dogu
	localDogus, err := localDoguReg.GetCurrentOfAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed dogus: %w", err)
	}

	for _, d := range localDogus {
		slog.Debug(fmt.Sprintf("found dogu %s in version %s in local dogu registry", d.Name.String(), d.Version.String()))
		var size int64
		pvc, err := m.client.CoreV1().PersistentVolumeClaims(m.namespace).Get(context.TODO(), d.Name.String(), metav1.GetOptions{})
		if err != nil {
			slog.Debug(fmt.Sprintf("no pvc found for dogu %s so size is set to 0.", d.Name.String()))
		} else {
			size = pvc.Status.Capacity.Storage().Value()
		}

		dogus = append(dogus, dogu{
			Name:    d.Name.String(),
			Version: d.Version.String(),
			Volume: volume{
				SizeInBytes: size,
			},
		})
	}

	return dogus, nil
}

func (m *MultinodeSystemInfoProvider) getFqdn(ctx context.Context) (string, error) {
	slog.Debug("get fqdn from global registry")
	globalConfigRepo := repository.NewGlobalConfigRepository(m.client.CoreV1().ConfigMaps(m.namespace))

	globalConfig, err := globalConfigRepo.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get global config: %w", err)
	}

	value, exists := globalConfig.Get(globalConfigKeyFqdn)
	if !exists {
		return "", fmt.Errorf("critical error: no fqdn is configured in registry")
	}

	slog.Debug(fmt.Sprintf("found fqdn %s", value))

	return value.String(), nil
}

func NewMultinodeSystemInfoProvider(client *kubernetes.Clientset, namespace string, ecoClient *componentEcoClient.V1Alpha1Client) *MultinodeSystemInfoProvider {
	return &MultinodeSystemInfoProvider{
		client:    client,
		namespace: namespace,
		ecoClient: ecoClient,
	}
}
