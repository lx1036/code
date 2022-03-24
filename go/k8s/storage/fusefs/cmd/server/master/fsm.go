package master

import (
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type Fsm struct {
	sync.Mutex
}

func (f *Fsm) Apply(log *raft.Log) interface{} {
	panic("implement me")
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	panic("implement me")
}

func (f *Fsm) Restore(snapshot io.ReadCloser) error {
	panic("implement me")
}
