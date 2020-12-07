package handlers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

//
type Manager struct {
	// 处理event的各个handler，包括log,360home,email等
	Handlers []events.Handler

	// 发送event给每一个handler的超时时间，超过了这个时间，这个event会被丢弃且不会重试
	SendEventsTimeout time.Duration
	// 比SendEventsTimeout大
	StopTimeout time.Duration
}

func (manager *Manager) ExportEvents(data *events.EventBatch) {
	wg := &sync.WaitGroup{}
	for _, handler := range manager.Handlers {
		wg.Add(1)
		go func(wg *sync.WaitGroup, handler events.Handler) {
			defer wg.Done()

			select {
			case handler.EventBatchChan <- data:

			case <-time.After(manager.SendEventsTimeout):
				log.WithFields(log.Fields{
					"msg": fmt.Sprintf("failed to send batch events to handler: %s", handler.Name),
				}).Warn("[ExportEvents]")
			}
		}(wg, handler)
	}

	wg.Wait()
}

func (manager *Manager) Stop() {
	for _, handler := range manager.Handlers {

		go func(handler events.Handler) {
			select {
			case handler.StopChan <- true:

			case <-time.After(manager.StopTimeout):

			}
		}(handler)
	}
}

func NewHandlerManager() {

}
