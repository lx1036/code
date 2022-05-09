package daemon

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"k8s.io/client-go/tools/clientcmd"
)

const (
	podNetworkTypeENIMultiIP = "ENIMultiIP"
)

func podInfoKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

type K8sService struct {
	client kubernetes.Interface
}

func newK8sServiceOrDie(kubeconfig string, daemonMode string) *K8sService {
	k8sRestConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	client, err := kubernetes.NewForConfig(k8sRestConfig)
	if err != nil {
		klog.Fatal(err)
	}

	k8sService := &K8sService{
		client: client,
	}

	return k8sService
}
