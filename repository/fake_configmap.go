package repository

import (
	"k8s.io/client-go/pkg/api/v1"
)

// FakeConfigMap is an InMemory implementation of an ConfigMap repository
type FakeConfigMap struct {
	configMaps map[string]v1.ConfigMap
}

// Get retrieves an ingress object by its name
func (h *FakeConfigMap) Get(namespace string, key string) (*v1.ConfigMap, error) {
	config := h.configMaps[key]
	return &config, nil
}

// Save stores the given configmap to the repository
func (h *FakeConfigMap) Save(configMap *v1.ConfigMap) (*v1.ConfigMap, error) {
	h.configMaps[configMap.Name] = *configMap
	return configMap, nil
}

// NewFakeConfigMapRepository returns an instance of the repository
func NewFakeConfigMapRepository() ConfigMapRepository {
	return &FakeConfigMap{
		configMaps: make(map[string]v1.ConfigMap),
	}
}
