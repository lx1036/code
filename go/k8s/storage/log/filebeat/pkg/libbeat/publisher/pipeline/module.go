package pipeline

import (
	"k8s-lx1036/k8s/storage/log/filebeat/pkg/libbeat/outputs/console"
	"k8s-lx1036/k8s/storage/log/filebeat/pkg/libbeat/publisher/queue"
	"k8s-lx1036/k8s/storage/log/filebeat/pkg/libbeat/publisher/queue/memoryqueue"
	"sync"
	"time"
)

type Pipeline struct {
	beatInfo Info

	monitors Monitors

	queue  Queue
	output *outputController

	eventer pipelineEventer

	// wait close support
	waitCloseMode    WaitCloseMode
	waitCloseTimeout time.Duration
	waitCloser       *waitCloser

	// pipeline ack
	eventSema *sema

	// closeRef signal propagation support
	guardStartSigPropagation sync.Once
	sigNewClient             chan *client

	processors Supporter
}

// NewPipeline INFO: pipeline = queue + output
func NewPipeline(
	beatInfo Info,
	config Config,
	processors Supporter,
	makeOutput func(Observer) (string, Group, error),
) (*Pipeline, error) {
	settings := Settings{
		WaitClose:     0,
		WaitCloseMode: NoWaitOnClose,
		Processors:    processors,
	}

	output, err := console.NewConsoleOutput()
	queue := memoryNewMemoryQueue()

	pipeline := &Pipeline{
		beatInfo:         beat,
		waitCloseMode:    settings.WaitCloseMode,
		waitCloseTimeout: settings.WaitClose,
		processors:       settings.Processors,
	}

	pipeline.queue = queue
	pipeline.output = newOutputController(beat, pipeline.queue)
	pipeline.output.Set(output)

	return pipeline, nil
}
