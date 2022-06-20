package v1

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// INFO: 使用 controller-runtime client 可以不用生成 crd 的 client/lister/informer，省事多了

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
)

func TestControllerRuntimeClient(test *testing.T) {
	flag.Parse()

	if len(*kubeconfig) == 0 {
		klog.Errorf("--kubeconfig should be required")
		return
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Fatal(err)
	}

	pods, err := kubeClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}
	for _, pod := range pods.Items {
		klog.Info(fmt.Sprintf("%s/%s", pod.Namespace, pod.Name))
	}

	// INFO: crdClient/k8sClient, 不过还是使用 kubeClient 熟悉些。
	mapper, err := apiutil.NewDiscoveryRESTMapper(restConfig)
	if err != nil {
		klog.Fatal(err)
	}
	//scheme := runtime.NewScheme()
	_ = AddToScheme(scheme.Scheme) // INFO: 使用 scheme.Scheme，这样 crdClient 就可以 crd resource, 也可以内置的 k8s resource
	crdClient, err := client.New(restConfig, client.Options{Scheme: scheme.Scheme, Mapper: mapper})
	if err != nil {
		klog.Fatal(err)
	}

	namespace := "default"
	name := "ippool1"
	pool := &IPPool{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
	}
	if err = crdClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, pool); errors.IsNotFound(err) {
		klog.Error(err)
		return
	}
	klog.Info(fmt.Sprintf("cidr is %s", pool.Spec.Range)) // cidr is 100.30.30.0/24

	// has namespace
	podName := "nginx-demo-7d99f85fc5-cfmvm"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: podName},
	}
	if err = crdClient.Get(context.TODO(), types.NamespacedName{Name: podName, Namespace: namespace}, pod); err != nil {
		klog.Error(err)
	}
	klog.Info(fmt.Sprintf("pod image is %s", pod.Spec.Containers[0].Image))

	// has no namespace
	nodeName := ""
	if len(nodeName) != 0 {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: nodeName},
		}
		if err = crdClient.Get(context.TODO(), types.NamespacedName{Name: nodeName}, node); err != nil {
			klog.Error(err)
		}
		klog.Info(fmt.Sprintf("pod cidr %s in node %s", node.Spec.PodCIDR, node.Name))
	}
}
