package events

import "time"

type EventType string

const (
	Create EventType = "create"
	Update EventType = "update"
	Delete EventType = "delete"
)

type Event struct {
	Name         string
	EventType    EventType
	Namespace    string
	ResourceType string
}

type Handler struct {
	EventHandler   EventHandler
	EventBatchChan chan *EventBatch
	StopChan       chan bool
}

type EventBatch struct {
	Events    []*Event
	Timestamp time.Time
}

type EventHandler interface {
}
