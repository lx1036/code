package podautoscaler

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"math"
	"time"

	"k8s-lx1036/k8s/monitor/hpa/pkg/podautoscaler/metrics"

	corelisters "k8s.io/client-go/listers/core/v1"
)

type ReplicaCalculator struct {
	metricsClient                 *metrics.RestMetricsClient
	podLister                     corelisters.PodLister
	tolerance                     float64
	cpuInitializationPeriod       time.Duration
	delayOfInitialReadinessStatus time.Duration
}

// NewReplicaCalculator creates a new ReplicaCalculator and passes all necessary information to the new instance
func NewReplicaCalculator(metricsClient *metrics.RestMetricsClient, podLister corelisters.PodLister, tolerance float64, cpuInitializationPeriod, delayOfInitialReadinessStatus time.Duration) *ReplicaCalculator {
	return &ReplicaCalculator{
		metricsClient:                 metricsClient,
		podLister:                     podLister,
		tolerance:                     tolerance,
		cpuInitializationPeriod:       cpuInitializationPeriod,
		delayOfInitialReadinessStatus: delayOfInitialReadinessStatus,
	}
}

// GetResourceReplicas calculates the desired replica count based on a target resource utilization percentage
// of the given resource for pods matching the given selector in the given namespace, and the current replica count
func (c *ReplicaCalculator) GetResourceReplicas(currentReplicas int32, targetUtilization int32,
	resource v1.ResourceName, namespace string, selector labels.Selector) (replicaCount int32,
	utilization int32, rawUtilization int64, timestamp time.Time, err error) {
	resourceMetrics, timestamp, err := c.metricsClient.GetResourceMetric(resource, namespace, selector)
	if err != nil {
		return 0, 0, 0, time.Time{}, fmt.Errorf("unable to get metrics for resource %s: %v", resource, err)
	}

	podList, err := c.podLister.Pods(namespace).List(selector)
	if err != nil {
		return 0, 0, 0, time.Time{}, fmt.Errorf("unable to get pods while calculating replica count: %v", err)
	}
	itemsLen := len(podList)
	if itemsLen == 0 {
		return 0, 0, 0, time.Time{}, fmt.Errorf("no pods returned by selector while calculating replica count")
	}

	// 过滤下坏pods
	readyPodCount, unreadyPods, missingPods, ignoredPods := groupPods(podList, resourceMetrics, resource, c.cpuInitializationPeriod, c.delayOfInitialReadinessStatus)
	removeMetricsForPods(resourceMetrics, ignoredPods)
	removeMetricsForPods(resourceMetrics, unreadyPods)
	requests, err := calculatePodRequests(podList, resource)
	if err != nil {
		return 0, 0, 0, time.Time{}, err
	}

	// 计算 resource 当前实际使用资源量
	usageRatio, utilization, rawUtilization, err := metrics.GetResourceUtilizationRatio(resourceMetrics, requests, targetUtilization)
	if err != nil {
		return 0, 0, 0, time.Time{}, err
	}

	// re-run the utilization calculation with our new numbers
	newUsageRatio, _, _, err := metrics.GetResourceUtilizationRatio(resourceMetrics, requests, targetUtilization)
	if err != nil {
		return 0, utilization, rawUtilization, time.Time{}, err
	}

	// len(resourceMetrics) 是当前pod副本数
	newReplicas := int32(math.Ceil(newUsageRatio * float64(len(resourceMetrics))))

	// return the result, where the number of replicas considered is
	// however many replicas factored into our calculation
	return newReplicas, utilization, rawUtilization, timestamp, nil
}
