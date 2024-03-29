package myraft

type eventName int8

// Event appointment
// Inn: Innernal event, by local, include candidate vote, leader append ...
// Ext: External event, by peer, include vote, append ...
// App: Application event, by application, include append log ...
// Notify: Notify event, include state machine, role switch
const (
	EventInnVoteRequest eventName = iota
	EventInnAppendLogRequest
	EventExtVoteRequest
	EventExtAppendLogRequest
	EventAppAppendLogRequest
	EvnetAppPeerChange
	EventNotifySwitchRole
	EventNotifyApply
)

type raftEvent struct {
	name eventName
	data interface{}
	resc chan interface{}
}

func newEvent(name eventName, data interface{}, resc chan interface{}) *raftEvent {
	return &raftEvent{
		name: name,
		data: data,
		resc: resc,
	}
}
