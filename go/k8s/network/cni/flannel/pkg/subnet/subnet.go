package subnet

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type Manager interface {
	GetNetworkConfig(ctx context.Context) (*Config, error)
	AcquireLease(ctx context.Context, attrs *LeaseAttrs) (*Lease, error)
	RenewLease(ctx context.Context, lease *Lease) error
	WatchLease(ctx context.Context, sn ip.IP4Net, cursor interface{}) (LeaseWatchResult, error)
	WatchLeases(ctx context.Context) (LeaseWatchResult, error)

	Name() string
}

type EventType int

const (
	EventAdded EventType = iota
	EventRemoved
)

type Event struct {
	Type  EventType `json:"type"`
	Lease Lease     `json:"lease,omitempty"`
}

type kubeSubnetManager struct {
	enableIPv4 bool
	subnetConf *Config
	nodeName   string

	kubeClient kubernetes.Interface
	nodeStore  listersv1.NodeLister
	controller cache.Controller

	annotations               annotations
	nodeController            cache.Controller
	setNodeNetworkUnavailable bool

	events chan Event
}

// NewSubnetManager netConfPath=/etc/kube-flannel/net-conf.json
// net-conf.json: |
//    {
//      "Network": "10.244.0.0/16",
//      "Backend": {
//        "Type": "vxlan"
//      }
//    }
func NewSubnetManager(ctx context.Context, kubeConfig, prefix,
	netConfPath string, setNodeNetworkUnavailable bool) (Manager, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("fail to create kubernetes config: %v", err)
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize client: %v", err)
	}

	nodeName, err := getNodeName(kubeClient)
	if err != nil {
		return nil, err
	}

	config, err := getSubnetConfig(netConfPath)
	if err != nil {
		return nil, err
	}

	// watch K8s node
	subnetManager, err := newKubeSubnetManager(ctx, kubeClient, config, nodeName, prefix, setNodeNetworkUnavailable)
	if err != nil {
		return nil, fmt.Errorf("error creating network manager: %s", err)
	}
	go subnetManager.Run(ctx)
	syncCh := make(chan struct{})
	go func() {
		cache.WaitForCacheSync(ctx.Done(), subnetManager.controller.HasSynced)
		close(syncCh)
	}()
	select {
	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("cache sync timeout")
	case <-syncCh:
	}
	klog.Infof("Node controller sync successful")

	return subnetManager, nil
}

func newKubeSubnetManager(ctx context.Context, kubeClient kubernetes.Interface,
	config *Config, nodeName, prefix string, setNodeNetworkUnavailable bool) (*kubeSubnetManager, error) {
	annotation, err := newAnnotations(prefix)
	if err != nil {
		klog.Fatal(err)
	}
	subnetMgr := &kubeSubnetManager{
		enableIPv4:                config.EnableIPv4,
		kubeClient:                kubeClient,
		subnetConf:                config,
		nodeName:                  nodeName,
		annotations:               annotation,
		events:                    make(chan Event, 5000),
		setNodeNetworkUnavailable: setNodeNetworkUnavailable,
	}

	store, controller := cache.NewTransformingInformer(
		cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(),
			"nodes", corev1.NamespaceAll, fields.Everything()),
		&corev1.Node{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				subnetMgr.handleAddLeaseEvent(EventAdded, obj)
			},
			UpdateFunc: subnetMgr.handleUpdateLeaseEvent,
			DeleteFunc: func(obj interface{}) {
				_, isNode := obj.(*corev1.Node)
				// We can get DeletedFinalStateUnknown instead of *api.Node here and we need to handle that correctly.
				if !isNode {
					deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
					if !ok {
						klog.Infof(fmt.Sprintf("Error received unexpected object: %v", obj))
						return
					}
					node, ok := deletedState.Obj.(*corev1.Node)
					if !ok {
						klog.Infof(fmt.Sprintf("Error deletedFinalStateUnknown contained non-Node object: %v", deletedState.Obj))
						return
					}
					obj = node
				}
				subnetMgr.handleAddLeaseEvent(EventRemoved, obj)
			},
		}, nil)

	subnetMgr.nodeStore = listersv1.NewNodeLister(store.(cache.Indexer))
	subnetMgr.controller = controller

	return subnetMgr, nil
}

func (subnetMgr *kubeSubnetManager) Run(ctx context.Context) {
	go subnetMgr.controller.Run(ctx.Done())
}

