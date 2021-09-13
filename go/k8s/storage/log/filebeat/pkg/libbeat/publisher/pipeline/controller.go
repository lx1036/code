package pipeline

func newOutputController(beat Info, queue Queue) *outputController {
	c := &outputController{
		beat:      beat,
		queue:     queue,
		workQueue: make(chan Batch, 0),
	}

	ctx := &batchContext{}
	c.consumer = newEventConsumer(queue, ctx)
	c.retryer = newRetryer(c.workQueue, c.consumer)
	ctx.retryer = c.retryer

	c.consumer.sigContinue()

	return c
}
