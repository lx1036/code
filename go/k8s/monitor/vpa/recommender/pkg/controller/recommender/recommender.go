package recommender

import (
	"fmt"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/controller/clusterstate"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/controller/clusterstate/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/logic"
	"time"

	"k8s-lx1036/k8s/monitor/vpa/recommender/cmd/app/options"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client/clientset/versioned"
	v1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client/clientset/versioned/typed/autoscaling.k9s.io/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type Recommender struct {
	clusterState *clusterstate.ClusterState

	clusterStateFeeder clusterstate.ClusterStateFeeder

	vpaClient v1.AutoscalingV1Interface

	podResourceRecommender PodResourceRecommender
}

func (r *Recommender) RunUntil(stopCh <-chan struct{}) error {
	err := r.clusterStateFeeder.Start(stopCh)
	if err != nil {
		return err
	}

	go wait.Until(func() {
		r.clusterStateFeeder.LoadVPAs()

		r.clusterStateFeeder.LoadPods()

		r.clusterStateFeeder.LoadRealTimeMetrics()

		r.UpdateVPAs()
	}, time.Second*60, stopCh)

	return nil
}

// Updates VPA CRD objects' statuses.
func (r *Recommender) UpdateVPAs() {
	for _, observedVpa := range r.clusterState.ObservedVpas {
		key := clusterstate.VpaID{
			Namespace: observedVpa.Namespace,
			VpaName:   observedVpa.Name,
		}
		vpa, found := r.clusterState.Vpas[key]
		if !found {
			continue
		}

		resources := r.podResourceRecommender.GetRecommendedPodResources(clusterstate.GetContainerNameToAggregateStateMap(vpa))
		vpa.UpdateRecommendation(getCappedRecommendation(vpa.ID, resources, observedVpa.Spec.ResourcePolicy))

		hasMatchingPods := vpa.PodCount > 0
		vpa.UpdateConditions(hasMatchingPods)

		_, err := clusterstate.UpdateVpaStatusIfNeeded(
			r.vpaClient.VerticalPodAutoscalers(vpa.ID.Namespace), vpa.ID.VpaName, vpa.AsStatus(), &observedVpa.Status)
		if err != nil {
			klog.Errorf(
				"Cannot update VPA %v object. Reason: %+v", vpa.ID.VpaName, err)
		}
	}
}

func NewRecommender(option *options.Options) (*Recommender, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", option.Kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}

	vpaClient := versioned.NewForConfigOrDie(restConfig)
	clusterStateFeeder := clusterstate.NewClusterStateFeeder(restConfig)

	recommender := &Recommender{
		clusterState:       nil,
		clusterStateFeeder: clusterstate.ClusterStateFeeder{},
		vpaClient:          vpaClient.AutoscalingV1(),
	}

	return recommender, nil
}

// getCappedRecommendation creates a recommendation based on recommended pod
// resources, setting the UncappedTarget to the calculated recommended target
// and if necessary, capping the Target, LowerBound and UpperBound according
// to the ResourcePolicy.
func getCappedRecommendation(vpaID types.VpaID, resources logic.RecommendedPodResources,
	policy *v1.PodResourcePolicy) *v1.RecommendedPodResources {

}
