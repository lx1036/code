package memoryqueue

import (
	"github.com/elastic/beats/libbeat/publisher/queue"
	"sync"
)

type Broker struct {
	done chan struct{}

	bufSize int

	// api channels
	events    chan pushRequest
	requests  chan getRequest
	pubCancel chan producerCancelRequest

	// internal channels
	acks          chan int
	scheduledACKs chan chanList

	ackListener queue.ACKListener

	// wait group for worker shutdown
	wg          sync.WaitGroup
	waitOnClose bool
}

// NewMemoryQueue queue INFO: queue 里开了两个 loop: EventLoop 和 AckLoop
func NewMemoryQueue() *Broker {
	broker := &Broker{
		done: make(chan struct{}),

		// broker API channels
		events:    make(chan pushRequest, chanSize),
		requests:  make(chan getRequest),
		pubCancel: make(chan producerCancelRequest, 5),

		// internal broker and ACK handler channels
		acks:          make(chan int),
		scheduledACKs: make(chan chanList),

		waitOnClose: settings.WaitOnClose,

		ackListener: settings.ACKListener,
	}

	eventLoop := newBufferingEventLoop(broker, sz, minEvents, flushTimeout)
	broker.bufSize = sz
	ack := newACKLoop(broker, eventLoop.processACK)

	// INFO: 开启 eventloop 和 ack loop，从 channel 里读数据
	//  ack 是consumer从缓存队列queue中拿到数据，发送给output之后，需要output一个ack应答确认，这样保证日志被至少消费一次
	broker.wg.Add(2)
	go func() {
		defer broker.wg.Done()
		eventLoop.run()
	}()
	go func() {
		defer broker.wg.Done()
		ack.run()
	}()

	return broker
}

func (broker *Broker) Producer(cfg queue.ProducerConfig) queue.Producer {
	return newProducer(broker, cfg.ACK, cfg.OnDrop, cfg.DropOnCancel)
}

func (broker *Broker) Consumer() queue.Consumer {
	return newConsumer(broker)
}
