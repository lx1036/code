package event

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type EventListChannel struct {
	List chan *corev1.EventList
	Error chan error
}


func GetEventListChannelWithOptions(
	client kubernetes.Interface,
	namespace string,
	options metav1.ListOptions,
	numReads int) EventListChannel {
	channel := EventListChannel{
		List: make(chan *corev1.EventList, numReads),
		Error: make(chan error, numReads),
	}
	
	
	go func() {
		list, err := client.CoreV1().Events(namespace).List(context.TODO(), options)
		
		for i:=0; i< numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()
	
	return channel
}



