package scraper

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type Scraper struct {
	metricsClient *v1beta1.MetricsV1beta1Client
}

type PodMetrics map[*v1.Pod]v1.ResourceList

// 计算在离线 pod 的资源使用总和
func (scraper *Scraper) Scrape(baseCtx context.Context, pods []*v1.Pod) (v1.ResourceList, v1.ResourceList, error) {
	prodTotalUsageResource, nonProdTotalUsageResource := v1.ResourceList{
		v1.ResourceCPU:    *resource.NewQuantity(0, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(0, resource.BinarySI),
	}, v1.ResourceList{
		v1.ResourceCPU:    *resource.NewQuantity(0, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(0, resource.BinarySI),
	}

	var errs []error

	responseChannel := make(chan PodMetrics, len(pods))
	errChannel := make(chan error, len(pods))
	defer close(responseChannel)
	defer close(errChannel)
	for _, pod := range pods {
		go func(pod *v1.Pod) {
			ctx, cancelTimeout := context.WithTimeout(baseCtx, scraper.scrapeTimeout)
			defer cancelTimeout()

			metrics, err := scraper.collectPodMetrics(ctx, pod)
			if err != nil {
				err = fmt.Errorf("unable to fully scrape metrics from node %s: %v", node.Name, err)
			}

			podMetrics := make(PodMetrics)
			podMetrics[pod] = metrics
			responseChannel <- podMetrics
			errChannel <- err
		}(pod)
	}

	for range pods {
		err := <-errChannel
		podMetrics := <-responseChannel
		if err != nil {
			errs = append(errs, err)
		}
		if podMetrics == nil {
			continue
		}

		for pod, metrics := range podMetrics {
			if isProd(pod) {
				cpu := prodTotalUsageResource[v1.ResourceCPU]
				cpu.Add(metrics[v1.ResourceCPU])
				prodTotalUsageResource[v1.ResourceCPU] = cpu
				memory := prodTotalUsageResource[v1.ResourceMemory]
				memory.Add(metrics[v1.ResourceMemory])
				prodTotalUsageResource[v1.ResourceMemory] = memory
			} else if isNonProd(pod) {
				cpu := nonProdTotalUsageResource[v1.ResourceCPU]
				cpu.Add(metrics[v1.ResourceCPU])
				nonProdTotalUsageResource[v1.ResourceCPU] = cpu
				memory := nonProdTotalUsageResource[v1.ResourceMemory]
				memory.Add(metrics[v1.ResourceMemory])
				nonProdTotalUsageResource[v1.ResourceMemory] = memory
			} else {
				// TODO
			}
		}
	}

	return prodTotalUsageResource, nonProdTotalUsageResource, utilerrors.NewAggregate(errs)
}

func (scraper *Scraper) collectPodMetrics(ctx context.Context, pod *v1.Pod) (v1.ResourceList, error) {
	podMetrics, err := scraper.metricsClient.PodMetricses(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		//klog.Errorf("fail to get pod %s/%s metrics: %v", pod.Namespace, pod.Name, err)
		return nil, fmt.Errorf("")
	}

	podUsageResource := v1.ResourceList{
		v1.ResourceCPU:    *resource.NewQuantity(0, resource.DecimalSI),
		v1.ResourceMemory: *resource.NewQuantity(0, resource.BinarySI),
	}
	for _, containerMetrics := range podMetrics.Containers {
		cpu := podUsageResource[v1.ResourceCPU]
		cpu.Add(*containerMetrics.Usage.Cpu())
		podUsageResource[v1.ResourceCPU] = cpu
		memory := podUsageResource[v1.ResourceMemory]
		memory.Add(*containerMetrics.Usage.Memory())
		podUsageResource[v1.ResourceMemory] = memory
	}

	return podUsageResource, nil
}

func NewScraper(metricsClient *v1beta1.MetricsV1beta1Client) *Scraper {
	return &Scraper{
		metricsClient: metricsClient,
	}
}
