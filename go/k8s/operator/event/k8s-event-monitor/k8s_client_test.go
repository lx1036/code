package main

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubewatch "k8s.io/apimachinery/pkg/watch"
	kubeclient "k8s.io/client-go/kubernetes"
	kuberest "k8s.io/client-go/rest"
	"strconv"
	"testing"
)

// kubectl proxy --port=8001
func TestCmdClient(test *testing.T) {
	kubeConfig := &kuberest.Config{
		Host: "http://127.0.0.1:8001",
		TLSClientConfig: kuberest.TLSClientConfig{
			Insecure: true,
		},
	}

	clientSet, err := kubeclient.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	pods, err := clientSet.CoreV1().Pods(apiv1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("numbers of pods: " + strconv.Itoa(len(pods.Items)))

	go watch(clientSet)

	go log()

	select {}
}

func watch(clientSet *kubeclient.Clientset) {
	for {
		events, err := clientSet.CoreV1().Events(apiv1.NamespaceDefault).List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		/*for _, event := range events.Items {
			fmt.Println(fmt.Sprintf("message: %s, reason: %s, type: %s", event.Message, event.Reason, event.Type))
		}*/

		watcher, err := clientSet.CoreV1().Events(apiv1.NamespaceDefault).Watch(metav1.ListOptions{
			ResourceVersion: events.ResourceVersion,
		})
		if err != nil {
			panic(err.Error())
		}
		watchChannel := watcher.ResultChan()
	nextTick:
		for {
			select {
			case eventObj, ok := <-watchChannel:
				if !ok {
					break nextTick
				}
				if event, ok := eventObj.Object.(*apiv1.Event); ok {
					switch eventObj.Type {
					case kubewatch.Deleted:
						fmt.Println(fmt.Sprintf("[deleted]: %s", event.Message))
					case kubewatch.Added:
						eventsBuffer <- event
					case kubewatch.Modified:
						fmt.Println(fmt.Sprintf("[modified]: %s", event.Message))
					default:
						fmt.Println(fmt.Sprintf("[%s]: %s", event.Type, event.Message))
					}
				} else {
					break nextTick
				}
			}
		}
	}
}

var (
	eventsBuffer = make(chan *apiv1.Event, 5)
)

func log() {
	for {
		var events []*apiv1.Event
	nextTick:
		for {
			select {
			case event := <-eventsBuffer:
				events = append(events, event)
			default:
				break nextTick
			}

			for _, event := range events {
				fmt.Println(fmt.Sprintf("[%s]: %s", event.Type, event.Message))
			}
		}
	}

}
