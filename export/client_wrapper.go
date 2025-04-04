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
	Get(ctx context.Context, name string) (*v2.Dogu, error)
	List(ctx context.Context) (*v2.DoguList, error)
	Update(ctx context.Context, dogu *v2.Dogu) (*v2.Dogu, error)
}

type DoguClient struct {
	namespace  string
	doguclient ecoSystemV2.EcoSystemV2Interface
}

func NewDoguClient(namespace string, client ecoSystemV2.EcoSystemV2Interface) *DoguClient {
	return &DoguClient{namespace: namespace, doguclient: client}
}

func (dc DoguClient) Get(ctx context.Context, name string) (*v2.Dogu, error) {
	return dc.doguclient.Dogus(dc.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (dc DoguClient) List(ctx context.Context) (*v2.DoguList, error) {
	return dc.doguclient.Dogus(dc.namespace).List(ctx, metav1.ListOptions{})
}

func (dc DoguClient) Update(ctx context.Context, dogu *v2.Dogu) (*v2.Dogu, error) {
	return dc.doguclient.Dogus(dc.namespace).Update(ctx, dogu, metav1.UpdateOptions{})
}

type ServiceClientInterface interface {
	Get(ctx context.Context, name string) (*corev1.Service, error)
	Update(ctx context.Context, service *corev1.Service) (*corev1.Service, error)
}

type ServiceClient struct {
	namespace  string
	coreclient *kubernetes.Clientset
}

func NewServiceClient(namespace string, client *kubernetes.Clientset) *ServiceClient {
	return &ServiceClient{namespace: namespace, coreclient: client}
}

func (dc ServiceClient) Get(ctx context.Context, name string) (*corev1.Service, error) {
	return dc.coreclient.CoreV1().Services(dc.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (dc ServiceClient) Update(ctx context.Context, service *corev1.Service) (*corev1.Service, error) {
	return dc.coreclient.CoreV1().Services(dc.namespace).Update(ctx, service, metav1.UpdateOptions{})
}
