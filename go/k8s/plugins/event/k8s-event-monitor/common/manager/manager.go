package manager

import (
	"time"
)

type manager struct {
	Source          *kubernetes.EventSource
	ReceiverManager *receivers.ReceiverManager
	StopChan        chan struct{}
}

func NewManager(source common.EventSource, receiverManager *receivers.ReceiverManager) *manager {
	return &manager{
		StopChan:        make(chan struct{}),
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
