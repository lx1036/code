package controller

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/plugins/event/monitor/pkg/handlers"
	"k8s-lx1036/k8s/plugins/event/monitor/pkg/sources"
	"k8s-lx1036/k8s/plugins/event/monitor/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
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
