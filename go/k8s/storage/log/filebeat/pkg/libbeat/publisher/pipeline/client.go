package pipeline

import (
	"sync"
)

type Client struct {
	pipeline   *Pipeline
	processors Processor
	producer   Producer
	mutex      sync.Mutex
	acker      ACKer
	waiter     *clientCloseWaiter

	eventFlags   EventFlags
	canDrop      bool
	reportEvents bool

	// Open state, signaling, and sync primitives for coordinating client Close.
	isOpen    Bool          // set to false during shutdown, such that no new events will be accepted anymore.
	closeOnce sync.Once     // closeOnce ensure that the client shutdown sequence is only executed once
	closeRef  CloseRef      // extern closeRef for sending a signal that the client should be closed.
	done      chan struct{} // the done channel will be closed if the closeReg gets closed, or Close is run.

	eventer ClientEventer
}

func (c *Client) Publish(e Event) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.publish(e)
}

func (c *Client) publish(e Event) {

	published := c.producer.Publish(pubEvent)

}
