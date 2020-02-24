package main

import (
	"flag"
	"fmt"
	informers "k8s-lx1036/k8s-ui/backend/kubernetes/crd/sample-controller/pkg/generated/informers/externalversions"
	"k8s-lx1036/k8s-ui/backend/kubernetes/crd/sample-controller/pkg/generated/informers/externalversions/samplecontroller"
	"k8s-lx1036/k8s-ui/backend/kubernetes/crd/sample-controller/pkg/generated/informers/externalversions/samplecontroller/v1alpha1"
	"k8s-lx1036/k8s-ui/backend/kubernetes/crd/sample-controller/pkg/signals"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"time"
)

var (
	masterURL  string
	kubeconfig *string
)

func init() {
	flag.StringVar(kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func main() {
	if home, _ := os.UserHomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}

	fmt.Println("kube config path: " + *kubeconfig)

	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, *kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// k8s.io/apiextensions-apiserver focus api-extension
	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)
	var sampleController samplecontroller.Interface
	sampleController = exampleInformerFactory.Samplecontroller()
	var sampleControllerV1alpha1 v1alpha1.Interface
	sampleControllerV1alpha1 =  sampleController.V1alpha1()

	controller := NewController(kubeClient, exampleClient,
		kubeInformerFactory.Apps().V1().Deployments(),
		sampleControllerV1alpha1.Foos())

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
