package repository

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestThatIngressesCanBeSavedAndRetrieved(t *testing.T) {
	ingressRepository := NewFakeIngressRepository()
	ingress := &v1beta1.Ingress{}
	ingress.Name = "my-ingress"
	ingress.Namespace = "namespace"

	ingressRepository.Save(ingress)

	fetchedIngress, _ := ingressRepository.Get("namespace", "my-ingress")

	assert := assert.New(t)
	assert.Equal(ingress, fetchedIngress, "The saved Ingress object was not fetched correctly")
}
