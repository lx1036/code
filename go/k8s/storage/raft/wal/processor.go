package wal

type Processor interface {
	// HandleEvent external event and application event
	HandleEvent(evt *raftEvent) (interface{}, error)
	Init() error
	Loop()
	Stop() bool
}

type CommandProcessor struct {
}

func (processor *CommandProcessor) Init() error {

}
