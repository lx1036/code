package node

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/calico"
	"k8s-lx1036/k8s/network/calico/kube-controllers/pkg/kube"

	log "github.com/sirupsen/logrus"

	apisv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	bapi "github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	client "github.com/projectcalico/libcalico-go/lib/clientv3"
	"github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/ipam"
	"github.com/projectcalico/libcalico-go/lib/options"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var (
	accessor = meta.NewAccessor()
)

type NodeController struct {
	ctx             context.Context
	informerFactory informers.SharedInformerFactory

	queue workqueue.RateLimitingInterface

	calicoClient  client.Interface
	kubeClientset *kubernetes.Clientset
}

func NewNodeController() *NodeController {
	kubeClientset := kube.GetKubernetesClientset()
	calicoClient := calico.GetCalicoClientOrDie()

	factory := informers.NewSharedInformerFactory(kubeClientset, time.Minute*2)

	nodeController := &NodeController{
		informerFactory: factory,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "node"),
		calicoClient:    calicoClient,
		kubeClientset:   kubeClientset,
		ctx:             context.TODO(),
	}

	nodeInformer := factory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			nodeController.AddNode(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			nodeController.UpdateNode(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			nodeController.DeleteNode(obj)
		},
	})

	return nodeController
}

type Action string

const (
	Add    Action = "add"
	Update Action = "update"
	Delete Action = "delete"
)

type item struct {
	object interface{}
	key    string
	action Action
}

// node对象会有心跳检查status.conditions，貌似是5mins一次，每次resourceVersion都会变化，这样每次都会是个新的对象
func (controller *NodeController) UpdateNode(oldObj, newObj interface{}) {
	o, _ := accessor.ResourceVersion(oldObj.(runtime.Object))
	n, _ := accessor.ResourceVersion(newObj.(runtime.Object))
	// 只有resource version不同才是新对象
	if o != n {
		controller.Enqueue(&item{
			object: newObj,
			action: Update,
		})
	}
}

func (controller *NodeController) AddNode(obj interface{}) {

	controller.Enqueue(&item{
		object: obj,
		action: Add,
	})
}

func (controller *NodeController) DeleteNode(obj interface{}) {
	controller.Enqueue(&item{
		object: obj,
		action: Delete,
	})
}

func (controller *NodeController) Enqueue(item *item) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(item.object)
	if err != nil {
		log.Errorf("fail to get key for %v", item.object)
		return
	}

	// filter pod without log configuration
	_, ok := item.object.(*corev1.Node)
	if !ok {
		log.Errorf("expected *corev1.Node but got %T", item.object)
		return
	}

	// only queue ready pod, 同一个pod在启动过程中，会出现多种状态，直至最后status.conditions都是ready状态
	// 但是会触发多次的update event
	//for _, condition := range node.Status.Conditions {
	//	if condition.Type == corev1.NodeCondition{}PodReady && condition.Status != coreV1.ConditionTrue {
	//		return
	//	}
	//}

	item.key = key
	defer log.Infof("%s node %s into the queue", item.action, item.key)

	controller.queue.Add(item)
}

func (controller *NodeController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	go controller.informerFactory.Start(stopCh)

	log.Infof("Starting node controller")

	if !cache.WaitForNamedCacheSync("node", stopCh,
		controller.informerFactory.Core().V1().Nodes().Informer().HasSynced) {
		return fmt.Errorf("kubernetes informer is unable to sync cache")
	}

	log.Info("Starting workers of node controller")
	for i := 0; i < workers; i++ {
		go wait.Until(func() {
			for controller.process() {
			}
		}, time.Second, stopCh)
	}

	return nil
}

func (controller *NodeController) process() bool {
	key, quit := controller.queue.Get()
	if quit {
		return false
	}
	defer controller.queue.Done(key)

	err := controller.syncIPAM()
	if err != nil {
		return false
	}

	err = controller.syncDeleteEtcd()
	if err != nil {
		return false
	}

	return true
}

