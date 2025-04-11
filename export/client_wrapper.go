package export

import (
	"context"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	v2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type doguClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v2.Dogu, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v2.DoguList, error)
	Update(ctx context.Context, dogu *v2.Dogu, opts metav1.UpdateOptions) (*v2.Dogu, error)
}

type EcosystemDoguClient struct {
	ecoSystemV2.DoguInterface
}

func NewEcosystemDoguClient(namespace string, client ecoSystemV2.EcoSystemV2Interface) *EcosystemDoguClient {
	dc := client.Dogus(namespace)

	return &EcosystemDoguClient{dc}
}

type serviceClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.Service, error)
	Update(ctx context.Context, service *corev1.Service, opts metav1.UpdateOptions) (*corev1.Service, error)
}

type EcosystemServiceClient struct {
	v1.ServiceInterface
}

func NewServiceClient(client v1.ServiceInterface) *EcosystemServiceClient {
	return &EcosystemServiceClient{client}
}
