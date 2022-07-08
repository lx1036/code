package network

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// INFO: 这里有个坑，kubeconfig flag 不能置于函数内，否则报错 no flags --kubeconfig。应该是提前 flag 初始化。
var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
)

// INFO: 这里已经是 ListAndWatch 机制了，服务端增量数据会被 push 到本地 cache，但是这里没有设置 event handler 去 **立即** 处理，
//  而是周期处理，比如 BGP 宣告 pod cidr 或者 service clusterIP/loadbalancerIP ！！！
func TestK8sListAndWatch(test *testing.T) {
	flag.Parse()

	if len(*kubeconfig) == 0 {
		klog.Fatal("kubeconfig is empty")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 0,
		informers.WithTweakListOptions(func(listOption *metav1.ListOptions) {
			listOption.AllowWatchBookmarks = true
		}))
	podInformer := informerFactory.Core().V1().Pods()
	podsSynced := podInformer.Informer().HasSynced
	podsLister := podInformer.Lister()
	ctx := genericapiserver.SetupSignalContext()
	informerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), podsSynced) {
		klog.Fatalf("failed to wait for caches to sync")
		return
	}

	klog.Info("cache is synced")

	listPods := func() {
		pods, err := podsLister.Pods(corev1.NamespaceDefault).List(labels.Everything())
		if err != nil {
			klog.Fatal(err)
		}
		for _, pod := range pods {
			klog.Infof(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
		}
	}
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		listPods()
	}, time.Second*5)

	<-ctx.Done()
}

func TestLoadBalancer(test *testing.T) {
	flag.Parse()

	if len(*kubeconfig) == 0 {
		klog.Fatal("kubeconfig is empty")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	kubeClient := kubernetes.NewForConfigOrDie(restConfig)

	svc, err := kubeClient.CoreV1().Services("default").Get(context.TODO(), "nginx-demo", metav1.GetOptions{})
	if err == nil {
		for _, port := range svc.Spec.Ports {
			klog.Info(port.Name)
		}
	}

	//svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: "100.20.30.43"}}
	//kubeClient.CoreV1().Services("default").UpdateStatus(context.Background(), svc, metav1.UpdateOptions{})

	node, _ := kubeClient.CoreV1().Nodes().Get(context.TODO(), "xxx", metav1.GetOptions{})
	nodeCopy := node.DeepCopy()
	//nodeCopy.Spec.PodCIDRs = []string{}
	//nodeCopy.Spec.PodCIDR = ""
	nodeCopy.Spec = corev1.NodeSpec{}
	_, err = kubeClient.CoreV1().Nodes().Update(context.TODO(), nodeCopy, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err) // Forbidden: node updates may only change labels, taints, or capacity (or configSource, if the DynamicKubeletConfig feature gate is enabled)
	}
}

type nodeForCIDRMergePatch struct {
	Spec nodeSpecForMergePatch `json:"spec"`
}

type nodeSpecForMergePatch struct {
	PodCIDR  string   `json:"podCIDR"`
	PodCIDRs []string `json:"podCIDRs,omitempty"`
}

// staging/src/k8s.io/component-helpers/node/util/cidr.go
func PatchNodeCIDRs(c clientset.Interface, node types.NodeName, cidrs []string) error {
	// set the pod cidrs list and set the old pod cidr field
	patch := nodeForCIDRMergePatch{
		Spec: nodeSpecForMergePatch{
			PodCIDR:  cidrs[0],
			PodCIDRs: cidrs,
		},
	}

	patchBytes, err := json.Marshal(&patch)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal CIDR: %v", err)
	}
	klog.V(4).Infof("cidrs patch bytes are:%s", string(patchBytes))
	if _, err := c.CoreV1().Nodes().Patch(context.TODO(), string(node), types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("failed to patch node CIDR: %v", err)
	}
	return nil
}

func TestMap(test *testing.T) {
	requestedIPs := map[string]string{"a": "a"} //net.IP cannot be a key
	remainingIPs := map[string]string{}
	remainingIPs = requestedIPs
	delete(requestedIPs, "a")
	klog.Info(remainingIPs) // map[]
}

func TestRouteTable(test *testing.T) {
	//klog.Infof(fmt.Sprintf("%d", unix.RT_TABLE_MAIN))
	klog.Infof(fmt.Sprintf("%d", 0xfe)) // 254
}
