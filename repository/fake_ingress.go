package repository

import (
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// FakeIngress is an InMemory implementation of an Ingress repository
type FakeIngress struct {
	objects []v1beta1.Ingress
}

// Get retrieves an ingress object by its name
func (h *FakeIngress) Get(namespace string, key string) (*v1beta1.Ingress, error) {
	return &h.objects[0], nil
}

// Save saves the given Ingress object to the k8s API
func (h *FakeIngress) Save(ingress *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	h.objects = append(h.objects, *ingress)
	return ingress, nil
}

// NewFakeIngressRepository returns an instance of the repository
func NewFakeIngressRepository() IngressRepository {
	return &FakeIngress{}
}
