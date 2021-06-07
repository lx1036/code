package input

import (
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/types"
	"k8s.io/apimachinery/pkg/labels"
	
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

)


type ClusterStateFeeder struct {
	coreClient          corev1.CoreV1Interface
	specClient          spec.SpecClient
	metricsClient       *MetricsClient
	oomChan             <-chan oom.OomInfo
	vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	vpaLister           vpa_lister.VerticalPodAutoscalerLister
	clusterState        *model.ClusterState
	selectorFetcher     target.VpaTargetSelectorFetcher
	memorySaveMode      bool
	controllerFetcher   controllerfetcher.ControllerFetcher
}

func NewClusterStateFeeder(config *rest.Config, clusterState *types.ClusterState, memorySave bool, namespace string) *ClusterStateFeeder {
	kubeClient := kubernetes.NewForConfigOrDie(config)
	
	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(namespace))
	
	
	c := &ClusterStateFeeder{
		coreClient:          kubeClient.CoreV1(),
		specClient:          nil,
		metricsClient:       NewMetricsClient(resourceclient.NewForConfigOrDie(config), namespace),
		oomChan:             nil,
		vpaCheckpointClient: nil,
		vpaLister:           nil,
		clusterState:        nil,
		selectorFetcher:     nil,
		memorySaveMode:      false,
		controllerFetcher:   nil,
	}
	
}

// Fetch VPA objects and load them into the cluster state.
func (feeder *ClusterStateFeeder) LoadVPAs() {
	// List VPA API objects.
	vpaCRDs, err := feeder.vpaLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Cannot list VPAs. Reason: %+v", err)
		return
	}
	
	
	
}

// Load pod into the cluster state.
func (feeder *ClusterStateFeeder) LoadPods() {


}

func (feeder *ClusterStateFeeder) LoadRealTimeMetrics() {

}
