package wal

type Processor interface {
	// HandleEvent external event and application event
	HandleEvent(evt *raftEvent) (interface{}, error)
	Init() error
	Loop()
	Stop() bool
}

type FollowerProcessor struct {
}

func (processor *FollowerProcessor) HandleEvent(evt *raftEvent) (interface{}, error) {
	panic("implement me")
}

func (processor *FollowerProcessor) Init() error {
	panic("implement me")
}

func (processor *FollowerProcessor) Loop() {
	panic("implement me")
}

func (processor *FollowerProcessor) Stop() bool {
	panic("implement me")
}

type CandidateProcessor struct {
	FollowerProcessor

	ConnectionManager *PeerConnectionManager
}

type LeaderProcessor struct {
	CandidateProcessor

	syncC chan interface{}
}

func NewProcessor(role Role, processor Processor) Processor {

}
