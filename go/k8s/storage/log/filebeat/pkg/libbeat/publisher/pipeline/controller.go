package pipeline

import "github.com/elastic/beats/libbeat/publisher"

func newOutputController(beat beat.Info, queue queue.Queue) *outputController {
	c := &outputController{
		beat:      beat,
		queue:     queue,
		workQueue: make(chan publisher.Batch, 0),
	}

	ctx := &batchContext{}
	c.consumer = newEventConsumer(queue, ctx)
	c.retryer = newRetryer(c.workQueue, c.consumer)
	ctx.retryer = c.retryer

	c.consumer.sigContinue()

	return c
}
