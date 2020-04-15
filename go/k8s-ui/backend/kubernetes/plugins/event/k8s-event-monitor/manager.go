package main

import (
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/sources/kubernetes"
	"time"
)

type manager struct {
	Source *kubernetes.EventSource
	Sink
}

func NewManager(srouce *kubernetes.EventSource) *manager {
	return &manager{}
}

func (mgr *manager) Start() {
	go func() {
		for {
			select {
			case <-time.After(nextTick):
				events := mgr.Source.GetEvents()
				mgr.Sink.Send(events)
			case <-mgr.stopChan:

			}
		}
	}()
}
