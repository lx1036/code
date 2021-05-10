package scraper

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/utils"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type Scraper struct {
	metricsClient *v1beta1.MetricsV1beta1Client
}

type PodMetrics map[*v1.Pod]v1.ResourceList

// Scrape calculate prodPod/nonProdPod sum of resources
func (scraper *Scraper) Scrape(baseCtx context.Context, pods []*v1.Pod) (v1.ResourceList, v1.ResourceList, error) {
	var errs []error

	responseChannel := make(chan PodMetrics, len(pods))
	errChannel := make(chan error, len(pods))
	defer close(responseChannel)
	defer close(errChannel)
	for _, pod := range pods {
		go func(pod *v1.Pod) {
			ctx, cancelTimeout := context.WithTimeout(baseCtx, time.Second*10)
			defer cancelTimeout()

			metrics, err := scraper.collectPodMetrics(ctx, pod)
			podMetrics := make(PodMetrics)
			podMetrics[pod] = metrics
			responseChannel <- podMetrics
			errChannel <- err
		}(pod)
	}

	prodTotalUsageResource, nonProdTotalUsageResource := make(v1.ResourceList), make(v1.ResourceList)

	for range pods {
		err := <-errChannel
		podMetrics := <-responseChannel
		if err != nil {
			errs = append(errs, err)
		}
		if podMetrics == nil || len(podMetrics) == 0 {
			continue
		}

		for pod, metrics := range podMetrics {
			if utils.IsProdPod(pod) {
				for name := range metrics {
					value := prodTotalUsageResource[name]
					value.Add(metrics[name])
					if !value.IsZero() {
						prodTotalUsageResource[name] = value
					}
				}
			} else if utils.IsNonProdPod(pod) {
				for name := range metrics {
					value := nonProdTotalUsageResource[name]
					value.Add(metrics[name])
					if !value.IsZero() {
						nonProdTotalUsageResource[name] = value
					}
				}
			} else {
				// do something
			}
		}
	}

	return prodTotalUsageResource, nonProdTotalUsageResource, utilerrors.NewAggregate(errs)
}

func (scraper *Scraper) collectPodMetrics(ctx context.Context, pod *v1.Pod) (v1.ResourceList, error) {
	podMetrics, err := scraper.metricsClient.PodMetricses(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("fail to get pod %s/%s metrics: %v", pod.Namespace, pod.Name, err))
	}

	podUsageResource := make(v1.ResourceList)
	for _, containerMetrics := range podMetrics.Containers {
		for name := range containerMetrics.Usage {
			value := podUsageResource[name]
			value.Add(containerMetrics.Usage[name])
			if !value.IsZero() {
				podUsageResource[name] = value
			}
		}
	}

	return podUsageResource, nil
}

func NewScraper(metricsClient *v1beta1.MetricsV1beta1Client) *Scraper {
	return &Scraper{
		metricsClient: metricsClient,
	}
}
