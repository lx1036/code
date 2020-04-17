package main

import (
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/common"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/receivers"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/sources/kubernetes"
	"time"
)

type manager struct {
	Source *kubernetes.EventSource
	ReceiverManager *receivers.ReceiverManager
	StopChan chan struct{}
}

func NewManager(source common.EventSource, receiverManager *receivers.ReceiverManager) *manager {
	return &manager{
		StopChan: make(chan struct{}),
		ReceiverManager: receiverManager,
	}
}

func (mgr *manager) Start() {
	go func() {
		for {
			nextTick := time.Second * 5

			select {
			case <-time.After(nextTick):
				events := mgr.Source.GetEvents()
				mgr.ReceiverManager.Send(events)
			case <-mgr.StopChan:
				return
			}
		}
	}()
}
