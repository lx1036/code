package memoryqueue

import "time"

// bufferingEventLoop implements the broker main event loop.
// Events in the buffer are forwarded to consumers only if the buffer is full or on flush timeout.
type bufferingEventLoop struct {
	broker *Broker

	buf        *batchBuffer
	flushList  flushList
	eventCount int

	minEvents    int
	maxEvents    int
	flushTimeout time.Duration

	// active broker API channels
	events    chan pushRequest
	get       chan getRequest
	pubCancel chan producerCancelRequest

	// ack handling
	acks        chan int      // ackloop -> eventloop : total number of events ACKed by outputs
	schedACKS   chan chanList // eventloop -> ackloop : active list of batches to be acked
	pendingACKs chanList      // ordered list of active batches to be send to the ackloop
	ackSeq      uint          // ack batch sequence number to validate ordering

	// buffer flush timer state
	timer *time.Timer
	idleC <-chan time.Time
}

func newBufferingEventLoop(b *Broker, size int, minEvents int, flushTimeout time.Duration) *bufferingEventLoop {
	l := &bufferingEventLoop{
		broker:       b,
		maxEvents:    size,
		minEvents:    minEvents,
		flushTimeout: flushTimeout,

		events:    b.events,
		get:       nil,
		pubCancel: b.pubCancel,
		acks:      b.acks,
	}
	l.buf = newBatchBuffer(l.minEvents)

	l.timer = time.NewTimer(flushTimeout)
	if !l.timer.Stop() {
		<-l.timer.C
	}

	return l
}

func (l *bufferingEventLoop) run() {
	var (
		broker = l.broker
	)

	for {
		select {
		case <-broker.done:
			return

		case req := <-l.events: // producer pushing new event
			l.handleInsert(&req)

		//case req := <-l.pubCancel: // producer cancelling active events
		//	l.handleCancel(&req)

		case req := <-l.get: // consumer asking for next batch
			l.handleConsumer(&req)

		case l.schedACKS <- l.pendingACKs:
			l.schedACKS = nil
			l.pendingACKs = chanList{}

		case count := <-l.acks:
			l.handleACK(count)

		case <-l.idleC:
			l.idleC = nil
			l.timer.Stop()
			if l.buf.length() > 0 {
				l.flushBuffer()
			}
		}
	}
}
