package main

import (
	"context"
	"flag"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
	"k8s.io/kubernetes/pkg/kubelet/cri/remote"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	nodeName   = flag.String("node", "", "current node")
)

func main() {
	flag.Parse()

	if len(*kubeconfig) == 0 || len(*nodeName) == 0 {
		klog.Errorf("--kubeconfig or --node should be required")
		return
	}

	clientConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	// INFO: list 当前 node 上的 pods
	stopCh := context.TODO().Done()
	factory := informers.NewSharedInformerFactoryWithOptions(clientSet, time.Second*10, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
		options.FieldSelector = fields.Set{api.PodHostField: string(*nodeName)}.String()
	}))
	podLister := factory.Core().V1().Pods().Lister()
	factory.Start(stopCh)
	informersSynced := []cache.InformerSynced{
		factory.Core().V1().Pods().Informer().HasSynced,
	}
	if !cache.WaitForCacheSync(stopCh, informersSynced...) {
		klog.Errorf("can not sync pods in node %s", *nodeName)
		return
	}

	pods, err := podLister.Pods(metav1.NamespaceAll).List(labels.Everything())
	if err != nil {
		panic(err)
	}
	klog.Infof("%d pods in node %s", len(pods), *nodeName)
	for _, pod := range pods {
		klog.Infof("%s/%s", pod.Namespace, pod.Name)
	}

	remoteRuntimeEndpoint := "unix:///var/run/dockershim.sock"
	remoteRuntimeService, err := remote.NewRemoteRuntimeService(remoteRuntimeEndpoint, time.Minute*2)
	if err != nil {
		panic(err)
	}

	cpus := cpuset.NewCPUSet(1, 13)
	containerID := "0e8b25a584ce27c6c88a59d9411cafc6ac82bd90ee67ccaead109ffbccd46cf4"
	err = remoteRuntimeService.UpdateContainerResources(containerID,
		&runtimeapi.LinuxContainerResources{
			CpusetCpus: cpus.String(),
		})
	if err != nil {
		panic(err)
	}
}
