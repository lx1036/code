package pod

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type PodListChannel struct {
	List chan *corev1.PodList
	Error chan error
}





func GetPodListChannelWithOptions(client kubernetes.Interface, numReads int) PodListChannel {
	channel := PodListChannel{
		List: make(chan *corev1.PodList, numReads),
		Error: make(chan error, numReads),
	}
	
	
	go func() {
		list, err := client.CoreV1().Pods(namespace).List(context.TODO(), options)
		
		
		for i:=0; i< numReads; i++ {
			channel.List <- list
			channel.Error <- err
		}
	}()
	
	return channel
}
