package server

import (
	"context"
	"fmt"
	"time"

	"k8s-lx1036/k8s/monitor/metrcis-server/pkg/api"
	"k8s-lx1036/k8s/monitor/metrcis-server/pkg/scraper"
	"k8s-lx1036/k8s/monitor/metrcis-server/pkg/storage"

	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type Config struct {
	MetricResolution time.Duration
	ScrapeTimeout    time.Duration

	ApiserverConfig *genericapiserver.Config

	Kubeconfig string
}

// server scrapes metrics and serves then using k8s api.
type Server struct {
	*genericapiserver.GenericAPIServer

	// The resolution at which metrics-server will retain metrics
	resolution time.Duration

	scraper *scraper.Scraper
	storage *storage.Storage

	sync     cache.InformerSynced
	informer informers.SharedInformerFactory
}

func NewServer(config *Config) (*Server, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}
	// we should never need to resync, since we're not worried about missing events,
	// and resync is actually for regular interval-based reconciliation these days,
	// so set the default resync interval to 0
	informer := informers.NewSharedInformerFactory(kubeClient, 0)
	nodes := informer.Core().V1().Nodes()
	kubeletClient, err := scraper.NewKubeletClient(restConfig)
	if err != nil {
		return nil, err
	}

	// INFO: 这里得研究下 metrics-server 是怎么通过 api-aggregator 来扩展 apiserver 的
	// https://kubernetes.io/zh/docs/tasks/extend-kubernetes/configure-aggregation-layer/
	// Disable default metrics handler and create custom one
	config.ApiserverConfig.EnableMetrics = false
	genericServer, err := config.ApiserverConfig.Complete(informer).New("metrics-server", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	store := storage.NewStorage()
	if err := api.Install(store, informer.Core().V1(), genericServer); err != nil {
		return nil, err
	}

	return &Server{
		sync:             nodes.Informer().HasSynced,
		informer:         informer,
		GenericAPIServer: genericServer,
		storage:          store,
		scraper:          scraper.NewScraper(nodes.Lister(), kubeletClient, config.ScrapeTimeout),
		resolution:       config.MetricResolution,
	}, nil
}

// RunUntil starts background scraping goroutine and runs apiserver serving metrics.
func (server *Server) RunUntil(stopCh <-chan struct{}) error {
	server.informer.Start(stopCh)
	shutdown := cache.WaitForCacheSync(stopCh, server.sync)
	if !shutdown {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		server.tick(ctx, time.Now())
	}, server.resolution)

	// 这段代码啥意思？？？
	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}

func (server *Server) tick(ctx context.Context, startTime time.Time) {
	data, scrapeErr := server.scraper.Scrape(ctx)
	if scrapeErr != nil {
		klog.Errorf("unable to fully scrape metrics: %v", scrapeErr)
		return
	}

	klog.Infof("...Storing metrics...")
	server.storage.Store(data)
}
