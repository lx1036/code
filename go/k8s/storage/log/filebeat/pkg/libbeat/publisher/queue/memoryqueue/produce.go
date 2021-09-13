package memoryqueue

import ()

type OpenState struct {
	done   chan struct{}
	events chan pushRequest
}

func (st *OpenState) publish(req pushRequest) bool {
	select {
	case st.events <- req:
		return true
	case <-st.done:
		st.events = nil
		return false
	}
}

type forgetfulProducer struct {
	broker    *Broker
	openState OpenState
}

func newProducer(b *Broker, cb ackHandler, dropCB func(Event), dropOnCancel bool) Producer {
	openState := OpenState{
		done:   make(chan struct{}),
		events: b.events,
	}

	return &forgetfulProducer{broker: b, openState: openState}
}

func (p *forgetfulProducer) Publish(event Event) bool {
	return p.openState.publish(p.makeRequest(event))
}
