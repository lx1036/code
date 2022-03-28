package network

import (
	"context"
	"flag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"testing"
)

// INFO: 这里有个坑，kubeconfig flag 不能置于函数内，否则报错 no flags --kubeconfig。应该是提前 flag 初始化。
var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
)

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
	svc, _ := kubeClient.CoreV1().Services("default").Get(context.TODO(), "nginx-demo", metav1.GetOptions{})
	svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{IP: "100.20.30.43"}}
	kubeClient.CoreV1().Services("default").UpdateStatus(context.Background(), svc, metav1.UpdateOptions{})
}
