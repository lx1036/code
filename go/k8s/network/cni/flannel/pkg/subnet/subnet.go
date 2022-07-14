package subnet

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/clientcmd"
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

type kubeSubnetManager struct {
	enableIPv4 bool

	kubeClient kubernetes.Interface
	store      cache.Store
	controller cache.Controller

	annotations               annotations
	nodeName                  string
	nodeStore                 listersv1.NodeLister
	nodeController            cache.Controller
	subnetConf                *Config
	events                    chan Event
	setNodeNetworkUnavailable bool
}

func NewSubnetManager(ctx context.Context, kubeconfig, prefix,
	netConfPath string, setNodeNetworkUnavailable bool) (Manager, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
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

	subnetMgr := &kubeSubnetManager{
		enableIPv4: config.EnableIPv4,
		kubeClient: kubeClient,
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

func (subnetMgr *kubeSubnetManager) GetNetworkConfig(ctx context.Context) (*Config, error) {
	return subnetMgr.subnetConf, nil
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
