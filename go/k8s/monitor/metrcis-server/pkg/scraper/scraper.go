package scraper

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/monitor/metrcis-server/pkg/storage"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

type Scraper struct {
	nodeLister    v1listers.NodeLister
	kubeletClient *KubeletClient
	scrapeTimeout time.Duration // scrape timeout is 90% of the scrape interval
}

func NewScraper(nodeLister v1listers.NodeLister, client *KubeletClient, scrapeTimeout time.Duration) *Scraper {
	return &Scraper{
		nodeLister:    nodeLister,
		kubeletClient: client,
		scrapeTimeout: scrapeTimeout,
	}
}

func (scraper *Scraper) Scrape(baseCtx context.Context) (*storage.MetricsBatch, error) {
	var errs []error
	nodes, err := scraper.nodeLister.List(labels.Everything())
	if err != nil {
		errs = append(errs, err)
	}

	klog.Infof("Scraping metrics from %v nodes", len(nodes))

	responseChannel := make(chan *storage.MetricsBatch, len(nodes))
	errChannel := make(chan error, len(nodes))
	defer close(responseChannel)
	defer close(errChannel)

	// TODO: 这个通过 channel 并发 collect metrics 数据方式值得抄抄
	res := &storage.MetricsBatch{}
	for _, node := range nodes {
		go func(node *corev1.Node) {
			ctx, cancelTimeout := context.WithTimeout(baseCtx, scraper.scrapeTimeout)
			defer cancelTimeout()
			metrics, err := scraper.collectNode(ctx, node)
			if err != nil {
				err = fmt.Errorf("unable to fully scrape metrics from node %s: %v", node.Name, err)
			}
			responseChannel <- metrics
			errChannel <- err
		}(node)
	}

	for range nodes {
		err := <-errChannel
		srcBatch := <-responseChannel
		if err != nil {
			errs = append(errs, err)
		}
		if srcBatch == nil {
			continue
		}

		res.Nodes = append(res.Nodes, srcBatch.Nodes...)
		res.Pods = append(res.Pods, srcBatch.Pods...)
	}

	return res, utilerrors.NewAggregate(errs)
}

func (scraper *Scraper) collectNode(ctx context.Context, node *corev1.Node) (*storage.MetricsBatch, error) {
	summary, err := scraper.kubeletClient.GetSummary(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch metrics from node %s: %v", node.Name, err)
	}

	return decodeBatch(summary), nil
}
