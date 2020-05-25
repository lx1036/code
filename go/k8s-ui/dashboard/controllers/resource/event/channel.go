package event

import (
	"context"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/namespace"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type EventListChannel struct {
	List chan *corev1.EventList
	Error chan error
}


func GetEventListChannelWithOptions(
	client kubernetes.Interface,
	namespaceQuery *namespace.NamespaceQuery,
	numReads int) EventListChannel {
	channel := EventListChannel{
		List: make(chan *corev1.EventList, numReads),
		Error: make(chan error, numReads),
	}
	
	
	go func() {
		list, err := client.CoreV1().Events(namespaceQuery.GetNamespace()).List(context.TODO(), dataselect.ListEverything)
		
		for i:=0; i< numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()
	
	return channel
}



