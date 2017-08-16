package repository

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/informers"
)

// ConfigMap acceses k8s API to fetch/save ConfigMap objects
type ConfigMap struct {
	client          kubernetes.Interface
	informerFactory informers.SharedInformerFactory
}

// Get retrieves an ingress object by its name
func (h *ConfigMap) Get(namespace string, key string) (*v1.ConfigMap, error) {
	return h.informerFactory.Core().V1().ConfigMaps().Lister().ConfigMaps(namespace).Get(key)
}

func (h *ConfigMap) Save(configMap *v1.ConfigMap) error {
	return nil
}

func NewConfigMapRepository(client kubernetes.Interface, informerFactory informers.SharedInformerFactory) ConfigMapRepository {
	return &ConfigMap{
		client:          client,
		informerFactory: informerFactory,
	}
}
