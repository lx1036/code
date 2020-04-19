package common

import (
	kubeapi "k8s.io/api/core/v1"
	"time"
)

type Receiver interface {
	Send()
}

type Events struct {
	Timestamp time.Time
	Events    []*kubeapi.Event
}

type EventSource interface {
	GetEvents() Events
}
