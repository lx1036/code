package watcher

import (
	"context"
	"crypto/tls"
	"github.com/bep/debounce"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sync"
	"time"
)

// A Watcher watches for ingresses in the kubernetes cluster
type Watcher struct {
	client   kubernetes.Interface
	onChange func(*Payload)
}

// A Payload is a collection of Kubernetes data loaded by the watcher.
type Payload struct {
	Ingresses       []IngressPayload
	TLSCertificates map[string]*tls.Certificate
}

// An IngressPayload is an ingress + its service ports.
type IngressPayload struct {
	Ingress      *v1beta1.Ingress
	ServicePorts map[string]map[string]int
}

func New(client kubernetes.Interface, onChange func(payload *Payload)) *Watcher {
	return &Watcher{
		client:   client,
		onChange: onChange,
	}
}

func (watcher *Watcher) Run(context context.Context) error {
	factory := informers.NewSharedInformerFactory(watcher.client, time.Minute)
	secretLister := factory.Core().V1().Secrets().Lister()
	serviceLister := factory.Core().V1().Services().Lister()
	ingressLister := factory.Extensions().V1beta1().Ingresses().Lister()

	onChange := func() {

	}

	debounced := debounce.New(time.Second)
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			debounced(onChange)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			debounced(onChange)
		},
		DeleteFunc: func(obj interface{}) {
			debounced(onChange)
		},
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		informer := factory.Core().V1().Secrets().Informer()
		informer.AddEventHandler(handler)
		informer.Run(context.Done())
		waitGroup.Done()
	}()

	waitGroup.Add(1)
	go func() {

	}()

	waitGroup.Add(1)
	go func() {

	}()

	waitGroup.Wait()

	return nil
}
