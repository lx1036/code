package serviceaccount

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"time"

	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/calico"
	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/converter"
	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/kube"

	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func NewServiceAccountController() {
	kubeClientset := kube.GetKubernetesClientset()
	calicoClient := calico.GetCalicoClientOrDie()

	factory := informers.NewSharedInformerFactory(kubeClientset, time.Minute*10)

	serviceAccountConverter := converter.NewServiceAccountConverter()
	
	
	
	
	listWatcher := cache.NewListWatchFromClient(kubeClientset.CoreV1().RESTClient(), "serviceaccounts", metav1.NamespaceAll, fields.Everything())
	_, informer := cache.NewIndexerInformer(listWatcher, &v1.ServiceAccount{}, time.Minute * 2, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Debugf("Got ADD event for ServiceAccount: %#v", obj)
			profile, err := serviceAccountConverter.Convert(obj)
			if err != nil {
				log.WithError(err).Errorf("Error while converting %#v to Calico profile.", obj)
				return
			}
			
			// Add to cache.
			k := serviceAccountConverter.GetKey(profile)
			ccache.Set(k, profile)
		},
		UpdateFunc: nil,
		DeleteFunc: nil,
	}, cache.Indexers{})
	

	factory.Core().V1().ServiceAccounts().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			profile, err := serviceAccountConverter.Convert(obj)
			if err != nil {
				log.WithError(err).Errorf("Error while converting %#v to Calico profile.", obj)
				return
			}

			key := serviceAccountConverter.GetKey(profile)

		},
		UpdateFunc: nil,
		DeleteFunc: nil,
	})

}
