package repository

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"k8s.io/client-go/pkg/api/v1"
)

func TestThatConfigMapsCanBeSavedAndRetrieved(t *testing.T) {
	configMapRepository := NewFakeConfigMapRepository()
	configMap := &v1.ConfigMap{}
	configMap.Name = "my-cm"
	configMap.Namespace = "namespace"

	configMapRepository.Save(configMap)

	fetchedConfigMap, _ := configMapRepository.Get("namespace", "my-cm")

	assert := assert.New(t)
	assert.Equal(configMap, fetchedConfigMap, "The saved ConfigMap object was not fetched correctly")
}
