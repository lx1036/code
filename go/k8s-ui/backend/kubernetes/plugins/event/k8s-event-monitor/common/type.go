package common

import (
	"time"
	kubeapi "k8s.io/api/core/v1"
	)

type Receiver interface {
	Send()
}

type Events struct {
	Timestamp time.Time
	Events []*kubeapi.Event
}
