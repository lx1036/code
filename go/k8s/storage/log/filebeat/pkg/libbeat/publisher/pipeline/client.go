package pipeline

import (
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/elastic/beats/libbeat/publisher/queue"
	"sync"
)

type Client struct {
	pipeline   *Pipeline
	processors beat.Processor
	producer   queue.Producer
	mutex      sync.Mutex
	acker      beat.ACKer
	waiter     *clientCloseWaiter

	eventFlags   publisher.EventFlags
	canDrop      bool
	reportEvents bool

	// Open state, signaling, and sync primitives for coordinating client Close.
	isOpen    atomic.Bool   // set to false during shutdown, such that no new events will be accepted anymore.
	closeOnce sync.Once     // closeOnce ensure that the client shutdown sequence is only executed once
	closeRef  beat.CloseRef // extern closeRef for sending a signal that the client should be closed.
	done      chan struct{} // the done channel will be closed if the closeReg gets closed, or Close is run.

	eventer beat.ClientEventer
}

func (c *Client) Publish(e beat.Event) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.publish(e)
}

func (c *Client) publish(e beat.Event) {

	published := c.producer.Publish(pubEvent)

}
