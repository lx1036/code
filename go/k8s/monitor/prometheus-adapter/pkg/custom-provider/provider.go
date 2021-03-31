package custom_provider

import (
	"context"
	"fmt"
	"time"

	"github.com/kubernetes-sigs/custom-metrics-apiserver/pkg/provider"
	"github.com/prometheus/common/model"
	"k8s-lx1036/k8s/monitor/prometheus-adapter/pkg/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/custom_metrics"
)

// Runnable represents something that can be run until told to stop.
type Runnable interface {
	// Run runs the runnable forever.
	Run()
	// RunUntil runs the runnable until the given channel is closed.
	RunUntil(stopChan <-chan struct{})
}

type prometheusProvider struct {
	mapper     meta.RESTMapper
	kubeClient dynamic.Interface
	promClient client.Client

	SeriesRegistry
}

func (p *prometheusProvider) GetMetricByName(name types.NamespacedName, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	panic("implement me")
}

func (p *prometheusProvider) GetMetricBySelector(namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	panic("implement me")
}

func (p *prometheusProvider) ListAllMetrics() []provider.CustomMetricInfo {
	panic("implement me")
}

func NewPrometheusProvider(mapper meta.RESTMapper,
	kubeClient dynamic.Interface,
	promClient client.Client,
	namers []naming.MetricNamer,
	updateInterval time.Duration,
	maxAge time.Duration) (provider.CustomMetricsProvider, Runnable) {

	lister := &cachingMetricsLister{
		updateInterval: updateInterval,
		maxAge:         maxAge,
		promClient:     promClient,
		namers:         namers,

		SeriesRegistry: &basicSeriesRegistry{
			mapper: mapper,
		},
	}

	return &prometheusProvider{
		mapper:     mapper,
		kubeClient: kubeClient,
		promClient: promClient,

		SeriesRegistry: lister,
	}, lister
}

// 从 prometheus api中读取数据，并缓存起来，供 HPA customMetricsClient 消费读取数据
type cachingMetricsLister struct {
	SeriesRegistry

	promClient     client.Client
	updateInterval time.Duration
	maxAge         time.Duration
	namers         []naming.MetricNamer
}

func (l *cachingMetricsLister) Run() {
	l.RunUntil(wait.NeverStop)
}

// 缓存数据，并周期性更新
func (l *cachingMetricsLister) RunUntil(stopChan <-chan struct{}) {
	go wait.Until(func() {
		if err := l.updateMetrics(); err != nil {
			utilruntime.HandleError(err)
		}
	}, l.updateInterval, stopChan)
}

// 缓存数据，并周期性更新
func (l *cachingMetricsLister) updateMetrics() error {
	startTime := model.Now().Add(-1 * l.maxAge)
	// don't do duplicate queries when it's just the matchers that change
	seriesCacheByQuery := make(map[client.Selector][]client.Series)
	// these can take a while on large clusters, so launch in parallel
	// and don't duplicate
	selectors := make(map[client.Selector]struct{})
	selectorSeriesChan := make(chan selectorSeries, len(l.namers))
	errs := make(chan error, len(l.namers))

	// 1. 并发读取prometheus api数据
	for _, namer := range l.namers {
		sel := namer.Selector()
		if _, ok := selectors[sel]; ok {
			errs <- nil
			selectorSeriesChan <- selectorSeries{}
			continue
		}
		selectors[sel] = struct{}{}
		go func() {
			series, err := l.promClient.Series(context.TODO(), model.Interval{startTime, 0}, sel)
			if err != nil {
				errs <- fmt.Errorf("unable to fetch metrics for query %q: %v", sel, err)
				return
			}
			errs <- nil
			selectorSeriesChan <- selectorSeries{
				selector: sel,
				series:   series,
			}
		}()
	}
	// iterate through, blocking until we've got all results
	for range l.namers {
		if err := <-errs; err != nil {
			return fmt.Errorf("unable to update list of all metrics: %v", err)
		}
		if ss := <-selectorSeriesChan; ss.series != nil {
			seriesCacheByQuery[ss.selector] = ss.series
		}
	}
	close(errs)

	// 2. 缓存数据
	newSeries := make([][]client.Series, len(l.namers))
	for i, namer := range l.namers {
		series, cached := seriesCacheByQuery[namer.Selector()]
		if !cached {
			return fmt.Errorf("unable to update list of all metrics: no metrics retrieved for query %q", namer.Selector())
		}
		newSeries[i] = namer.FilterSeries(series)
	}

	klog.Infof("Set available metric list from Prometheus to: %v", newSeries)

	return l.SetSeries(newSeries, l.namers)
}
