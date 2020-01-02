package main

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		namespace := apiv1.NamespaceDefault
		pods, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("there are %d pods in the %s namespace of cluster\n", len(pods.Items), namespace)

		podName := "pod-123"
		pod, err := clientSet.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{
			TypeMeta:        metav1.TypeMeta{},
			ResourceVersion: "",
		})

		if errors.IsNotFound(err) {
			fmt.Printf("pod %s not found in %s namespace of cluster\n", podName, namespace)
		} else if err != nil {
			panic(err.Error())
		} else if err == nil {
			fmt.Println("pod name: " + pod.Name)
		}

		time.Sleep(time.Second * 10)
	}
}
