package demo

import (
	"context"
	"flag"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	apiregistrationclient "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/typed/apiregistration/v1"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
)

func TestApiRegistrationClient(test *testing.T) {
	flag.Parse()

	if len(*kubeconfig) == 0 {
		klog.Errorf("--kubeconfig should be required")
		return
	}

	clientConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client := apiregistrationclient.NewForConfigOrDie(clientConfig)
	apiServiceList, err := client.APIServices().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, apiService := range apiServiceList.Items {
		if apiService.Spec.Service != nil {
			klog.Info(fmt.Sprintf("[apiservice]name: %s, service: %s/%s in third party", apiService.Name, apiService.Spec.Service.Namespace, apiService.Spec.Service.Name))
		} else {
			klog.Info(fmt.Sprintf("[apiservice]name: %s in local", apiService.Name))
		}
	}
}
