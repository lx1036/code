package prometheus

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"

	"k8s.io/klog/v2"
)

type PrometheusHistoryProviderConfig struct {
	Address                             string
	QueryTimeout                        time.Duration
	HistoryLength, HistoryResolution    string
	PodLabelPrefix, PodLabelsMetricName string
	PodNamespaceLabel, PodNameLabel     string

	ContainerNameLabel      string
	ContainerPodNameLabel   string
	ContainerNamespaceLabel string

	CadvisorMetricsJobName string
	Namespace              string
}

// PodHistory represents history of usage and labels for a given pod.
type PodHistory struct {
	// Current samples if pod is still alive, last known samples otherwise.
	LastLabels map[string]string
	LastSeen   time.Time
	// A map for container name to a list of its usage samples, in chronological
	// order.
	Samples map[string][]ContainerUsageSample
}

func newEmptyHistory() *PodHistory {
	return &PodHistory{
		LastLabels: map[string]string{},
		Samples:    map[string][]ContainerUsageSample{},
	}
}

// HistoryProvider gives history of all pods in a cluster.
type HistoryProvider interface {
	GetClusterHistory() (map[PodID]*PodHistory, error)
}

type prometheusHistoryProvider struct {
	prometheusClient  prometheusv1.API
	config            PrometheusHistoryProviderConfig
	queryTimeout      time.Duration
	historyDuration   prommodel.Duration
	historyResolution prommodel.Duration
}

func (provider *prometheusHistoryProvider) GetClusterHistory() (map[PodID]*PodHistory, error) {
	clusterHistory := make(map[PodID]*PodHistory)

	var podSelector string
	// `job="kubernetes-cadvisor", `
	if provider.config.CadvisorMetricsJobName != "" {
		podSelector = fmt.Sprintf(`job="%s", `, provider.config.CadvisorMetricsJobName)
	}
	// `pod_name=~".+", name!="POD", name!=""`
	podSelector = podSelector + fmt.Sprintf(`%s=~".+", %s!="POD", %s!=""`,
		provider.config.ContainerPodNameLabel, provider.config.ContainerNameLabel, provider.config.ContainerNameLabel)
	if provider.config.Namespace != "" {
		// `job="kubernetes-cadvisor", pod_name=~".+", name!="POD", name!="", namespace="default"`
		podSelector = fmt.Sprintf(`%s, %s="%s"`, podSelector, provider.config.ContainerNamespaceLabel, provider.config.Namespace)
	}

	// INFO: https://github.com/google/cadvisor/blob/v0.39.2/metrics/prometheus.go#L163-L188
	//  rate(container_cpu_usage_seconds_total{name="redis"}[1m]): Cumulative cpu time consumed in seconds.
	// INFO: historicalCpuQuery=`rate(container_cpu_usage_seconds_total{job="cadvisor", pod=~".+", name!="POD", name!="", namespace="cattle-system"}[1h])`
	historicalCpuQuery := fmt.Sprintf("rate(container_cpu_usage_seconds_total{%s}[%s])", podSelector, provider.config.HistoryResolution)
	klog.V(2).Infof("Historical CPU usage query used: %s", historicalCpuQuery)
	err := provider.readResourceHistory(clusterHistory, historicalCpuQuery, ResourceCPU)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}

	// INFO: https://github.com/google/cadvisor/blob/v0.39.2/metrics/prometheus.go#L419-L425
	//  container_memory_working_set_bytes: Current working set in bytes.
	// INFO: historicalMemoryQuery=`container_memory_working_set_bytes{job="cadvisor", pod=~".+", name!="POD", name!="", namespace="cattle-system"}`
	historicalMemoryQuery := fmt.Sprintf("container_memory_working_set_bytes{%s}", podSelector)
	klog.V(4).Infof("Historical memory usage query used: %s", historicalMemoryQuery)
	err = provider.readResourceHistory(clusterHistory, historicalMemoryQuery, ResourceMemory)
	if err != nil {
		return nil, fmt.Errorf("cannot get usage history: %v", err)
	}

	// sort by MeasureStart
	for _, podHistory := range clusterHistory {
		for _, containerUsageSample := range podHistory.Samples {
			sort.Slice(containerUsageSample, func(i, j int) bool {
				return containerUsageSample[i].MeasureStart.Before(containerUsageSample[j].MeasureStart)
			})
		}
	}

	// INFO: up{job="kubernetes-custom"}
	provider.readLastLabels(clusterHistory, provider.config.PodLabelsMetricName)
	return clusterHistory, nil
}

