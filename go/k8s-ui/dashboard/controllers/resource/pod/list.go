package pod

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/dataselect"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/metric"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/event"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/namespace"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Pod struct {
	Metrics *PodMetrics `json:"metrics"`
	Warnings []event.Event `json:"warnings"`
}


type PodList struct {
	Pods []Pod `json:"pods"`
	Errors []error `json:"errors"`
	Status common.ResourceStatus `json:"status"`
	
	Metrics []common.Metric `json:"metrics"`
	
}


func ListPod(k8sClient kubernetes.Interface,
	metricClient metric.MetricClient,
	namespaceQuery *namespace.NamespaceQuery,
	dataSelectQuery *dataselect.DataSelectQuery)  {
	channels := common.ResourceChannels{
		PodListChannel: GetPodListChannelWithOptions(k8sClient, namespaceQuery, metav1.ListOptions{}, 1),
		EventListChannel: event.GetEventListChannelWithOptions(k8sClient, namespaceQuery, 1),
	}

	// get pods
	pods := <-channels.PodListChannel.List
	err1 := <-channels.PodListChannel.Error
	// get events
	events := <-channels.EventListChannel.List
	err2 := <-channels.EventListChannel.Error

	
	// merge pod/events
	podList := ToPodList(pods.Items, events.Items, dataSelectQuery, metricClient)
	podList.Status = getPodStatus(pods, events.Items)

	return podList
}

func GetPodList(client kubernetes.Interface) {

	
	
}

func ToPodList(pods []corev1.Pod, events []corev1.Event) PodList {
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