func (subnetMgr *kubeSubnetManager) handleAddLeaseEvent(et EventType, obj interface{}) {
	// 如果 k8s node 没有 "flannel.alpha.coreos.com/kube-subnet-manager" annotation 则 skip
	node := obj.(*corev1.Node)
	if node.Annotations == nil {
		return
	}
	if s, ok := node.Annotations[subnetMgr.annotations.SubnetKubeManaged]; !ok || s != "true" {
		return
	}

	l, err := subnetMgr.nodeToLease(*node)
	if err != nil {
		klog.Infof(fmt.Sprintf("Error turning node %s to lease: %v", node.Name, err))
		return
	}

	subnetMgr.events <- Event{Type: et, Lease: l}
}

func (subnetMgr *kubeSubnetManager) handleUpdateLeaseEvent(oldObj, newObj interface{}) {
	o := oldObj.(*corev1.Node)
	n := newObj.(*corev1.Node)
	if n.Annotations == nil {
		return
	}
	if s, ok := n.Annotations[subnetMgr.annotations.SubnetKubeManaged]; !ok || s != "true" {
		return
	}
	var changed = true
	if subnetMgr.enableIPv4 &&
		o.Annotations[subnetMgr.annotations.BackendData] == n.Annotations[subnetMgr.annotations.BackendData] &&
		o.Annotations[subnetMgr.annotations.BackendType] == n.Annotations[subnetMgr.annotations.BackendType] &&
		o.Annotations[subnetMgr.annotations.BackendPublicIP] == n.Annotations[subnetMgr.annotations.BackendPublicIP] {
		changed = false
	}

	if !changed {
		return // No change to lease
	}

	l, err := subnetMgr.nodeToLease(*n)
	if err != nil {
		klog.Infof("Error turning node %q to lease: %v", n.ObjectMeta.Name, err)
		return
	}
	subnetMgr.events <- Event{Type: EventAdded, Lease: l}
}

func (subnetMgr *kubeSubnetManager) nodeToLease(node corev1.Node) (l Lease, err error) {
	if subnetMgr.enableIPv4 {
		l.Attrs.PublicIP, err = ip.ParseIP4(node.Annotations[subnetMgr.annotations.BackendPublicIP])
		if err != nil {
			return l, err
		}
		l.Attrs.BackendData = json.RawMessage(node.Annotations[subnetMgr.annotations.BackendData])

		// INFO: 注意这里读取的是 node.Spec.PodCIDR
		_, cidr, err := net.ParseCIDR(node.Spec.PodCIDR)
		if err != nil {
			return l, err
		}
		l.Subnet = ip.FromIPNet(cidr)
		l.EnableIPv4 = subnetMgr.enableIPv4
	}

	l.Attrs.BackendType = node.Annotations[subnetMgr.annotations.BackendType]
	return l, nil
}

func (subnetMgr *kubeSubnetManager) GetNetworkConfig(ctx context.Context) (*Config, error) {
	return subnetMgr.subnetConf, nil
}

func (subnetMgr *kubeSubnetManager) Name() string {
	return fmt.Sprintf("Kubernetes Subnet Manager - %s", subnetMgr.nodeName)
}

func (subnetMgr *kubeSubnetManager) patchNodeNetworkUnavailable(ctx context.Context) error {
	conditions := []corev1.NodeCondition{
		{
			Type:               corev1.NodeNetworkUnavailable,
			Status:             corev1.ConditionFalse,
			Reason:             "FlannelIsUp",
			Message:            "Flannel is running on this node",
			LastTransitionTime: metav1.Now(),
			LastHeartbeatTime:  metav1.Now(),
		},
	}
	raw, err := json.Marshal(&conditions)
	if err != nil {
		return err
	}
	patch := []byte(fmt.Sprintf(`{"status":{"conditions":%s}}`, raw))
	_, err = subnetMgr.kubeClient.CoreV1().Nodes().PatchStatus(ctx, subnetMgr.nodeName, patch)
	return err
}

func getNodeName(kubeClient *kubernetes.Clientset) (string, error) {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		podName := os.Getenv("POD_NAME")
		podNamespace := os.Getenv("POD_NAMESPACE")
		if podName == "" || podNamespace == "" {
			return "", fmt.Errorf("env variables POD_NAME and POD_NAMESPACE must be set")
		}

		pod, err := kubeClient.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("error retrieving pod spec for '%s/%s': %v", podNamespace, podName, err)
		}
		nodeName = pod.Spec.NodeName
		if nodeName == "" {
			return "", fmt.Errorf("node name not present in pod spec '%s/%s'", podNamespace, podName)
		}
	}

	return nodeName, nil
}

// ipnet1 网段包含 ipnet2
func containsCIDR(ipnet1, ipnet2 *net.IPNet) bool {
	ones1, _ := ipnet1.Mask.Size()
	ones2, _ := ipnet2.Mask.Size()
	return ones1 <= ones2 && ipnet1.Contains(ipnet2.IP)
}
