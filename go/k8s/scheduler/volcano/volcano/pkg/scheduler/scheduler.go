package scheduler

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"path/filepath"
	"time"

	"k8s-lx1036/k8s/scheduler/volcano/volcano/cmd/scheduler/app/options"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/filewatcher"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/cache"

	"k8s.io/client-go/rest"
)

// Scheduler watches for new unscheduled pods(PodGroup) in Volcano.
// It attempts to find nodes that can accommodate these pods and writes the binding information back to the API server.
type Scheduler struct {
	schedulerConf  string
	schedulePeriod time.Duration

	cache  cache.Cache
	dumper cache.Dumper

	fileWatcher filewatcher.FileWatcher
}

func (scheduler *Scheduler) Run(stopCh <-chan struct{}) {
	// load and watch conf
	scheduler.loadSchedulerConf()
	go scheduler.watchSchedulerConf(stopCh)

	scheduler.cache.SetMetricsConf(scheduler.metricsConf)
	scheduler.cache.Run(stopCh)
	scheduler.cache.WaitForCacheSync(stopCh)
	klog.V(2).Infof("Scheduler completes Initialization and start to run")

	go wait.Until(scheduler.runOnce, scheduler.schedulePeriod, stopCh)

	if options.ServerOpts.EnableCacheDumper {
		scheduler.dumper.ListenForSignal(stopCh)
	}

	go runSchedulerSocket()
}

func (scheduler *Scheduler) loadSchedulerConf() {

}

func NewScheduler(config *rest.Config, opt *options.ServerOption) (*Scheduler, error) {
	var watcher filewatcher.FileWatcher
	if opt.SchedulerConf != "" { // watch dir not specified file
		var err error
		path := filepath.Dir(opt.SchedulerConf)
		watcher, err = filewatcher.NewFileWatcher(path)
		if err != nil {
			return nil, fmt.Errorf("failed creating filewatcher for %s: %v", opt.SchedulerConf, err)
		}
	}

	c := cache.New(config, opt.SchedulerNames, opt.DefaultQueue, opt.NodeSelector,
		opt.NodeWorkerThreads, opt.IgnoredCSIProvisioners)
	scheduler := &Scheduler{
		cache:          c,
		schedulerConf:  opt.SchedulerConf,
		schedulePeriod: opt.SchedulePeriod,
		dumper:         cache.Dumper{Cache: c, RootDir: opt.CacheDumpFileDir},
		fileWatcher:    watcher,
	}

	return scheduler, nil
}
