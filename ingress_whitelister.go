package main

import (
	"fmt"
	"log"
	"github.com/fiunchinho/dmz-controller/whitelist"
	"strings"
	"github.com/fiunchinho/dmz-controller/repository"
)

const IngressWhitelistAnnotation = "ingress.kubernetes.io/whitelist-source-range"
const DMZProvidersAnnotation = "armesto.net/ingress"

// IngressWhitelister to process watched Ingress objects
type IngressWhitelister struct {
	namespace           string
	ingressRepository   repository.IngressRepository
	configMapRepository repository.ConfigMapRepository
}

func (whitelister *IngressWhitelister) Whitelist(name string) error {
	// retrieve the latest version in the cache of this object
	ingress, err := whitelister.ingressRepository.Get(whitelister.namespace, name)
	if err != nil {
		return fmt.Errorf("error getting object '%s/%s' from api: %s", whitelister.namespace, name, err.Error())
	}
	log.Printf("Got '%s/%s' object from cache.", whitelister.namespace, name)

	configMap, err := whitelister.configMapRepository.Get(whitelister.namespace, DMZConfigMapName)
	if err != nil {
		return fmt.Errorf("error getting dmz configmap: %s", err.Error())
	}
	log.Printf("Got '%s' object from cache.", DMZConfigMapName)

	// This is called whenever this controller starts, and whenever the resource changes, and also periodically every resyncPeriod.
	// Here we try to reconciliate the current and desired state.
	// If there is an error, we skip calling `queue.Forget`, causing the resource to be requeued at a later time.
	log.Printf("Processing Ingress resource '%s'", ingress.Name)
	if ingress.Annotations == nil {
		ingress.Annotations = make(map[string]string)
	}

	provider, ok := ingress.Annotations[DMZProvidersAnnotation]
	if ok {
		currentWhitelistedIps := whitelist.NewWhitelistFromString(ingress.Annotations[IngressWhitelistAnnotation])
		if _, ok := ingress.Annotations["dmz-controller"]; ok {
			currentWhitelistedIps.Minus(whitelist.NewWhitelistFromString(ingress.Annotations["dmz-controller"]))
		}
		whitelistToApply := getWhitelistFromProvider(provider, configMap.Data)
		ingress.Annotations["dmz-controller"] = whitelistToApply.ToString()
		whitelistToApply.Merge(currentWhitelistedIps)
		ingress.Annotations[IngressWhitelistAnnotation] = whitelistToApply.ToString()

		// Once the whitelist has been updated, we will update the resource accordingly.
		// If this request fails, this item will be requeued
		if _, err := whitelister.ingressRepository.Save(ingress); err != nil {
			return fmt.Errorf("error saving update to Ingress resource: %s", err.Error())
		}
		log.Printf("Saved update to Ingress resource '%s'", ingress.Name)
	}

	return nil
}

func getWhitelistFromProvider(provider string, whitelistProviders map[string]string) *whitelist.Whitelist {
	whitelistToApply := whitelist.NewEmptyWhitelist()
	for _, value := range strings.Split(provider, ",") {
		ipsToWhitelist := whitelistProviders[strings.TrimSpace(value)]
		log.Printf("Whitelisting %s", ipsToWhitelist)
		providerWhitelist := whitelist.NewWhitelistFromString(ipsToWhitelist)
		whitelistToApply.Merge(providerWhitelist)
	}

	return whitelistToApply
}
