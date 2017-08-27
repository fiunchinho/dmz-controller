package repository

import (
	"k8s.io/client-go/pkg/api/v1"
)

// ConfigMapRepository is an interface to fetch or store ConfigMaps
type ConfigMapRepository interface {
	Get(namespace string, key string) (*v1.ConfigMap, error)
	Save(configMap *v1.ConfigMap) (*v1.ConfigMap, error)
}
