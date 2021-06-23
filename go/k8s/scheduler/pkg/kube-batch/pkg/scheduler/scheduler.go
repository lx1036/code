package scheduler

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// Scheduler watches for new unscheduled pods for kubebatch. It attempts to find
// nodes that they fit on and writes bindings back to the api server.
type Scheduler struct {
	cache          schedcache.Cache
	config         *rest.Config
	actions        []framework.Action
	plugins        []conf.Tier
	schedulerConf  string
	schedulePeriod time.Duration
}

// NewScheduler returns a scheduler
func NewScheduler(config *rest.Config, schedulerName string, conf string, period time.Duration,
	defaultQueue string) (*Scheduler, error) {
	scheduler := &Scheduler{
		config:         config,
		schedulerConf:  conf,
		cache:          schedcache.New(config, schedulerName, defaultQueue),
		schedulePeriod: period,
	}

	return scheduler, nil
}

// Run runs the Scheduler
func (pc *Scheduler) Run(stopCh <-chan struct{}) {
	var err error

	// Start cache for policy.
	go pc.cache.Run(stopCh)
	pc.cache.WaitForCacheSync(stopCh)

	// Load configuration of scheduler
	schedConf := defaultSchedulerConf
	if len(pc.schedulerConf) != 0 {
		if schedConf, err = readSchedulerConf(pc.schedulerConf); err != nil {
			klog.Errorf("Failed to read scheduler configuration '%s', using default configuration: %v",
				pc.schedulerConf, err)
			schedConf = defaultSchedulerConf
		}
	}

	pc.actions, pc.plugins, err = loadSchedulerConf(schedConf)
	if err != nil {
		panic(err)
	}

	go wait.Until(pc.runOnce, pc.schedulePeriod, stopCh)
}

func (pc *Scheduler) runOnce() {
	klog.V(4).Infof("Start scheduling ...")
	scheduleStartTime := time.Now()
	defer klog.V(4).Infof("End scheduling ...")
	defer metrics.UpdateE2eDuration(metrics.Duration(scheduleStartTime))

	ssn := framework.OpenSession(pc.cache, pc.plugins)
	defer framework.CloseSession(ssn)

	for _, action := range pc.actions {
		actionStartTime := time.Now()
		action.Execute(ssn)
		metrics.UpdateActionDuration(action.Name(), metrics.Duration(actionStartTime))
	}
}
