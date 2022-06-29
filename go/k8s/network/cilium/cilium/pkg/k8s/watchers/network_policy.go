package watchers

import (
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (k *K8sWatcher) watchK8sNetworkPolicy(k8sClient kubernetes.Interface) {

	_, networkPolicyController := cache.NewTransformingInformer(
		cache.NewListWatchFromClient(k8sClient.NetworkingV1().RESTClient(),
			"networkpolicies", corev1.NamespaceAll, fields.Everything()),
		&networkingv1.NetworkPolicy{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricNetworkPolicy, metricCreate, valid, equal) }()

				networkPolicy, ok := obj.(*networkingv1.NetworkPolicy)
				if !ok {
					return
				}

				err := k.addK8sNetworkPolicy(networkPolicy, swgSvcs)
				k.K8sEventProcessed(metricNetworkPolicy, metricCreate, err == nil)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricNetworkPolicy, metricUpdate, valid, equal) }()

				oldK8sNetworkPolicy, ok := oldObj.(*networkingv1.NetworkPolicy)
				if !ok {
					return
				}
				newK8sNetworkPolicy, ok := newObj.(*networkingv1.NetworkPolicy)
				if !ok {
					return
				}
				if EqualNetworkPolicy(oldK8sNetworkPolicy, newK8sNetworkPolicy) {
					equal = true
					return
				}

				err := k.updateK8sNetworkPolicy(oldk8sSvc, newk8sSvc, swgSvcs)
				k.K8sEventProcessed(metricNetworkPolicy, metricUpdate, err == nil)
			},
			DeleteFunc: func(obj interface{}) {
				var valid, equal bool
				defer func() { k.K8sEventReceived(metricNetworkPolicy, metricDelete, valid, equal) }()

				k8sSvc := ObjToV1Services(obj)
				if k8sSvc == nil {
					return
				}

				valid = true
				err := k.deleteK8sNetworkPolicy(k8sSvc, swgSvcs)
				k.K8sEventProcessed(metricNetworkPolicy, metricDelete, err == nil)
			},
		},
		nil,
	)

	go networkPolicyController.Run(wait.NeverStop)

}

func (k *K8sWatcher) addK8sNetworkPolicy(k8sNetworkPolicy *networkingv1.NetworkPolicy) error {

}

func (k *K8sWatcher) updateK8sNetworkPolicyV1(oldk8sNetworkPolicy, newk8sNetworkPolicy *networkingv1.NetworkPolicy) error {
	log.WithFields(log.Fields{
		logfields.K8sAPIVersion:                 oldk8sNetworkPolicy.TypeMeta.APIVersion,
		logfields.K8sNetworkPolicyName + ".old": oldk8sNetworkPolicy.ObjectMeta.Name,
		logfields.K8sNamespace + ".old":         oldk8sNetworkPolicy.ObjectMeta.Namespace,
		logfields.K8sNetworkPolicyName:          oldk8sNetworkPolicy.ObjectMeta.Name,
		logfields.K8sNamespace:                  oldk8sNetworkPolicy.ObjectMeta.Namespace,
	}).Debug("Received policy update")

	return k.addK8sNetworkPolicy(newk8sNetworkPolicy)
}

func (k *K8sWatcher) deleteK8sNetworkPolicyV1(k8sNetworkPolicy *networkingv1.NetworkPolicy) error {

}
