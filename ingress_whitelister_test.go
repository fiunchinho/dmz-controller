package main

import (
	"testing"

	"errors"

	"github.com/fiunchinho/dmz-controller/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func TestIpsAreAdded(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "namespace/my-ingress"
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

	NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "4.4.4.4/32", "IP is missing")
	assert.NotContains(ingress.Annotations[IngressWhitelistAnnotation], "1.2.3.4/32", "This IP should not be here")
}

func TestThatNothingChangesWhenTheDMZAnnotationIsNotPresent(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "namespace/my-ingress"
	ingress := &v1beta1.Ingress{}
	ingress.Name = ingressName

	ingressRepository.Save(ingress)

	configMap := &v1.ConfigMap{
		Data: map[string]string{
			"offices": "1.2.3.4/32",
		},
	}
	configMap.Name = DMZConfigMapName
	configMapRepository.Save(configMap)

	NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.NotContains(ingress.Annotations, IngressWhitelistAnnotation, "It's not annotated to be whitelisted")
}

func TestThatAssignIpsWhenThereAreTwoIpSources(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "namespace/my-ingress"
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

	NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "4.4.4.4/32", "IP is missing")
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "1.2.3.4/32", "IP is missing")
}

func TestThatItKeepsExistingWhitelistedIpsNotManagedByTheController(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "namespace/my-ingress"
	ingress := BuildIngressObject().Named(ingressName).WithAnnotation(IngressWhitelistAnnotation, "123.1.2.3/32").WithAnnotation(DMZProvidersAnnotation, "offices").Build()

	ingressRepository.Save(ingress)

	configMap := &v1.ConfigMap{
		Data: map[string]string{
			"offices": "1.2.3.4/32",
		},
	}
	configMap.Name = DMZConfigMapName
	configMapRepository.Save(configMap)

	NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "1.2.3.4/32", "IP is missing")
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "123.1.2.3/32", "IP that was whitelisted before is missing now")
}

func TestUpdatesIpsWhenProviderChanges(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "namespace/my-ingress"
	ingress := BuildIngressObject().Named(ingressName).WithAnnotation(DMZProvidersAnnotation, "vpn,offices").Build()

	ingressRepository.Save(ingress)

	configMap := &v1.ConfigMap{
		Data: map[string]string{
			"offices": "1.2.3.4/32",
			"vpn":     "4.4.4.4/32",
		},
	}
	configMap.Name = DMZConfigMapName
	configMap.Namespace = "namespace"
	configMapRepository.Save(configMap)

	NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)

	configMap.Data = map[string]string{
		"offices": "1.2.3.4/32",
		"vpn":     "8.8.8.8/32",
	}
	configMapRepository.Save(configMap)

	NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.NotContains(ingress.Annotations[IngressWhitelistAnnotation], "4.4.4.4/32", "Old IP was not removed")
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "8.8.8.8/32", "IP is missing")
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "1.2.3.4/32", "IP is missing")
}

func TestThatItSkipsNonExistingProviders(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "namespace/my-ingress"
	ingress := BuildIngressObject().Named(ingressName).WithAnnotation(DMZProvidersAnnotation, "vpn,non-existing,offices").Build()

	ingressRepository.Save(ingress)

	configMap := &v1.ConfigMap{
		Data: map[string]string{
			"offices": "1.2.3.4/32",
			"vpn":     "4.4.4.4/32",
		},
	}
	configMap.Name = DMZConfigMapName
	configMapRepository.Save(configMap)

	NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "4.4.4.4/32", "IP is missing")
	assert.Contains(ingress.Annotations[IngressWhitelistAnnotation], "1.2.3.4/32", "IP is missing")
}

func TestThatItFailsWhenNameHasWrongFormat(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "wrong/name/format"
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

	err := NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)

	assert := assert.New(t)
	assert.Error(err)
	assert.NotContains(ingress.Annotations[IngressWhitelistAnnotation], "4.4.4.4/32", "Shouldn't have any IP")
}

