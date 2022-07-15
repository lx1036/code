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

const (
	nodeControllerSyncTimeout = 10 * time.Minute
)

type Manager interface {
	GetNetworkConfig(ctx context.Context) (*Config, error)
	AcquireLease(ctx context.Context, attrs *LeaseAttrs) (*Lease, error)
	RenewLease(ctx context.Context, lease *Lease) error
	WatchLease(ctx context.Context, sn ip.IP4Net, sn6 ip.IP6Net, cursor interface{}) (LeaseWatchResult, error)
	WatchLeases(ctx context.Context, cursor interface{}) (LeaseWatchResult, error)

	Name() string
}

type EventType int

const (
	EventAdded EventType = iota
	EventRemoved
)

type LeaseAttrs struct {
	PublicIP      ip.IP4
	BackendType   string          `json:",omitempty"`
	BackendData   json.RawMessage `json:",omitempty"`
	BackendV6Data json.RawMessage `json:",omitempty"`
}

type Lease struct {
	EnableIPv4 bool
	Subnet     ip.IP4Net
	Attrs      LeaseAttrs
	Expiration time.Time

	Asof int64
}

func (l *Lease) Key() string {
	return MakeSubnetKey(l.Subnet)
}

func MakeSubnetKey(sn ip.IP4Net) string {
	return sn.StringSep(".", "-")
}

type Event struct {
	Type  EventType `json:"type"`
	Lease Lease     `json:"lease,omitempty"`
}

type kubeSubnetManager struct {
	enableIPv4 bool
	subnetConf *Config
	nodeName   string

	kubeClient kubernetes.Interface
	store      cache.Store
	controller cache.Controller

	annotations               annotations
	nodeStore                 listersv1.NodeLister
	nodeController            cache.Controller
	setNodeNetworkUnavailable bool

	events chan Event
}

// NewSubnetManager netConfPath=/etc/kube-flannel/net-conf.json
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
	subnetManager, err := newKubeSubnetManager(ctx, kubeClient, config, nodeName, prefix)
	if err != nil {
		return nil, fmt.Errorf("error creating network manager: %s", err)
	}
	subnetManager.setNodeNetworkUnavailable = setNodeNetworkUnavailable
	go subnetManager.Run(context.Background())
	klog.Infof("Waiting %s for node controller to sync", nodeControllerSyncTimeout)
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
	config *Config, nodeName, prefix string) (*kubeSubnetManager, error) {
	annotation, err := newAnnotations(prefix)
	if err != nil {
		klog.Fatal(err)
	}
	subnetMgr := &kubeSubnetManager{
		enableIPv4:  config.EnableIPv4,
		kubeClient:  kubeClient,
		subnetConf:  config,
		nodeName:    nodeName,
		annotations: annotation,
		events:      make(chan Event, 5000),
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

	subnetMgr.store = store
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
		klog.Infof(fmt.Sprintf("Error turning node %q to lease: %v", n.ObjectMeta.Name, err))
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

func (subnetMgr *kubeSubnetManager) AcquireLease(ctx context.Context, attrs *interface{}) (*interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (subnetMgr *kubeSubnetManager) RenewLease(ctx context.Context, lease *interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (subnetMgr *kubeSubnetManager) WatchLease(ctx context.Context, sn interface{}, sn6 interface{}, cursor interface{}) (LeaseWatchResult, error) {
	//TODO implement me
	panic("implement me")
}

func (subnetMgr *kubeSubnetManager) WatchLeases(ctx context.Context, cursor interface{}) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (subnetMgr *kubeSubnetManager) Name() string {
	//TODO implement me
	panic("implement me")
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