// syncDelete is the main work routine of the controller. It queries Calico and
// K8s, and deletes any Calico nodes which do not exist in K8s.
func (controller *NodeController) syncDeleteEtcd() error {
	calicoNodes, err := controller.calicoClient.Nodes().List(controller.ctx, options.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Error listing Calico nodes")
		return err
	}

	k8sNodes, err := controller.kubeClientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Error listing K8s nodes")
		return err
	}

	var k8sNodeNames []string
	for _, k8sNode := range k8sNodes.Items {
		k8sNodeNames = append(k8sNodeNames, k8sNode.Name)
	}

	k8sNodeNamesSet := sets.NewString(k8sNodeNames...)

	for _, calicoNode := range calicoNodes.Items {
		k8sNodeName := toK8sNodeName(calicoNode)
		if len(k8sNodeName) != 0 && !k8sNodeNamesSet.Has(k8sNodeName) { // 在calico nodes里但不在k8s nodes集群里
			// 从calico中删除该node相关所有资源
			_, err := controller.calicoClient.Nodes().Delete(controller.ctx, calicoNode.Name, options.DeleteOptions{})
			if err != nil {
				_, notExist := err.(errors.ErrorResourceDoesNotExist)
				if notExist {
					log.Warnf("calico node resource does not exist with: %v", err)
				} else {
					log.WithError(err).Error("error to delete calico node")
					return err
				}
			}
		}
	}

	return nil
}

func toK8sNodeName(calicoNode apisv3.Node) string {
	for _, ref := range calicoNode.Spec.OrchRefs {
		if ref.Orchestrator == "k8s" {
			return ref.NodeName
		}
	}

	return ""
}

func (controller *NodeController) syncIPAM() error {
	type accessor interface {
		Backend() bapi.Client
	}
	backendClient := controller.calicoClient.(accessor).Backend()
	blocks, err := backendClient.List(controller.ctx, model.BlockListOptions{}, "")
	if err != nil {
		log.Errorf("list ip blocks with err: %v", err)
		return err
	}

	// Build a list of all the nodes in the cluster based on IPAM allocations across all
	// blocks, plus affinities. Entries are Calico node names.
	calicoNodes := map[string][]model.AllocationAttribute{}
	for _, kvp := range blocks.KVPairs {
		block := kvp.Value.(*model.AllocationBlock)
		log.Infof("block %s", block.CIDR.String())

		// Include affinity if it exists. We want to track nodes even
		// if there are no IPs actually assigned to that node.
		if block.Affinity != nil {
			n := strings.TrimPrefix(*block.Affinity, "host:")
			if _, ok := calicoNodes[n]; !ok {
				calicoNodes[n] = []model.AllocationAttribute{}
			}
		}

		// Go through each IPAM allocation, check its attributes for the node it is assigned to.
		for _, idx := range block.Allocations {
			if idx == nil {
				// Not allocated.
				continue
			}
			attr := block.Attributes[*idx]

			// Track nodes based on IP allocations.
			if val, ok := attr.AttrSecondary[ipam.AttributeNode]; ok {
				if _, ok := calicoNodes[val]; !ok {
					calicoNodes[val] = []model.AllocationAttribute{}
				}

				// If there is no handle, then skip this IP. We need the handle
				// in order to release the IP below.
				if attr.AttrPrimary == nil {
					log.WithFields(log.Fields{"ip": calico.OrdinalToIP(block, *idx), "block": block.CIDR.String()}).
						Debugf("Skipping IP with no handle")
					continue
				}

				// Add this allocation to the node, so we can release it later if
				// we need to.
				calicoNodes[val] = append(calicoNodes[val], attr)
			}
		}
	}

	log.Debugf("Calico nodes found in IPAM: %v", calicoNodes)

	// For each node present in IPAM, if it doesn't exist in the Kubernetes API then we
	// should consider it a candidate for cleanup.
	/*for cnode, allocations := range calicoNodes {

	}*/

	return nil
}
