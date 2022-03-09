package main

import (
	"context"
	"flag"
	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/ipam/allocator/clusterpool"

	"k8s-lx1036/k8s/network/cilium/cilium-ipam/pkg/ippool"

	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	ciliumClientSet "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// go run . --kubeconfig=`echo $HOME`/.kube/config --config=`pwd`/config.yaml
func main() {
	var (
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	)
	flag.Parse()

	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Fatal(err)
	}

	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}

	_, calicoClient := ippool.CreateCalicoClient(*kubeconfig)
	ippoolList, err := calicoClient.IPPools().List(context.TODO(), options.ListOptions{})
	if err != nil {
		klog.Fatal(err)
	}

	for _, node := range nodeList.Items {
		cidrs := ippool.DetermineEnabledIPPoolCIDRs(node, *ippoolList)
		cidrAllocators, err := clusterpool.NewCIDRSets(false, cidrs, 27)
		if err != nil {
			klog.Fatal(err)
		}

		_, v4CIDR, err := allocateFirstFreeCIDR(cidrAllocators)

	}

	// create CiliumNode
	cn := &ciliumv2.CiliumNode{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Spec: ciliumv2.NodeSpec{},
	}
	ciliumClient := ciliumClientSet.NewForConfigOrDie(restConfig)
	_, err = ciliumClient.CiliumV2().CiliumNodes().Create(context.TODO(), cn, metav1.CreateOptions{})
	if err != nil {
		klog.Fatal(err)
	}
}