func (provider *prometheusHistoryProvider) readResourceHistory(clusterHistory map[PodID]*PodHistory, query string, resource ResourceName) error {
	end := time.Now()
	start := end.Add(-time.Duration(provider.historyDuration))

	ctx, cancel := context.WithTimeout(context.Background(), provider.queryTimeout)
	defer cancel()

	// INFO: query=`rate(container_cpu_usage_seconds_total{job="cadvisor", pod=~".+", name!="POD", name!="", namespace="cattle-system"}[1h])`
	result, _, err := provider.prometheusClient.QueryRange(ctx, query, prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  time.Duration(provider.historyResolution),
	})
	if err != nil {
		return fmt.Errorf("cannot get timeseries for %v: %v", resource, err)
	}

	matrix, ok := result.(prommodel.Matrix)
	if !ok {
		return fmt.Errorf("expected query to return a matrix; got result type %T", result)
	}

	for _, sampleStream := range matrix {
		containerID, err := provider.getContainerIDFromLabels(sampleStream.Metric)
		if err != nil {
			return fmt.Errorf("cannot get container ID from labels: %v", sampleStream.Metric)
		}

		newSamples := getContainerUsageSamplesFromSamples(sampleStream.Values, resource)
		podHistory, ok := clusterHistory[containerID.PodID]
		if !ok {
			podHistory = newEmptyHistory()
			clusterHistory[containerID.PodID] = podHistory
		}
		podHistory.Samples[containerID.ContainerName] = append(
			podHistory.Samples[containerID.ContainerName],
			newSamples...)
	}

	return nil
}

func (provider *prometheusHistoryProvider) getContainerIDFromLabels(metric prommodel.Metric) (*ContainerID, error) {
	labels := promMetricToLabelMap(metric)
	namespace, ok := labels[provider.config.ContainerNamespaceLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label", provider.config.ContainerNamespaceLabel)
	}
	podName, ok := labels[provider.config.ContainerPodNameLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label", provider.config.ContainerPodNameLabel)
	}
	containerName, ok := labels[provider.config.ContainerNameLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label on container data", provider.config.ContainerNameLabel)
	}
	return &ContainerID{
		PodID: PodID{
			Namespace: namespace,
			PodName:   podName,
		},
		ContainerName: containerName,
	}, nil
}

func (provider *prometheusHistoryProvider) readLastLabels(clusterHistory map[PodID]*PodHistory, query string) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.queryTimeout)
	defer cancel()

	result, _, err := provider.prometheusClient.Query(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("cannot get timeseries for labels: %v", err)
	}

	matrix, ok := result.(prommodel.Matrix)
	if !ok {
		return fmt.Errorf("expected query to return a matrix; got result type %T", result)
	}

	for _, sampleStream := range matrix {
		podID, err := provider.getPodIDFromLabels(sampleStream.Metric)
		if err != nil {
			return fmt.Errorf("cannot get container ID from labels %v: %v", sampleStream.Metric, err)
		}
		podHistory, ok := clusterHistory[*podID]
		if !ok {
			podHistory = newEmptyHistory()
			clusterHistory[*podID] = podHistory
		}
		podLabels := provider.getPodLabelsMap(sampleStream.Metric)

		// time series results will always be sorted chronologically from oldest to
		// newest, so the last element is the latest sample
		lastSample := sampleStream.Values[len(sampleStream.Values)-1]
		if lastSample.Timestamp.Time().After(podHistory.LastSeen) {
			podHistory.LastSeen = lastSample.Timestamp.Time()
			podHistory.LastLabels = podLabels
		}
	}

	return nil
}

func (provider *prometheusHistoryProvider) getPodIDFromLabels(metric prommodel.Metric) (*PodID, error) {
	labels := promMetricToLabelMap(metric)
	namespace, ok := labels[provider.config.PodNamespaceLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label", provider.config.PodNamespaceLabel)
	}
	podName, ok := labels[provider.config.PodNameLabel]
	if !ok {
		return nil, fmt.Errorf("no %s label", provider.config.PodNameLabel)
	}
	return &PodID{Namespace: namespace, PodName: podName}, nil
}

func (provider *prometheusHistoryProvider) getPodLabelsMap(metric prommodel.Metric) map[string]string {
	podLabels := make(map[string]string)
	for key, value := range metric {
		podLabelKey := strings.TrimPrefix(string(key), provider.config.PodLabelPrefix)
		if podLabelKey != string(key) {
			podLabels[podLabelKey] = string(value)
		}
	}

	return podLabels
}

func NewPrometheusHistoryProvider(config PrometheusHistoryProviderConfig) (HistoryProvider, error) {
	promClient, err := promapi.NewClient(promapi.Config{
		Address: config.Address,
	})
	if err != nil {
		return &prometheusHistoryProvider{}, err
	}

	// Use Prometheus's model.Duration; this can additionally parse durations in days, weeks and years (as well as seconds, minutes, hours etc)
	historyDuration, err := prommodel.ParseDuration(config.HistoryLength)
	if err != nil {
		return &prometheusHistoryProvider{}, fmt.Errorf("history length %s is not a valid Prometheus duration: %v", config.HistoryLength, err)
	}

	historyResolution, err := prommodel.ParseDuration(config.HistoryResolution)
	if err != nil {
		return &prometheusHistoryProvider{}, fmt.Errorf("history resolution %s is not a valid Prometheus duration: %v", config.HistoryResolution, err)
	}

	return &prometheusHistoryProvider{
		prometheusClient:  prometheusv1.NewAPI(promClient),
		config:            config,
		queryTimeout:      config.QueryTimeout,
		historyDuration:   historyDuration,
		historyResolution: historyResolution,
	}, nil
}
