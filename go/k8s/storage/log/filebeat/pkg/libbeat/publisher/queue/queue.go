package queue

import (
	"github.com/elastic/beats/libbeat/publisher"
	"io"
)

// Queue INFO: 非常重要的缓存对象，主要是 buffer events from producer。
type Queue interface {
	io.Closer

	BufferConfig() BufferConfig

	Producer(cfg ProducerConfig) Producer
	Consumer() Consumer
}

type Producer interface {
	Publish(event publisher.Event) bool
}
