package repository

import (
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// IngressRepository is an interface to fetch or store Ingress
type IngressRepository interface {
	Get(namespace string, key string) (*v1beta1.Ingress, error)
	Save(ingress *v1beta1.Ingress) (*v1beta1.Ingress, error)
}
