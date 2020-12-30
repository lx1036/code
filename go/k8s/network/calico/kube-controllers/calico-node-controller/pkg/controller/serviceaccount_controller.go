package controller

import (
	"fmt"
	"time"
	
	"k8s-lx1036/k8s/network/calico/kube-controllers/calico-node-controller/pkg/calico"
	"k8s-lx1036/k8s/network/calico/kube-controllers/calico-node-controller/pkg/converter"
	"k8s-lx1036/k8s/network/calico/kube-controllers/calico-node-controller/pkg/kube"
	
	log "github.com/sirupsen/logrus"
	
	
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func NewServiceAccountController()  {
	
	
	
	kubeClientset := kube.GetKubernetesClientset()
	calicoClient := calico.GetCalicoClientOrDie()
	
	factory := informers.NewSharedInformerFactory(kubeClientset, time.Minute*10)
	
	serviceAccountConverter := converter.NewServiceAccountConverter()
	
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
