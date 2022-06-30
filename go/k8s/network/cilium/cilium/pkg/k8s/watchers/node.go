package watchers

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"reflect"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/endpoint"
	nodeTypes "k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/node/types"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (k *K8sWatcher) watchK8sNode(k8sClient kubernetes.Interface) {
	_, nodeController := cache.NewTransformingInformer(
		cache.NewListWatchFromClient(k8sClient.CoreV1().RESTClient(),
			"nodes", corev1.NamespaceAll, fields.ParseSelectorOrDie("metadata.name="+nodeTypes.GetName())),
		&corev1.Node{},
		0,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				var valid, equal bool
				if oldNode := ObjToV1Node(oldObj); oldNode != nil {
					valid = true
					if newNode := ObjToV1Node(newObj); newNode != nil {
						oldNodeLabels := oldNode.GetLabels()
						newNodeLabels := newNode.GetLabels()
						if reflect.DeepEqual(oldNodeLabels, newNodeLabels) {
							equal = true
						} else {
							err := k.updateK8sNodeV1(oldNode, newNode)
							k.K8sEventProcessed(metricNode, metricUpdate, err == nil)
						}
					}
				}
				k.K8sEventReceived(metricNode, metricUpdate, valid, equal)
			},
		},
		nil,
	)
	go nodeController.Run(wait.NeverStop)
}

func (k *K8sWatcher) updateK8sNodeV1(oldK8sNode, newK8sNode *corev1.Node) error {
	oldNodeLabels := oldK8sNode.GetLabels()
	newNodeLabels := newK8sNode.GetLabels()

	nodeEP := k.endpointManager.GetHostEndpoint()
	if nodeEP == nil {
		klog.Error("Host endpoint not found")
		return nil
	}

	err := updateEndpointLabels(nodeEP, oldNodeLabels, newNodeLabels)
	if err != nil {
		return err
	}
	return nil
}

func updateEndpointLabels(ep *endpoint.Endpoint, oldLbls, newLbls map[string]string) error {

}

func ObjToV1Node(obj interface{}) *corev1.Node {
	node, ok := obj.(*corev1.Node)
	if ok {
		return node
	}
	deletedObj, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		// Delete was not observed by the watcher but is
		// removed from kube-apiserver. This is the last
		// known state and the object no longer exists.
		node, ok := deletedObj.Obj.(*corev1.Node)
		if ok {
			return node
		}
	}
	return nil
}
