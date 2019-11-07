package workers

import (
	"fmt"
	"k8s-lx1036/wayne/backend/bus"
	"os"
	"sync/atomic"
)

type Worker interface {
	Run() error
	Stop() error
}

type BaseMessageWorker struct {
	Bus      *bus.Bus
	queue    string
	consumer string
	stopChan chan struct{}

	MessageWorker
}

func NewBaseMessageWorker(b *bus.Bus, queue string) *BaseMessageWorker {
	consumer := fmt.Sprintf("ctag-%s-%d", os.Args[0], atomic.AddUint64(&consumerSeq, 1))
	return &BaseMessageWorker{b, queue, consumer, make(chan struct{}), nil}
}
