package podautoscaler

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/monitor/hpa/pkg/podautoscaler/metrics"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	autoscalinginformers "k8s.io/client-go/informers/autoscaling/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	autoscalinglisters "k8s.io/client-go/listers/autoscaling/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	scaleclient "k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type HorizontalController struct {

	// autoscalinglisters 这里使用的是v1版本，代码里会转换v2beta2
	hpaLister       autoscalinglisters.HorizontalPodAutoscalerLister
	hpaListerSynced cache.InformerSynced
	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced
	scaleNamespacer scaleclient.ScalesGetter

	queue workqueue.RateLimitingInterface

	mapper apimeta.RESTMapper

	replicaCalc *ReplicaCalculator

	// Latest autoscaler events
	scaleUpEvents   map[string][]timestampedScaleEvent
	scaleDownEvents map[string][]timestampedScaleEvent
}

// NewHorizontalController creates a new HorizontalController.
func NewHorizontalController(
	hpaInformer autoscalinginformers.HorizontalPodAutoscalerInformer,
	podInformer coreinformers.PodInformer,
	scaleNamespacer scaleclient.ScalesGetter,
	mapper apimeta.RESTMapper,
	metricsClient *metrics.RestMetricsClient,
) *HorizontalController {

	hpaController := &HorizontalController{
		// @see pkg/controller/podautoscaler/config/v1alpha1/defaults.go
		queue:           workqueue.NewNamedRateLimitingQueue(NewDefaultHPARateLimiter(time.Second*15), "horizontalpodautoscaler"),
		scaleNamespacer: scaleNamespacer,
		mapper:          mapper,
	}

	// @see pkg/controller/podautoscaler/config/v1alpha1/defaults.go
	hpaController.replicaCalc = NewReplicaCalculator(
		metricsClient,
		hpaController.podLister,
		0.1,
		time.Minute*5,
		time.Second*30,
	)

	hpaInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc:    nil,
		UpdateFunc: nil,
		DeleteFunc: nil,
	}, time.Second*10)
	hpaController.hpaLister = hpaInformer.Lister()
	hpaController.hpaListerSynced = hpaInformer.Informer().HasSynced
	hpaController.podLister = podInformer.Lister()
	hpaController.podListerSynced = podInformer.Informer().HasSynced

	return hpaController
}

func (hpa *HorizontalController) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer hpa.queue.ShutDown()

	klog.Infof("Starting HPA controller")
	defer klog.Infof("Shutting down HPA controller")

	if !cache.WaitForNamedCacheSync("HPA", stopCh, hpa.hpaListerSynced, hpa.podListerSynced) {
		return
	}

	// start a single worker (we may wish to start more in the future)
	go wait.Until(func() {
		for hpa.processNextWorkItem() {
		}
	}, time.Second, stopCh)

	<-stopCh
}

func (hpa *HorizontalController) processNextWorkItem() bool {
	key, quit := hpa.queue.Get()
	if quit {
		return false
	}
	defer hpa.queue.Done(key)

	deleted, err := hpa.reconcileKey(key.(string))
	if err != nil {
		utilruntime.HandleError(err)
	}
	// ???
	if !deleted {
		hpa.queue.AddRateLimited(key)
	}

	return true
}

func (a *HorizontalController) reconcileKey(key string) (deleted bool, err error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return true, err
	}

	hpa, err := a.hpaLister.HorizontalPodAutoscalers(namespace).Get(name)
	if errors.IsNotFound(err) {
		klog.Infof("Horizontal Pod Autoscaler %s has been deleted in %s", name, namespace)
		//delete(a.recommendations, key)
		//delete(a.scaleUpEvents, key)
		//delete(a.scaleDownEvents, key)
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return false, a.reconcileAutoscaler(hpa, key)
}

