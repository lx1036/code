package pipeline

import (
	"github.com/elastic/beats/libbeat/outputs"
	"github.com/elastic/beats/libbeat/publisher/processing"
	"k8s-lx1036/k8s/storage/log/filebeat/pkg/libbeat/outputs/console"
	"k8s-lx1036/k8s/storage/log/filebeat/pkg/libbeat/publisher/queue"
	"k8s-lx1036/k8s/storage/log/filebeat/pkg/libbeat/publisher/queue/memoryqueue"
	"sync"
	"time"
)

type Pipeline struct {
	beatInfo beat.Info

	monitors Monitors

	queue  queue.Queue
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

	processors processing.Supporter
}

// NewPipeline INFO: pipeline = queue + output
func NewPipeline(
	beatInfo beat.Info,
	config Config,
	processors processing.Supporter,
	makeOutput func(outputs.Observer) (string, outputs.Group, error),
) (*Pipeline, error) {
	settings := Settings{
		WaitClose:     0,
		WaitCloseMode: NoWaitOnClose,
		Processors:    processors,
	}

	output, err := console.NewConsoleOutput()
	queue := memoryqueue.NewMemoryQueue()

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
