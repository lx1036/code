package pipeline

func (pipeline *Pipeline) Connect() {
	client := &Client{
		pipeline:     p,
		closeRef:     cfg.CloseRef,
		done:         make(chan struct{}),
		isOpen:       MakeBool(true),
		eventer:      cfg.Events,
		processors:   processors,
		eventFlags:   eventFlags,
		canDrop:      canDrop,
		reportEvents: reportEvents,
	}

	client.acker = ackHandler
	client.waiter = waiter
	client.producer = pipeline.Producer(producerCfg)

	return client, nil
}
