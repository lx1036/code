package clusterstate

import (
	"fmt"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/controller/clusterstate/prometheus"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/controller/clusterstate/types"
	"time"

	apisv1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/apis/autoscaling.k9s.io/v1"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client/clientset/versioned"
	vpaInformers "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client/informers/externalversions"
	listersv1 "k8s-lx1036/k8s/monitor/vpa/recommender/pkg/client/listers/autoscaling.k9s.io/v1"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/controller/clusterstate/metrics"
	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/input/target"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	typedCorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

const (
	defaultResyncPeriod = 10 * time.Minute
)

type condition struct {
	conditionType apisv1.VerticalPodAutoscalerConditionType
	delete        bool
	message       string
}

type ClusterStateFeeder struct {
	coreClient    typedCorev1.CoreV1Interface
	specClient    SpecClient
	metricsClient *metrics.MetricsClient

	//oomChan             <-chan oom.OomInfo
	//vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	vpaLister   listersv1.VerticalPodAutoscalerLister
	vpaInformer cache.SharedIndexInformer

	clusterState      *types.ClusterState
	selectorFetcher   target.VpaTargetSelectorFetcher
	memorySaveMode    bool
	controllerFetcher controllerfetcher.ControllerFetcher
}

func NewClusterStateFeeder(config *rest.Config, clusterState *types.ClusterState, memorySave bool, namespace string) *ClusterStateFeeder {
	kubeClient := kubernetes.NewForConfigOrDie(config)

	factory := informers.NewSharedInformerFactoryWithOptions(kubeClient, defaultResyncPeriod, informers.WithNamespace(namespace))

	vpaClient := versioned.NewForConfigOrDie(config)
	vpaFactory := vpaInformers.NewSharedInformerFactoryWithOptions(vpaClient, defaultResyncPeriod)

	c := &ClusterStateFeeder{
		coreClient:    kubeClient.CoreV1(),
		specClient:    nil,
		metricsClient: metrics.NewMetricsClient(config, namespace),
		//oomChan:             nil,
		//vpaCheckpointClient: nil,
		vpaLister:         vpaFactory.Autoscaling().V1().VerticalPodAutoscalers().Lister(),
		vpaInformer:       vpaFactory.Autoscaling().V1().VerticalPodAutoscalers().Informer(),
		clusterState:      nil,
		selectorFetcher:   nil,
		memorySaveMode:    false,
		controllerFetcher: nil,
	}

}

func (clusterStateFeeder *ClusterStateFeeder) Start(stopCh <-chan struct{}) error {
	go clusterStateFeeder.vpaInformer.Run(stopCh)

	shutdown := cache.WaitForCacheSync(stopCh, clusterStateFeeder.vpaInformer.HasSynced)
	if !shutdown {
		return fmt.Errorf("can not sync sparkApplication and pods in clusterStateFeeder controller")
	}

	return nil
}

// INFO: load vpa 对象到 clusterstate 缓存
func (clusterStateFeeder *ClusterStateFeeder) LoadVPAs() {
	vpas, err := clusterStateFeeder.vpaLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("Cannot list VPAs. Reason: %+v", err)
		return
	}

	klog.V(2).Infof("Fetched %d VPAs.", len(vpas))

	// Add or update existing VPAs in the model.
	vpaKeys := make(map[types.VpaID]bool)
	for _, vpa := range vpas {
		vpaID := types.VpaID{
			Namespace: vpa.Namespace,
			VpaName:   vpa.Name,
		}

		selector, conditions := clusterStateFeeder.getSelector(vpa)
		klog.Infof("Using selector %s for VPA %s/%s", selector.String(), vpa.Namespace, vpa.Name)

		if clusterStateFeeder.clusterState.AddOrUpdateVpa(vpa, selector) == nil {
			// Successfully added VPA to the model.
			vpaKeys[vpaID] = true

			for _, condition := range conditions {
				if condition.delete {
					// INFO: 每一个属性值都是指针或者map类型
					delete(clusterStateFeeder.clusterState.Vpas[vpaID].Conditions, condition.conditionType)
				} else {
					clusterStateFeeder.clusterState.Vpas[vpaID].Conditions.Set(condition.conditionType, true, "", condition.message)
				}
			}
		}
	}

	// Delete non-existent VPAs from the model.
	for vpaID := range clusterStateFeeder.clusterState.Vpas {
		if _, exists := vpaKeys[vpaID]; !exists {
			klog.V(3).Infof("Deleting VPA %v", vpaID)
			clusterStateFeeder.clusterState.DeleteVpa(vpaID)
		}
	}

	clusterStateFeeder.clusterState.ObservedVpas = vpas
}

