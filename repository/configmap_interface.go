package repository

import (
	"k8s.io/client-go/pkg/api/v1"
)

type ConfigMapRepository interface {
	Get(namespace string, key string) (*v1.ConfigMap, error)
	Save(configMap *v1.ConfigMap) error
}