func TestThatReturnsErrorOnIngressRepositoryFailingToFetchObject(t *testing.T) {
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "namespace/my-ingress"

	ingressRepository := new(IngressRepositoryThatFailsToGet)
	ingressRepository.On("Get", "namespace", "my-ingress")

	err := NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)
	assert := assert.New(t)
	assert.Error(err)
}

func TestThatReturnsErrorOnIngressRepositoryFailingToSaveObject(t *testing.T) {
	configMapRepository := repository.NewFakeConfigMapRepository()
	ingressName := "namespace/my-ingress"
	ingress := BuildIngressObject().Named(ingressName).WithAnnotation(DMZProvidersAnnotation, "offices").Build()

	ingressRepository := &IngressRepositoryThatFailsToSave{
		ingressObj: *ingress,
	}
	ingressRepository.mock.On("Get", "namespace", "my-ingress")
	ingressRepository.mock.On("Save", ingress)

	err := NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)
	assert := assert.New(t)
	assert.Error(err)
}

func TestThatReturnsErrorOnConfigMapRepositoryFailing(t *testing.T) {
	ingressRepository := repository.NewFakeIngressRepository()
	ingressName := "namespace/my-ingress"
	ingress := BuildIngressObject().Named(ingressName).WithAnnotation(DMZProvidersAnnotation, "vpn,non-existing,offices").Build()

	ingressRepository.Save(ingress)

	configMapRepository := new(StubConfigMapRepository)
	configMapRepository.On("Get", "namespace", DMZConfigMapName)

	err := NewIngressWhitelister(ingressRepository, configMapRepository).Whitelist(ingressName)
	assert := assert.New(t)
	assert.Error(err)
}

type IngressRepositoryThatFailsToGet struct {
	mock.Mock
}

func (m *IngressRepositoryThatFailsToGet) Get(namespace string, key string) (*v1beta1.Ingress, error) {
	m.Called(namespace, key)
	return nil, errors.New("Failed to fetch Ingress from repository")
}
func (m *IngressRepositoryThatFailsToGet) Save(ingress *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	m.Called(ingress)
	return nil, nil
}

type IngressRepositoryThatFailsToSave struct {
	mock       mock.Mock
	ingressObj v1beta1.Ingress
}

func (m *IngressRepositoryThatFailsToSave) Get(namespace string, key string) (*v1beta1.Ingress, error) {
	m.mock.Called(namespace, key)
	return &m.ingressObj, nil
}
func (m *IngressRepositoryThatFailsToSave) Save(ingress *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	m.mock.Called(ingress)
	return nil, errors.New("Failed to save Ingress in repository")
}

type StubConfigMapRepository struct {
	mock.Mock
}

func (m *StubConfigMapRepository) Get(namespace string, key string) (*v1.ConfigMap, error) {
	m.Called(namespace, key)
	return nil, errors.New("failed")
}
func (m *StubConfigMapRepository) Save(configMap *v1.ConfigMap) (*v1.ConfigMap, error) {
	m.Called(configMap)
	return nil, nil
}

func BuildIngressObject() *IngressBuilder {
	return &IngressBuilder{
		annotations: make(map[string]string),
	}
}

type IngressBuilder struct {
	ingressName string
	annotations map[string]string
}

func (builder *IngressBuilder) Named(name string) *IngressBuilder {
	builder.ingressName = name

	return builder
}

func (builder *IngressBuilder) WithAnnotation(key string, value string) *IngressBuilder {
	builder.annotations[key] = value

	return builder
}

func (builder *IngressBuilder) Build() *v1beta1.Ingress {
	ingress := &v1beta1.Ingress{}
	ingress.Name = builder.ingressName
	ingress.Annotations = builder.annotations

	return ingress
}

func NewIngressWhitelister(ingressRepository repository.IngressRepository, configMapRepository repository.ConfigMapRepository) *IngressWhitelister {
	return &IngressWhitelister{
		ingressRepository:   ingressRepository,
		configMapRepository: configMapRepository,
	}
}
