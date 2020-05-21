package pod

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/event"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Pod struct {
	Metrics *PodMetrics `json:"metrics"`
	Warnings []common.Event `json:"warnings"`
}


type PodList struct {
	Pods []Pod `json:"pods"`
	Errors []error `json:"errors"`
	Status common.ResourceStatus `json:"status"`
	
	Metrics []common.Metric `json:"metrics"`
	
}


func ListPod(client kubernetes.Interface)  {
	
	
	
	pods := <-channels.PodList.List
	err := <-channels.PodList.Error
	
	
	
}

func GetPodList(client kubernetes.Interface) {
	channels := common.ResourceChannels{
		PodListChannel: GetPodListChannelWithOptions(client, 1),
		EventListChannel: event.GetEventListChannelWithOptions(client, 1),
	}
	
	pods := <-channels.PodListChannel.List
	err1 := <-channels.PodListChannel.Error
	
	events := <-channels.EventListChannel.List
	err2 := <-channels.EventListChannel.Error
	
	
}

func ToPodList(pods []corev1.Pod, events []corev1.Event) {
	podList := PodList{
		Pods: make([]Pod, 0),
	}
	
	for _, pod := range pods {
		warnings := event.GetPodWarningEvents(events, []corev1.Pod{pod})
		podDetail := podWithMetricsEvents(&pod, metrics, warnings)
		podList.Pods = append(podList.Pods, podDetail)
	}
	
}

func podWithMetricsEvents(pod *Pod, metrics *PodMetrics, events []corev1.Event) Pod {
	podDetail := Pod{
		Metrics: metrics,
	}
}




