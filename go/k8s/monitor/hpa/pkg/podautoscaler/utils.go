package podautoscaler

import (
	"fmt"

	"k8s-lx1036/k8s/monitor/hpa/pkg/podautoscaler/metrics"

	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	utilpointer "k8s.io/utils/pointer"
)

var (
	Scheme = runtime.NewScheme()

	// Codecs provides access to encoding and decoding for the scheme
	Codecs = serializer.NewCodecFactory(Scheme)

	// ParameterCodec handles versioning of objects that are converted to query parameters.
	ParameterCodec = runtime.NewParameterCodec(Scheme)
)

func calculatePodRequests(pods []*v1.Pod, resource v1.ResourceName) (map[string]int64, error) {
	requests := make(map[string]int64, len(pods))
	for _, pod := range pods {
		podSum := int64(0)
		for _, container := range pod.Spec.Containers {
			if containerRequest, ok := container.Resources.Requests[resource]; ok {
				podSum += containerRequest.MilliValue()
			} else {
				return nil, fmt.Errorf("missing request for %s", resource)
			}
		}
		requests[pod.Name] = podSum
	}

	return requests, nil
}

func removeMetricsForPods(metrics metrics.PodMetricsInfo, pods sets.String) {
	for _, pod := range pods.UnsortedList() {
		delete(metrics, pod)
	}
}

func UnsafeConvertToVersionVia(obj runtime.Object, externalVersion schema.GroupVersion) (runtime.Object, error) {
	objInt, err := Scheme.UnsafeConvertToVersion(obj, schema.GroupVersion{Group: externalVersion.Group, Version: runtime.APIVersionInternal})
	if err != nil {
		return nil, fmt.Errorf("failed to convert the given object to the internal version: %v", err)
	}

	objExt, err := Scheme.UnsafeConvertToVersion(objInt, externalVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to convert the given object back to the external version: %v", err)
	}

	return objExt, err
}

func generateScalingRules(pods, podsPeriod, percent, percentPeriod, stabilizationWindow int32) *autoscalingv2.HPAScalingRules {
	policy := autoscalingv2.MaxPolicySelect
	directionBehavior := autoscalingv2.HPAScalingRules{
		StabilizationWindowSeconds: utilpointer.Int32Ptr(stabilizationWindow),
		SelectPolicy:               &policy,
	}
	if pods != 0 {
		directionBehavior.Policies = append(directionBehavior.Policies,
			autoscalingv2.HPAScalingPolicy{Type: autoscalingv2.PodsScalingPolicy, Value: pods, PeriodSeconds: podsPeriod})
	}
	if percent != 0 {
		directionBehavior.Policies = append(directionBehavior.Policies,
			autoscalingv2.HPAScalingPolicy{Type: autoscalingv2.PercentScalingPolicy, Value: percent, PeriodSeconds: percentPeriod})
	}
	return &directionBehavior
}
