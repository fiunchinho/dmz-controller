package main

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"io/ioutil"
	"os"
	"strings"

	"github.com/fiunchinho/dmz-controller/repository"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// DMZConfigMapName is an internal annotation used to store addreses whitelisted by this controller
const DMZConfigMapName = "dmz-controller"

var (
	namespace string
	// queue is a queue of resources to be processed.
	// It performs exponential backoff rate limiting, with a minimum retry period of 5 seconds and a maximum of 1 minute.
	queue = workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*15, time.Minute))

	// stopCh can be used to stop all the informer, as well as control loops within the application.
	stopCh = make(chan struct{})

	// sharedFactory is a shared informer factory that is used as a cache for
	// items in the API server. It saves each informer listing and watching the
	// same resources independently of each other, thus providing more up to
	// date results with less 'effort'
	sharedFactory informers.SharedInformerFactory

	// client is a Kubernetes API client for our custom resource definition type
	client kubernetes.Interface
)

func getNamespace() string {
	if ns := os.Getenv("TILLER_NAMESPACE"); ns != "" {
		return ns
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}

	return ""
}

func main() {
	// When running as a pod in-cluster, a kubeconfig is not needed. Instead this will make use of the service account injected into the pod.
	// However, allow the use of a local kubeconfig as this can make local development & testing easier.
	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig file")

	// We log to stderr because glog will default to logging to a file.
	// By setting this debugging is easier via `kubectl logs`
	flag.Set("logtostderr", "true")
	flag.Parse()

	namespace = getNamespace()
	if namespace == "" {
		glog.Fatalf("The NAMESPACE environment variable is not set, and the file /var/run/secrets/kubernetes.io/serviceaccount/namespace can't be read")
	}

	// Build the client config - optionally using a provided kubeconfig file.
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		glog.Fatalf("Failed to load client config: %v", err)
	}

	// Construct the Kubernetes client
	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating kubernetes client: %s", err.Error())
	}

	log.Printf("Created Kubernetes client.")

	// we use a shared informer from the informer factory, to save calls to the API as we grow our application
	// and so state is consistent between our control loops.
	// We set a resync period of 30 seconds, in case any create/replace/update/delete operations are missed when watching
	sharedFactory = informers.NewSharedInformerFactory(client, time.Second*30)
	informer := sharedFactory.Extensions().V1beta1().Ingresses().Informer()
	cmInformer := sharedFactory.Core().V1().ConfigMaps().Informer()

	// we add a new event handler, watching for changes to API resources.
	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: enqueue,
			UpdateFunc: func(old, cur interface{}) {
				if !reflect.DeepEqual(old, cur) {
					enqueue(cur)
				}
			},
		},
	)
	cmInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(old, cur interface{}) {
				if cur.(*v1.ConfigMap).Name == DMZConfigMapName {
					if !reflect.DeepEqual(old, cur) {
						ingresses, err := sharedFactory.Extensions().V1beta1().Ingresses().Lister().Ingresses("default").List(labels.Everything())
						if err != nil {
							log.Fatalf("error listing ingresses to notify change on configmap: %s", err.Error())
						}
						for _, ingress := range ingresses {
							log.Printf("Queuing ingress %s object so it gets notified of the ConfigMap change", ingress.Name)
							enqueue(ingress)
						}
					}
				}
			},
		},
	)

	// start the informer. This will cause it to begin receiving updates from the configured API server and firing event handlers in response.
	sharedFactory.Start(stopCh)
	log.Printf("Started informer factory.")

	// wait for the informer cache to finish performing it's initial applyWhiteList of resources
	if !cache.WaitForCacheSync(stopCh, cmInformer.HasSynced, informer.HasSynced) {
		log.Fatalf("error waiting for informer cache to applyWhiteList: %s", err.Error())
	}

	ingressWhitelister := IngressWhitelister{
		namespace:           namespace,
		ingressRepository:   repository.NewIngressRepository(client, sharedFactory),
		configMapRepository: repository.NewConfigMapRepository(client, sharedFactory),
	}

	log.Printf("Finished populating shared informers cache.")

	// here we start reading objects off the queue
	for {
		// we read a message off the queue
		key, shutdown := queue.Get()

		// if the queue has been shut down, we should exit the work queue here
		if shutdown {
			stopCh <- struct{}{}
			return
		}

		// convert the queue item into a string. If it's not a string, we'll simply discard it as invalid data and log a message.
		var strKey string
		var ok bool
		if strKey, ok = key.(string); !ok {
			runtime.HandleError(fmt.Errorf("key in queue should be of type string but got %T. discarding", key))
			return
		}

		// we define a function here to process a queue item, so that we can use 'defer' to make sure the message is marked as Done on the queue
		// Done marks item as done processing, and if it has been marked as dirty again while it was being processed,
		// it will be re-added to the queue for re-processing.
		func(key string) {
			defer queue.Done(key)

			// attempt to split the 'key' into namespace and object name
			_, name, err := cache.SplitMetaNamespaceKey(strKey)
			if err != nil {
				runtime.HandleError(fmt.Errorf("error splitting meta namespace key into parts: %s", err.Error()))
				return
			}

			log.Printf("Read key '%s/%s' off workqueue. Fetching from cache...", namespace, name)

			err = ingressWhitelister.Whitelist(name)

			if err != nil {
				runtime.HandleError(fmt.Errorf("error getting object '%s/%s' from api: %s", namespace, name, err.Error()))
				return
			}

			// as we managed to process this successfully, we can forget it
			// from the work queue altogether.
			// Forget indicates that an item is finished being retried.  Doesn't matter whether its for perm failing
			// or for success, we'll stop the rate limiter from tracking it.  This only clears the `rateLimiter`, you
			// still have to call `Done` on the queue.
			log.Printf("Finished processing '%s/%s' successfully! Removing from queue.", namespace, name)
			queue.Forget(key)
		}(strKey)
	}
}

// enqueue will add an object 'obj' into the workqueue. The object being added
// must be of type metav1.Object, metav1.ObjectAccessor or cache.ExplicitKey.
func enqueue(obj interface{}) {
	// DeletionHandlingMetaNamespaceKeyFunc will convert an object into a
	// 'namespace/name' string. We do this because our item may be processed
	// much later than now, and so we want to ensure it gets a fresh copy of
	// the resource when it starts. Also, this allows us to keep adding the
	// same item into the work queue without duplicates building up.
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(fmt.Errorf("error obtaining key for object being enqueue: %s", err.Error()))
		return
	}
	// add the item to the queue
	queue.Add(key)
}
