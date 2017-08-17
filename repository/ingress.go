package repository

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// Ingress acceses k8s API to fetch/save Ingress objects
type Ingress struct {
	client          kubernetes.Interface
	informerFactory informers.SharedInformerFactory
}

// Get retrieves an ingress object by its name
func (h *Ingress) Get(namespace string, key string) (*v1beta1.Ingress, error) {
	return h.informerFactory.Extensions().V1beta1().Ingresses().Lister().Ingresses(namespace).Get(key)
}

// Save saves the given Ingress object to the k8s API
func (h *Ingress) Save(ingress *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	return h.client.ExtensionsV1beta1().Ingresses(ingress.Namespace).Update(ingress)
}

// NewIngressRepository returns a repository instance
func NewIngressRepository(client kubernetes.Interface, informerFactory informers.SharedInformerFactory) IngressRepository {
	return &Ingress{
		client:          client,
		informerFactory: informerFactory,
	}
}
