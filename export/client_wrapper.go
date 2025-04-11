package export

import (
	"context"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	v2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DoguClientInterface interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v2.Dogu, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v2.DoguList, error)
	Update(ctx context.Context, dogu *v2.Dogu, opts metav1.UpdateOptions) (*v2.Dogu, error)
}

type EcosystemDoguClient struct {
	Doguclient ecoSystemV2.DoguInterface
}

func NewEcosystemDoguClient(namespace string, client ecoSystemV2.EcoSystemV2Interface) *EcosystemDoguClient {
	dc := client.Dogus(namespace)

	return &EcosystemDoguClient{dc}
}

type ServiceClientInterface interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Service, error)
	Update(ctx context.Context, service *corev1.Service, opts metav1.UpdateOptions) (*corev1.Service, error)
}

type EcosystemServiceClient struct {
	Coreclient *kubernetes.Clientset
}

func NewServiceClient(client *kubernetes.Clientset) *EcosystemServiceClient {
	return &EcosystemServiceClient{Coreclient: client}
}