func (a *HorizontalController) reconcileAutoscaler(hpav1Shared *autoscalingv1.HorizontalPodAutoscaler, key string) error {
	// make a copy so that we never mutate the shared informer cache (conversion can mutate the object)
	hpav1 := hpav1Shared.DeepCopy()
	// autoscaling/v1 转换成 autoscaling/v2beta2 版本
	hpaRaw, err := UnsafeConvertToVersionVia(hpav1, autoscalingv2.SchemeGroupVersion)
	if err != nil {
		//a.eventRecorder.Event(hpav1, v1.EventTypeWarning, "FailedConvertHPA", err.Error())
		return fmt.Errorf("failed to convert the given HPA to %s: %v", autoscalingv2.SchemeGroupVersion.String(), err)
	}
	hpa := hpaRaw.(*autoscalingv2.HorizontalPodAutoscaler)
	//hpaStatusOriginal := hpa.Status.DeepCopy()

	targetGV, err := schema.ParseGroupVersion(hpa.Spec.ScaleTargetRef.APIVersion)
	if err != nil {
		return fmt.Errorf("invalid API version in scale target reference: %v", err)
	}

	// 转换 targetRef gvk
	targetGK := schema.GroupKind{
		Group: targetGV.Group,
		Kind:  hpa.Spec.ScaleTargetRef.Kind,
	}
	mappings, err := a.mapper.RESTMappings(targetGK)
	if err != nil {
		//a.eventRecorder.Event(hpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
		setCondition(hpa, autoscalingv2.AbleToScale, v1.ConditionFalse, "FailedGetScale", "the HPA controller was unable to get the target's current scale: %v", err)
		//a.updateStatusIfNeeded(hpaStatusOriginal, hpa)
		return fmt.Errorf("unable to determine resource for scale target reference: %v", err)
	}

	// 获取 scale 对象
	scale, targetGR, err := a.scaleForResourceMappings(hpa.Namespace, hpa.Spec.ScaleTargetRef.Name, mappings)
	if err != nil {
		panic(err)
	}

	// 1. 判断是否要扩缩容
	var minReplicas int32
	rescaleReason := ""
	currentReplicas := scale.Spec.Replicas
	desiredReplicas := int32(0)
	if hpa.Spec.MinReplicas != nil {
		minReplicas = *hpa.Spec.MinReplicas
	} else {
		// Default value
		minReplicas = 1
	}
	rescale := true
	if scale.Spec.Replicas == 0 && minReplicas != 0 {
		// Autoscaling is disabled for this resource
		// 业务pod的replicas已经置于0
		desiredReplicas = 0
		rescale = false
		setCondition(hpa, autoscalingv2.ScalingActive, v1.ConditionFalse, "ScalingDisabled", "scaling is disabled since the replica count of the target is zero")
	} else if currentReplicas > hpa.Spec.MaxReplicas {
		rescaleReason = "Current number of replicas above Spec.MaxReplicas"
		desiredReplicas = hpa.Spec.MaxReplicas
	} else if currentReplicas < minReplicas {
		rescaleReason = "Current number of replicas below Spec.MinReplicas"
		desiredReplicas = minReplicas
	} else {
		// 计算期望副本数量，比如3副本pod，目标利用率是20%，实际是10%，HPA(min:1, max:10)，则metricDesiredReplicas=(20/10)*3为6副本
		metricDesiredReplicas, metricName, metricStatuses, metricTimestamp, err := a.computeReplicasForMetrics(hpa, scale, hpa.Spec.Metrics)
		if err != nil {
			panic(err)
		}

		// for test
		klog.Infof(metricName, metricStatuses, metricTimestamp)

		if metricDesiredReplicas > desiredReplicas {
			desiredReplicas = metricDesiredReplicas
			//rescaleMetric = metricName
		}
		if hpa.Spec.Behavior == nil {

		} else {
			desiredReplicas = a.normalizeDesiredReplicasWithBehaviors(hpa, key, currentReplicas, desiredReplicas, minReplicas)
		}

		// 期望副本数量不等于当前副本数量
		rescale = desiredReplicas != currentReplicas
	}

	// 2. 如果扩缩容，实质上是更新 targetRef GroupResource 的 scale 子资源对象
	if rescale {
		// 扩缩容实质上是更新 scales 对象
		scale.Spec.Replicas = desiredReplicas
		_, err = a.scaleNamespacer.Scales(hpa.Namespace).Update(context.TODO(), targetGR, scale, metav1.UpdateOptions{})
		if err != nil {

		}
	} else {

	}

	klog.Infof(rescaleReason)
	return nil
}

// 根据 GroupResource 查找对应的子对象 scale
func (a *HorizontalController) scaleForResourceMappings(namespace, name string, mappings []*apimeta.RESTMapping) (*autoscalingv1.Scale, schema.GroupResource, error) {
	var firstErr error
	for i, mapping := range mappings {
		targetGR := mapping.Resource.GroupResource()
		scale, err := a.scaleNamespacer.Scales(namespace).Get(context.TODO(), targetGR, name, metav1.GetOptions{})
		if err == nil {
			return scale, targetGR, nil
		}

		if i == 0 {
			firstErr = err
		}
	}

	// make sure we handle an empty set of mappings
	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, schema.GroupResource{}, firstErr
}

