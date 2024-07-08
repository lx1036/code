package endpoint

import (
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
)

type StatusCode int

const (
	OK       StatusCode = 0
	Warning  StatusCode = -1
	Failure  StatusCode = -2
	Disabled StatusCode = -3
)

// StatusType represents the type for the given status, higher the value, higher
// the priority.
type StatusType int

const (
	BPF    StatusType = 200
	Policy StatusType = 100
	Other  StatusType = 0
)

type Status struct {
	Code  StatusCode `json:"code"`
	Msg   string     `json:"msg"`
	Type  StatusType `json:"status-type"`
	State string     `json:"state"`
}

// statusLogMsg represents a log message.
type statusLogMsg struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// statusLog represents a slice of statusLogMsg.
type statusLog []*statusLogMsg

// componentStatus represents a map of a single statusLogMsg by StatusType.
type componentStatus map[StatusType]*statusLogMsg

// EndpointStatus represents the endpoint status.
type EndpointStatus struct {
	// CurrentStatuses is the last status of a given priority.
	CurrentStatuses componentStatus `json:"current-status,omitempty"`
	// Contains the last maxLogs messages for this endpoint.
	Log statusLog `json:"log,omitempty"`
	// Index is the index in the statusLog, is used to keep track the next
	// available position to write a new log message.
	Index int `json:"index"`
	// indexMU is the Mutex for the CurrentStatus and Log RW operations.
	indexMU lock.RWMutex
}

func NewEndpointStatus() *EndpointStatus {
	return &EndpointStatus{
		CurrentStatuses: componentStatus{},
		Log:             statusLog{},
	}
}
