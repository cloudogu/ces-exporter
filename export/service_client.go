package export

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