// INFO:
func (clusterStateFeeder *ClusterStateFeeder) getSelector(vpa *apisv1.VerticalPodAutoscaler) (labels.Selector, []condition) {
	selector, fetchErr := clusterStateFeeder.selectorFetcher.Fetch(vpa)
	if selector != nil {
		validTargetRef, unsupportedCondition := clusterStateFeeder.validateTargetRef(vpa)
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
// INFO:
func (clusterStateFeeder *ClusterStateFeeder) LoadPods() {
	podSpecs, err := clusterStateFeeder.specClient.GetPodSpecs()
	if err != nil {
		klog.Errorf("Cannot get SimplePodSpecs. Reason: %+v", err)
	}
	pods := make(map[types.PodID]*BasicPodSpec)
	for _, spec := range podSpecs {
		pods[spec.ID] = spec
	}

	for _, pod := range pods {
		if clusterStateFeeder.memorySaveMode && !clusterStateFeeder.matchesVPA(pod) {
			continue
		}
		clusterStateFeeder.clusterState.AddOrUpdatePod(pod.ID, pod.PodLabels, pod.Phase)
		for _, container := range pod.Containers {
			if err = clusterStateFeeder.clusterState.AddOrUpdateContainer(container.ID, container.Request); err != nil {
				klog.Warningf("Failed to add container %+v. Reason: %+v", container.ID, err)
			}
		}
	}
}

func newContainerUsageSamplesWithKey(metrics *metrics.ContainerMetricsSnapshot) []*types.ContainerUsageSampleWithKey {
	var samples []*ContainerUsageSampleWithKey

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

func (clusterStateFeeder *ClusterStateFeeder) LoadRealTimeMetrics() {
	containerMetricsSnapshots, err := clusterStateFeeder.metricsClient.GetContainersMetrics()
	if err != nil {
		klog.Errorf("Cannot get ContainerMetricsSnapshot from MetricsClient. Reason: %+v", err)
	}

	sampleCount := 0
	droppedSampleCount := 0
	for _, containerMetricsSnapshot := range containerMetricsSnapshots {
		for _, sample := range newContainerUsageSamplesWithKey(containerMetricsSnapshot) {
			if err := clusterStateFeeder.clusterState.AddSample(sample); err != nil { // INFO: 缓存 sample
				droppedSampleCount++
			} else {
				sampleCount++
			}
		}
	}

	klog.V(3).Infof("ClusterSpec fed with #%v ContainerUsageSamples for #%v containers. Dropped #%v samples.",
		sampleCount, len(containerMetricsSnapshots), droppedSampleCount)
}

// INFO: 从 prometheus 中读取出cluster history数据，然后缓存到clusterstate中
func (clusterStateFeeder *ClusterStateFeeder) LoadFromPrometheusProvider(config prometheus.PrometheusHistoryProviderConfig) error {
	provider, err := prometheus.NewPrometheusHistoryProvider(config)
	if err != nil {
		return err
	}

	clusterHistory, err := provider.GetClusterHistory()
	if err != nil {
		return err
	}

	for podID, podHistory := range clusterHistory {
		clusterStateFeeder.clusterState.AddOrUpdatePod(podID, podHistory.LastLabels, corev1.PodUnknown)
		for containerName, containerUsageSamples := range podHistory.Samples {
			containerID := types.ContainerID{
				PodID:         podID,
				ContainerName: containerName,
			}

			for _, containerUsageSample := range containerUsageSamples {
				err = clusterStateFeeder.clusterState.AddSample(&types.ContainerUsageSampleWithKey{
					ContainerUsageSample: containerUsageSample,
					Container:            containerID,
				})
				if err != nil {
					klog.Warning(err)
				}
			}
		}
	}
}