func (a *HorizontalController) computeReplicasForMetrics(hpa *autoscalingv2.HorizontalPodAutoscaler, scale *autoscalingv1.Scale,
	metricSpecs []autoscalingv2.MetricSpec) (replicas int32, metric string, statuses []autoscalingv2.MetricStatus, timestamp time.Time, err error) {
	selector, err := labels.Parse(scale.Status.Selector)
	if err != nil {
		panic(err)
	}

	/*
		metrics:
			- type: Resource
		      resource:
					name: memory
					target:
						type: Utilization
						averageUtilization: 20
			- type: Resource
				resource:
					name: cpu
		            target:
				      type: Utilization
			          averageUtilization: 20
	*/
	specReplicas := scale.Spec.Replicas
	statusReplicas := scale.Status.Replicas
	statuses = make([]autoscalingv2.MetricStatus, len(metricSpecs))
	for i, metricSpec := range metricSpecs {
		replicaCountProposal, metricNameProposal, timestampProposal, condition, err := a.computeReplicasForMetric(hpa,
			metricSpec, specReplicas, statusReplicas, selector, &statuses[i])
		if err != nil {
			panic(err)
		}
		klog.Info(replicaCountProposal, metricNameProposal, timestampProposal, condition)
	}

	return 0, "", nil, time.Time{}, nil
}

// Computes the desired number of replicas for a specific hpa and metric specification,
// returning the metric status and a proposed condition to be set on the HPA object.
func (a *HorizontalController) computeReplicasForMetric(hpa *autoscalingv2.HorizontalPodAutoscaler,
	spec autoscalingv2.MetricSpec,
	specReplicas, statusReplicas int32,
	selector labels.Selector,
	status *autoscalingv2.MetricStatus) (replicaCountProposal int32, metricNameProposal string,
	timestampProposal time.Time, condition autoscalingv2.HorizontalPodAutoscalerCondition, err error) {

	switch spec.Type {
	case autoscalingv2.ResourceMetricSourceType:
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err = a.computeStatusForResourceMetric(specReplicas, spec, hpa, selector, status)
		if err != nil {
			return 0, "", time.Time{}, condition, err
		}
	}

	return replicaCountProposal, metricNameProposal, timestampProposal, autoscalingv2.HorizontalPodAutoscalerCondition{}, nil
}

func (a *HorizontalController) computeStatusForResourceMetric(currentReplicas int32,
	metricSpec autoscalingv2.MetricSpec,
	hpa *autoscalingv2.HorizontalPodAutoscaler,
	selector labels.Selector,
	status *autoscalingv2.MetricStatus) (replicaCountProposal int32, timestampProposal time.Time, metricNameProposal string,
	condition autoscalingv2.HorizontalPodAutoscalerCondition, err error) {

	if metricSpec.Resource.Target.AverageValue != nil {
		// 忽略 averageValue
		return 0, time.Time{}, "", autoscalingv2.HorizontalPodAutoscalerCondition{}, err
	}
	if metricSpec.Resource.Target.AverageUtilization == nil {
		// 忽略不是 AverageUtilization
		return 0, time.Time{}, "", autoscalingv2.HorizontalPodAutoscalerCondition{}, err
	}
	targetUtilization := *metricSpec.Resource.Target.AverageUtilization
	replicaCountProposal, percentageProposal, rawProposal, timestampProposal, err := a.replicaCalc.GetResourceReplicas(currentReplicas, targetUtilization, metricSpec.Resource.Name, hpa.Namespace, selector)
	if err != nil {
		return 0, time.Time{}, "", autoscalingv2.HorizontalPodAutoscalerCondition{}, err
	}

	klog.Info(percentageProposal, rawProposal)

	return replicaCountProposal, timestampProposal, metricNameProposal, autoscalingv2.HorizontalPodAutoscalerCondition{}, nil
}

