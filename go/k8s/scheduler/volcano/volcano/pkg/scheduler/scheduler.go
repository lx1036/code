package scheduler

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/scheduler/volcano/volcano/cmd/scheduler/app/options"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/filewatcher"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/cache"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/conf"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/framework"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/metrics"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/plugins"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
)

var DefaultSchedulerConf = `
actions: "enqueue, allocate, backfill"
tiers:
- plugins:
  - name: priority
  - name: gang
  - name: conformance
- plugins:
  - name: overcommit
  - name: drf
  - name: predicates
  - name: proportion
  - name: nodeorder
`

// Scheduler watches for new unscheduled pods(PodGroup) in Volcano.
// It attempts to find nodes that can accommodate these pods and writes the binding information back to the API server.
type Scheduler struct {
	once  sync.Once
	mutex sync.Mutex

	schedulerConf  string
	schedulePeriod time.Duration
	metricsConf    map[string]string

	cache  cache.Cache
	dumper cache.Dumper

	actions        []framework.Action
	plugins        []conf.Tier
	configurations []conf.Configuration

	fileWatcher filewatcher.FileWatcher
}

func (scheduler *Scheduler) Run(stopCh <-chan struct{}) {
	// load and watch conf
	scheduler.loadSchedulerConf()
	go scheduler.watchSchedulerConf(stopCh)

	scheduler.cache.SetMetricsConf(scheduler.metricsConf)
	scheduler.cache.Run(stopCh)
	klog.V(2).Infof("Scheduler completes Initialization and start to run")

	go wait.Until(scheduler.runOnce, scheduler.schedulePeriod, stopCh) // 1s

	if options.ServerOpts.EnableCacheDumper {
		scheduler.dumper.ListenForSignal(stopCh)
	}
}

// runOnce executes a single scheduling cycle. This function is called periodically
// as defined by the Scheduler's schedule period.
func (scheduler *Scheduler) runOnce() {
	klog.V(4).Infof("Start scheduling ...")
	scheduleStartTime := time.Now()
	defer klog.V(4).Infof("End scheduling ...")

	scheduler.mutex.Lock()
	actions := scheduler.actions
	tiers := scheduler.plugins
	configurations := scheduler.configurations
	scheduler.mutex.Unlock()

	// TODO: 在 allocate action 里使用。有些鸡肋，需要重构!!!
	// Load ConfigMap to check which action is enabled.
	conf.EnabledActionMap = make(map[string]bool)
	for _, action := range actions {
		conf.EnabledActionMap[action.Name()] = true
	}

	ssn := framework.OpenSession(scheduler.cache, tiers, configurations)
	defer func() {
		framework.CloseSession(ssn)
		metrics.UpdateE2eDuration(metrics.Duration(scheduleStartTime))
	}()

	for _, action := range actions {
		actionStartTime := time.Now()
		action.Execute(ssn)
		metrics.UpdateActionDuration(action.Name(), metrics.Duration(actionStartTime))
	}
}

func (scheduler *Scheduler) loadSchedulerConf() {
	klog.V(4).Infof("Start loadSchedulerConf ...")
	defer func() {
		actions, plgs := scheduler.getSchedulerConf()
		klog.V(2).Infof("Successfully loaded Scheduler conf, actions: %v, plugins: %v", actions, plgs)
	}()

	var err error
	scheduler.once.Do(func() {
		scheduler.actions, scheduler.plugins, scheduler.configurations, scheduler.metricsConf, err = UnmarshalSchedulerConf(DefaultSchedulerConf)
		if err != nil {
			klog.Errorf("unmarshal Scheduler config %s failed: %v", DefaultSchedulerConf, err)
			panic("invalid default configuration")
		}
	})

	var config string
	if len(scheduler.schedulerConf) != 0 {
		confData, err := os.ReadFile(scheduler.schedulerConf)
		if err != nil {
			klog.Errorf("Failed to read the Scheduler config in '%s', using previous configuration: %v",
				scheduler.schedulerConf, err)
			return
		}
		config = strings.TrimSpace(string(confData))
	}
	actions, plgs, configurations, metricsConf, err := UnmarshalSchedulerConf(config)
	if err != nil {
		klog.Errorf("Scheduler config %s is invalid: %v", config, err)
		return
	}

	scheduler.mutex.Lock()
	scheduler.actions = actions
	scheduler.plugins = plgs
	scheduler.configurations = configurations
	scheduler.metricsConf = metricsConf
	scheduler.mutex.Unlock()
}

func (scheduler *Scheduler) watchSchedulerConf(stopCh <-chan struct{}) {
	if scheduler.fileWatcher == nil {
		return
	}
	eventCh := scheduler.fileWatcher.Events()
	errCh := scheduler.fileWatcher.Errors()
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				return
			}
			klog.V(4).Infof("watch %s event: %v", scheduler.schedulerConf, event)
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				scheduler.loadSchedulerConf()
				scheduler.cache.SetMetricsConf(scheduler.metricsConf)
			}
		case err, ok := <-errCh:
			if !ok {
				return
			}
			klog.Infof("watch %s error: %v", scheduler.schedulerConf, err)
		case <-stopCh:
			return
		}
	}
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

func UnmarshalSchedulerConf(confStr string) ([]framework.Action, []conf.Tier, []conf.Configuration, map[string]string, error) {
	var actions []framework.Action

	schedulerConf := &conf.SchedulerConfiguration{}
	if err := yaml.Unmarshal([]byte(confStr), schedulerConf); err != nil {
		return nil, nil, nil, nil, err
	}

	// Set default settings for each plugin if not set
	for i, tier := range schedulerConf.Tiers {
		// drf with hierarchy enabled
		hdrf := false
		// proportion enabled
		proportion := false
		for j := range tier.Plugins {
			if tier.Plugins[j].Name == "drf" &&
				tier.Plugins[j].EnabledHierarchy != nil &&
				*tier.Plugins[j].EnabledHierarchy {
				hdrf = true
			}
			if tier.Plugins[j].Name == "proportion" {
				proportion = true
			}
			plugins.ApplyPluginConfDefaults(&schedulerConf.Tiers[i].Plugins[j])
		}
		if hdrf && proportion {
			return nil, nil, nil, nil, fmt.Errorf("proportion and drf with hierarchy enabled conflicts")
		}
	}

	actionNames := strings.Split(schedulerConf.Actions, ",")
	for _, actionName := range actionNames {
		if action, found := framework.GetAction(strings.TrimSpace(actionName)); found {
			actions = append(actions, action)
		} else {
			return nil, nil, nil, nil, fmt.Errorf("failed to find Action %s, ignore it", actionName)
		}
	}

	return actions, schedulerConf.Tiers, schedulerConf.Configurations, schedulerConf.MetricsConfiguration, nil
}

func (scheduler *Scheduler) getSchedulerConf() (actions []string, plugins []string) {
	for _, action := range scheduler.actions {
		actions = append(actions, action.Name())
	}
	for _, tier := range scheduler.plugins {
		for _, plugin := range tier.Plugins {
			plugins = append(plugins, plugin.Name)
		}
	}
	return
}
