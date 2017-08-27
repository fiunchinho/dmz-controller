package main

import (
	"fmt"
	"strings"

	"github.com/fiunchinho/dmz-controller/repository"
	"github.com/fiunchinho/dmz-controller/whitelist"
	"github.com/golang/glog"
	"k8s.io/client-go/tools/cache"
)

const (
	// IngressWhitelistAnnotation is the whitelist annotation used by the Kubernetes Ingress
	IngressWhitelistAnnotation = "ingress.kubernetes.io/whitelist-source-range"

	// DMZProvidersAnnotation is the Ingress annotation that contains will trigger this controller
	DMZProvidersAnnotation = "armesto.net/ingress-providers"

	// ManagedWhitelistAnnotation is the name of the internal annotation used to keep track of the CIDRs managed by the controller
	ManagedWhitelistAnnotation = "armesto.net/dmz-controller-managed-cidr"
)

// IngressWhitelister to process watched Ingress objects
type IngressWhitelister struct {
	ingressRepository   repository.IngressRepository
	configMapRepository repository.ConfigMapRepository
}

// Whitelist adds the desired addresses as whitelisted to the given Ingress object
// This is called whenever this controller starts, and whenever the resource changes, and also periodically every resyncPeriod.
// Here we try to reconciliate the current and desired state.
func (whitelister *IngressWhitelister) Whitelist(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("Error splitting meta namespace key into parts: %s", err.Error())
	}

	ingress, err := whitelister.ingressRepository.Get(namespace, name)
	if err != nil {
		return err
	}
	glog.V(0).Infof("Got '%s/%s' Ingress object from cache.", namespace, name)

	configMap, err := whitelister.configMapRepository.Get(namespace, DMZConfigMapName)
	if err != nil {
		return err
	}
	glog.V(1).Infof("Got '%s' ConfigMap from cache, with the following data: %s", DMZConfigMapName, configMap.Data)

	provider, ok := ingress.Annotations[DMZProvidersAnnotation]
	if ok {
		currentWhitelistedIps := whitelist.NewWhitelistFromString(ingress.Annotations[IngressWhitelistAnnotation])
		if _, ok := ingress.Annotations[ManagedWhitelistAnnotation]; ok {
			currentWhitelistedIps.Minus(whitelist.NewWhitelistFromString(ingress.Annotations[ManagedWhitelistAnnotation]))
		}
		whitelistToApply := getWhitelistFromProvider(provider, configMap.Data)
		glog.V(0).Infof("Whitelisting the Ingress object with %s IPs: %s", provider, whitelistToApply.ToString())
		ingress.Annotations[ManagedWhitelistAnnotation] = whitelistToApply.ToString()
		whitelistToApply.Merge(currentWhitelistedIps)
		ingress.Annotations[IngressWhitelistAnnotation] = whitelistToApply.ToString()

		// Once the whitelist has been updated, we will update the resource accordingly.
		// If this request fails, this item will be requeued
		if _, err := whitelister.ingressRepository.Save(ingress); err != nil {
			return err
		}
		glog.V(0).Infof("Saved changes to Ingress resource '%s'", ingress.Name)
	}

	return nil
}

func getWhitelistFromProvider(providers string, whitelistProviders map[string]string) *whitelist.Whitelist {
	whitelistToApply := whitelist.NewEmptyWhitelist()
	for _, value := range strings.Split(providers, ",") {
		provider := strings.TrimSpace(value)
		if _, ok := whitelistProviders[provider]; ok {
			ipsToWhitelist := whitelistProviders[provider]
			providerWhitelist := whitelist.NewWhitelistFromString(ipsToWhitelist)
			whitelistToApply.Merge(providerWhitelist)
		}
	}

	return whitelistToApply
}