// normalizeDesiredReplicasWithBehaviors takes the metrics desired replicas value and normalizes it:
// 1. Apply the basic conditions (i.e. < maxReplicas, > minReplicas, etc...)
// 2. Apply the scale up/down limits from the hpaSpec.Behaviors (i.e. add no more than 4 pods)
// 3. Apply the constraints period (i.e. add no more than 4 pods per minute)
// 4. Apply the stabilization (i.e. add no more than 4 pods per minute, and pick the smallest recommendation during last 5 minutes)
func (a *HorizontalController) normalizeDesiredReplicasWithBehaviors(hpa *autoscalingv2.HorizontalPodAutoscaler, key string,
	currentReplicas, prenormalizedDesiredReplicas, minReplicas int32) int32 {

	normalizationArg := NormalizationArg{
		Key:               key,
		ScaleUpBehavior:   hpa.Spec.Behavior.ScaleUp,
		ScaleDownBehavior: hpa.Spec.Behavior.ScaleDown,
		MinReplicas:       minReplicas,
		MaxReplicas:       hpa.Spec.MaxReplicas,
		CurrentReplicas:   currentReplicas,
		DesiredReplicas:   prenormalizedDesiredReplicas,
	}

	desiredReplicas, reason, message := a.convertDesiredReplicasWithBehaviorRate(normalizationArg)
	klog.Info(reason, message)

	return desiredReplicas
}

type timestampedScaleEvent struct {
	replicaChange int32 // positive for scaleUp, negative for scaleDown
	timestamp     time.Time
	outdated      bool
}
type NormalizationArg struct {
	Key               string
	ScaleUpBehavior   *autoscalingv2.HPAScalingRules
	ScaleDownBehavior *autoscalingv2.HPAScalingRules
	MinReplicas       int32
	MaxReplicas       int32
	CurrentReplicas   int32
	DesiredReplicas   int32
}

func (a *HorizontalController) convertDesiredReplicasWithBehaviorRate(args NormalizationArg) (int32, string, string) {
	var possibleLimitingReason, possibleLimitingMessage string

	if args.DesiredReplicas > args.CurrentReplicas {
		scaleUpLimit := calculateScaleUpLimitWithScalingRules(args.CurrentReplicas, a.scaleUpEvents[args.Key], args.ScaleUpBehavior)
		if scaleUpLimit < args.CurrentReplicas {
			// We shouldn't scale up further until the scaleUpEvents will be cleaned up
			scaleUpLimit = args.CurrentReplicas
		}
		maximumAllowedReplicas := args.MaxReplicas
		if maximumAllowedReplicas > scaleUpLimit {
			maximumAllowedReplicas = scaleUpLimit
			possibleLimitingReason = "ScaleUpLimit"
			possibleLimitingMessage = "the desired replica count is increasing faster than the maximum scale rate"
		} else {
			possibleLimitingReason = "TooManyReplicas"
			possibleLimitingMessage = "the desired replica count is more than the maximum replica count"
		}
		if args.DesiredReplicas > maximumAllowedReplicas {
			return maximumAllowedReplicas, possibleLimitingReason, possibleLimitingMessage
		}
	} else if args.DesiredReplicas < args.CurrentReplicas {

	}

	return args.DesiredReplicas, "DesiredWithinRange", "the desired count is within the acceptable range"
}

// calculateScaleUpLimitWithScalingRules returns the maximum number of pods that could be added for the given HPAScalingRules
// scaleUp behavior policy 来计算需要增加的副本数量
func calculateScaleUpLimitWithScalingRules(currentReplicas int32, scaleEvents []timestampedScaleEvent, scalingRules *autoscalingv2.HPAScalingRules) int32 {

	return 0
}

func setCondition(hpa *autoscalingv2.HorizontalPodAutoscaler, conditionType autoscalingv2.HorizontalPodAutoscalerConditionType,
	status v1.ConditionStatus, reason, message string, args ...interface{}) {
	hpa.Status.Conditions = setConditionInList(hpa.Status.Conditions, conditionType, status, reason, message, args...)
}

func setConditionInList(inputList []autoscalingv2.HorizontalPodAutoscalerCondition, conditionType autoscalingv2.HorizontalPodAutoscalerConditionType,
	status v1.ConditionStatus, reason, message string, args ...interface{}) []autoscalingv2.HorizontalPodAutoscalerCondition {
	resList := inputList
	var existingCond *autoscalingv2.HorizontalPodAutoscalerCondition
	for i, condition := range resList {
		if condition.Type == conditionType {
			// can't take a pointer to an iteration variable
			existingCond = &resList[i]
			break
		}
	}

	if existingCond == nil {
		resList = append(resList, autoscalingv2.HorizontalPodAutoscalerCondition{
			Type: conditionType,
		})
		existingCond = &resList[len(resList)-1]
	}

	if existingCond.Status != status {
		existingCond.LastTransitionTime = metav1.Now()
	}

	existingCond.Status = status
	existingCond.Reason = reason
	existingCond.Message = fmt.Sprintf(message, args...)

	return resList
}
