package batchscheduler

import (
	"fmt"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/batchscheduler/schedulerinterface"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/batchscheduler/volcano"
	"k8s.io/client-go/rest"
	"sync"
)

type schedulerInitializeFunc func(config *rest.Config) (schedulerinterface.BatchScheduler, error)

var schedulerContainers = map[string]schedulerInitializeFunc{
	volcano.GetPluginName(): volcano.New,
}

type SchedulerManager struct {
	sync.Mutex
	config  *rest.Config
	plugins map[string]schedulerinterface.BatchScheduler
}

func (batch *SchedulerManager) GetScheduler(schedulerName string) (schedulerinterface.BatchScheduler, error) {
	initFunc, registered := schedulerContainers[schedulerName]
	if !registered {
		return nil, fmt.Errorf("unregistered scheduler plugin %s", schedulerName)
	}

	batch.Lock()
	defer batch.Unlock()

	if plugin, existed := batch.plugins[schedulerName]; existed && plugin != nil {
		return plugin, nil
	} else if existed && plugin == nil {
		return nil, fmt.Errorf("failed to get scheduler plugin %s, previous initialization has failed", schedulerName)
	} else {
		if plugin, err := initFunc(batch.config); err != nil {
			batch.plugins[schedulerName] = nil
			return nil, err
		} else {
			batch.plugins[schedulerName] = plugin
			return plugin, nil
		}
	}
}

func NewSchedulerManager(config *rest.Config) *SchedulerManager {
	manager := &SchedulerManager{
		config:  config,
		plugins: make(map[string]schedulerinterface.BatchScheduler),
	}

	return manager
}
