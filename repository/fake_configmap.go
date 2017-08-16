package repository

import (
	"k8s.io/client-go/pkg/api/v1"
)

// Ingress acceses k8s API to fetch/save Ingress objects
type FakeConfigMap struct {
	objects    []v1.ConfigMap
	configMaps map[string]v1.ConfigMap
}

// Get retrieves an ingress object by its name
func (h *FakeConfigMap) Get(namespace string, key string) (*v1.ConfigMap, error) {
	config := h.configMaps[key]
	return &config, nil
}

func (h *FakeConfigMap) Save(configMap *v1.ConfigMap) error {
	h.configMaps[configMap.Name] = *configMap
	return nil
}

func NewFakeConfigMapRepository() ConfigMapRepository {
	return &FakeConfigMap{
		configMaps: make(map[string]v1.ConfigMap),
	}
}
