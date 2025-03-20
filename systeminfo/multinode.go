package systeminfo

import (
	"context"
	"fmt"
	v1 "github.com/cloudogu/k8s-component-operator/pkg/api/v1"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/repository"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log/slog"
)

type configMaps interface {
	corev1.ConfigMapInterface
}

type pvcClaim interface {
	corev1.PersistentVolumeClaimInterface
}

type componentLister interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1.ComponentList, error)
}

type MultinodeSystemInfoProvider struct {
	configMaps      configMaps
	pvc             pvcClaim
	namespace       string
	componentLister componentLister
}

func NewMultinodeSystemInfoProvider(configMaps configMaps, pvc pvcClaim, namespace string, componentLister componentLister) *MultinodeSystemInfoProvider {
	return &MultinodeSystemInfoProvider{
		configMaps:      configMaps,
		pvc:             pvc,
		namespace:       namespace,
		componentLister: componentLister,
	}
}

func (m *MultinodeSystemInfoProvider) isMultinode() bool {
	return true
}

func (m *MultinodeSystemInfoProvider) getComponents(ctx context.Context) ([]component, error) {
	slog.Debug("collect components...")
	var components []component
	componentsList, err := m.componentLister.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get installed components: %w", err)
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
	localDoguReg := libdogu.NewDoguVersionRegistry(m.configMaps)

	var dogus []dogu
	localDogus, err := localDoguReg.GetCurrentOfAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed dogus: %w", err)
	}

	for _, d := range localDogus {
		slog.Debug(fmt.Sprintf("found dogu %s in version %s in local dogu registry", d.Name.String(), d.Version.String()))
		var size int64
		pvc, err := m.pvc.Get(context.TODO(), d.Name.String(), metav1.GetOptions{})
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
	globalConfigRepo := repository.NewGlobalConfigRepository(m.configMaps)

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
