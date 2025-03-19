package systeminfo

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	componentEcoClient "github.com/cloudogu/k8s-component-operator/pkg/api/ecosystem"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/repository"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log/slog"
	"net/http"
)

const (
	globalConfigKeyFqdn = "fqdn"
)

func handleError(w http.ResponseWriter, err error) {
	slog.Error(err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

type Controller struct {
	isMultinode bool
	client      *kubernetes.Clientset
	namespace   string
	ecoClient   *componentEcoClient.V1Alpha1Client
}

func NewController(client *kubernetes.Clientset, namespace string, ecoClient *componentEcoClient.V1Alpha1Client) *Controller {
	return &Controller{
		isMultinode: true,
		client:      client,
		namespace:   namespace,
		ecoClient:   ecoClient,
	}
}

func (sic Controller) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	globalConfigRepo := repository.NewGlobalConfigRepository(sic.client.CoreV1().ConfigMaps(sic.namespace))

	globalConfig, err := globalConfigRepo.Get(ctx)
	if err != nil {
		handleError(w, fmt.Errorf("failed to get global config: %w", err))
		return
	}

	value, exists := globalConfig.Get(globalConfigKeyFqdn)
	if !exists {
		handleError(w, fmt.Errorf("critical error: no fqdn is configured in registry"))
		return
	}

	localDoguReg := libdogu.NewDoguVersionRegistry(sic.client.CoreV1().ConfigMaps(sic.namespace))

	var dogus []dogu
	localDogus, err := localDoguReg.GetCurrentOfAll(ctx)
	if err != nil {
		handleError(w, fmt.Errorf("failed to get installed dogus: %w", err))
		return
	}

	var components []component
	componentsList, err := sic.ecoClient.Components(sic.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if err != nil {
			handleError(w, fmt.Errorf("failed to get installed components: %w", err))
			return
		}
	}

	for _, c := range componentsList.Items {
		components = append(components, component{
			Name:    c.Spec.Name,
			Version: c.Spec.Version,
		})
	}

	for _, d := range localDogus {
		var size int64
		pvc, err := sic.client.CoreV1().PersistentVolumeClaims(sic.namespace).Get(context.TODO(), d.Name.String(), metav1.GetOptions{})
		if err != nil {
			slog.Info(fmt.Sprintf("no pvc found for dogu %s so size is set to 0.", d.Name.String()))
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

	info := &systemInfo{
		FQDN:        value.String(),
		IsMultinode: sic.isMultinode,
		Dogus:       dogus,
		Components:  components,
	}

	core.JSON(w, http.StatusOK, info)
}
