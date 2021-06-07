package input

import (
	"fmt"

	apisv1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis/autoscaling.k9s.io/v1"
	listersv1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client/listers/autoscaling.k9s.io/v1"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/target"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/types"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type condition struct {
	conditionType apisv1.VerticalPodAutoscalerConditionType
	delete        bool
	message       string
}

type ClusterStateFeeder struct {
	coreClient    corev1.CoreV1Interface
	specClient    types.SpecClient
	metricsClient *MetricsClient
	//oomChan             <-chan oom.OomInfo
	//vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	vpaLister         listersv1.VerticalPodAutoscalerLister
	clusterState      *types.ClusterState
	selectorFetcher   target.VpaTargetSelectorFetcher
	memorySaveMode    bool
	controllerFetcher controllerfetcher.ControllerFetcher
}

func NewClusterStateFeeder(config *rest.Config, clusterState *types.ClusterState, memorySave bool, namespace string) *ClusterStateFeeder {
	kubeClient := kubernetes.NewForConfigOrDie(config)

	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(namespace))

	c := &ClusterStateFeeder{
		coreClient:    kubeClient.CoreV1(),
		specClient:    nil,
		metricsClient: NewMetricsClient(resourceclient.NewForConfigOrDie(config), namespace),
		//oomChan:             nil,
		//vpaCheckpointClient: nil,
		vpaLister:         nil,
		clusterState:      nil,
		selectorFetcher:   nil,
		memorySaveMode:    false,
		controllerFetcher: nil,
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

	klog.V(3).Infof("Fetched %d VPAs.", len(vpaCRDs))

	// Add or update existing VPAs in the model.
	vpaKeys := make(map[types.VpaID]bool)
	for _, vpaCRD := range vpaCRDs {
		vpaID := types.VpaID{
			Namespace: vpaCRD.Namespace,
			VpaName:   vpaCRD.Name,
		}

		selector, conditions := feeder.getSelector(vpaCRD)
		klog.Infof("Using selector %s for VPA %s/%s", selector.String(), vpaCRD.Namespace, vpaCRD.Name)

		if feeder.clusterState.AddOrUpdateVpa(vpaCRD, selector) == nil {
			// Successfully added VPA to the model.
			vpaKeys[vpaID] = true

			for _, condition := range conditions {
				if condition.delete {
					delete(feeder.clusterState.Vpas[vpaID].Conditions, condition.conditionType)
				} else {
					feeder.clusterState.Vpas[vpaID].Conditions.Set(condition.conditionType, true, "", condition.message)
				}
			}
		}
	}

	// Delete non-existent VPAs from the model.
	for vpaID := range feeder.clusterState.Vpas {
		if _, exists := vpaKeys[vpaID]; !exists {
			klog.V(3).Infof("Deleting VPA %v", vpaID)
			feeder.clusterState.DeleteVpa(vpaID)
		}
	}

	feeder.clusterState.ObservedVpas = vpaCRDs
}

func (feeder *ClusterStateFeeder) getSelector(vpa *apisv1.VerticalPodAutoscaler) (labels.Selector, []condition) {
	selector, fetchErr := feeder.selectorFetcher.Fetch(vpa)
	if selector != nil {
		validTargetRef, unsupportedCondition := feeder.validateTargetRef(vpa)
		if !validTargetRef {
			return labels.Nothing(), []condition{
				unsupportedCondition,
				{conditionType: apisv1.ConfigDeprecated, delete: true},
			}
		}
		return selector, []condition{
			{conditionType: apisv1.ConfigUnsupported, delete: true},
			{conditionType: apisv1.ConfigDeprecated, delete: true},
		}
	}

	msg := "Cannot read targetRef"
	if fetchErr != nil {
		klog.Errorf("Cannot get target selector from VPA's targetRef. Reason: %+v", fetchErr)
		msg = fmt.Sprintf("Cannot read targetRef. Reason: %s", fetchErr.Error())
	}
	return labels.Nothing(), []condition{
		{conditionType: apisv1.ConfigUnsupported, delete: false, message: msg},
		{conditionType: apisv1.ConfigDeprecated, delete: true},
	}
}

// Load pod into the cluster state.
func (feeder *ClusterStateFeeder) LoadPods() {
	podSpecs, err := feeder.specClient.GetPodSpecs()
	if err != nil {
		klog.Errorf("Cannot get SimplePodSpecs. Reason: %+v", err)
	}
	pods := make(map[types.PodID]*types.BasicPodSpec)
	for _, spec := range podSpecs {
		pods[spec.ID] = spec
	}

	for _, pod := range pods {
		if feeder.memorySaveMode && !feeder.matchesVPA(pod) {
			continue
		}
		feeder.clusterState.AddOrUpdatePod(pod.ID, pod.PodLabels, pod.Phase)
		for _, container := range pod.Containers {
			if err = feeder.clusterState.AddOrUpdateContainer(container.ID, container.Request); err != nil {
				klog.Warningf("Failed to add container %+v. Reason: %+v", container.ID, err)
			}
		}
	}
}

func newContainerUsageSamplesWithKey(metrics *ContainerMetricsSnapshot) []*types.ContainerUsageSampleWithKey {
	var samples []*types.ContainerUsageSampleWithKey

	for metricName, resourceAmount := range metrics.Usage {
		sample := &types.ContainerUsageSampleWithKey{
			Container: metrics.ID,
			ContainerUsageSample: types.ContainerUsageSample{
				MeasureStart: metrics.SnapshotTime,
				Resource:     metricName,
				Usage:        resourceAmount,
			},
		}
		samples = append(samples, sample)
	}
	return samples
}

func (feeder *ClusterStateFeeder) LoadRealTimeMetrics() {
	containerMetricsSnapshots, err := feeder.metricsClient.GetContainersMetrics()
	if err != nil {
		klog.Errorf("Cannot get ContainerMetricsSnapshot from MetricsClient. Reason: %+v", err)
	}

	sampleCount := 0
	droppedSampleCount := 0
	for _, containerMetricsSnapshot := range containerMetricsSnapshots {
		for _, sample := range newContainerUsageSamplesWithKey(containerMetricsSnapshot) {
			if err := feeder.clusterState.AddSample(sample); err != nil { // INFO: 缓存 sample
				droppedSampleCount++
			} else {
				sampleCount++
			}
		}
	}

	klog.V(3).Infof("ClusterSpec fed with #%v ContainerUsageSamples for #%v containers. Dropped #%v samples.",
		sampleCount, len(containerMetricsSnapshots), droppedSampleCount)
}
