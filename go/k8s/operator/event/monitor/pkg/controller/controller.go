package controller

import (
	"k8s-lx1036/k8s/operator/event/monitor/pkg/handlers"
	"k8s-lx1036/k8s/operator/event/monitor/pkg/sources"
	"time"
)

type Controller struct {
	SourceManager  sources.Manager
	HandlerManager handlers.Manager

	Frequency time.Duration

	StopChan chan struct{}
}

func (controller *Controller) Start() {
	go controller.Run()
}

func (controller *Controller) Run() {
	for {
		select {
		case <-time.Tick(controller.Frequency):
			controller.run()
		case <-controller.stopChan:
			rm.sink.Stop()
			return
		}
	}
}

func (controller *Controller) run() {
	events := controller.SourceManager.GetNewEvents()
	controller.HandlerManager.ExportEvents(events)
}

func New(sourceManager sources.Manager, handlerManager handlers.Manager, frequency time.Duration) *Controller {
	return &Controller{
		SourceManager:  sourceManager,
		HandlerManager: handlerManager,
		Frequency:      frequency,
	}
}
