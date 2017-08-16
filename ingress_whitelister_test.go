package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"github.com/fiunchinho/dmz-controller/repository"
	"k8s.io/client-go/pkg/api/v1"
)

func TestIpsAreAdded(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	irrelevantNamespace := "namespace"
	ingressName := "my-ingress"
	ingress := BuildIngressObject().Named(ingressName).WithAnnotation(DMZProvidersAnnotation, "vpn").Build()

	ingressRepository.Save(ingress)

	configMap := &v1.ConfigMap{
		Data: map[string]string{
			"offices": "1.2.3.4/32",
			"vpn":     "4.4.4.4/32",
		},
	}
	configMap.Name = DMZConfigMapName
	configMapRepository.Save(configMap)

	NewIngressWhitelister(irrelevantNamespace, ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "4.4.4.4/32", "IP is missing")
	assert.NotContains(ingress.Annotations[IngressWhitelistAnnotation], "1.2.3.4/32", "This IP should not be here")
}

func TestThatAssignIpsWhenThereAreTwoIpSources(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	irrelevantNamespace := "namespace"
	ingressName := "my-ingress"
	ingress := BuildIngressObject().Named(ingressName).WithAnnotation(DMZProvidersAnnotation, "vpn,offices").Build()

	ingressRepository.Save(ingress)

	configMap := &v1.ConfigMap{
		Data: map[string]string{
			"offices": "1.2.3.4/32",
			"vpn":     "4.4.4.4/32",
		},
	}
	configMap.Name = DMZConfigMapName
	configMapRepository.Save(configMap)

	NewIngressWhitelister(irrelevantNamespace, ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "4.4.4.4/32", "IP is missing")
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "1.2.3.4/32", "IP is missing")
}

func TestUpdatesIpsWhenProviderChanges(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	irrelevantNamespace := "namespace"
	ingressName := "my-ingress"
	ingress := BuildIngressObject().Named(ingressName).WithAnnotation(DMZProvidersAnnotation, "vpn,offices").Build()

	ingressRepository.Save(ingress)

	configMap := &v1.ConfigMap{
		Data: map[string]string{
			"offices": "1.2.3.4/32",
			"vpn":     "4.4.4.4/32",
		},
	}
	configMap.Name = DMZConfigMapName
	configMap.Namespace = irrelevantNamespace
	configMapRepository.Save(configMap)

	NewIngressWhitelister(irrelevantNamespace, ingressRepository, configMapRepository).Whitelist(ingressName)

	configMap.Data = map[string]string{
		"offices": "1.2.3.4/32",
		"vpn":     "8.8.8.8/32",
	}
	configMapRepository.Save(configMap)

	NewIngressWhitelister(irrelevantNamespace, ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.NotContains(ingress.Annotations[IngressWhitelistAnnotation], "4.4.4.4/32", "Old IP was not removed")
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "8.8.8.8/32", "IP is missing")
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "1.2.3.4/32", "IP is missing")
}

func BuildIngressObject() *IngressBuilder {
	return &IngressBuilder{}
}

type IngressBuilder struct {
	ingressName     string
	annotationKey   string
	annotationValue string
}

func (builder *IngressBuilder) Named(name string) (*IngressBuilder) {
	builder.ingressName = name

	return builder
}

func (builder *IngressBuilder) WithAnnotation(key string, value string) (*IngressBuilder) {
	builder.annotationKey = key
	builder.annotationValue = value

	return builder
}

func (builder *IngressBuilder) Build() (*v1beta1.Ingress) {
	ingress := &v1beta1.Ingress{}
	ingress.Name = builder.ingressName
	ingress.Annotations = make(map[string]string)
	ingress.Annotations[builder.annotationKey] = builder.annotationValue

	return ingress
}

func NewIngressWhitelister(namespace string, ingressRepository repository.IngressRepository, configMapRepository repository.ConfigMapRepository) *IngressWhitelister {
	return &IngressWhitelister{
		namespace:           namespace,
		ingressRepository:   ingressRepository,
		configMapRepository: configMapRepository,
	}
}
