package main

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/hashicorp/raft"
	"k8s.io/klog/v2"
)

type Fsm struct {
	mu   sync.Mutex
	Data database
}

func (f *Fsm) Apply(log *raft.Log) interface{} {
	klog.Infof(fmt.Sprintf("apply data: %s", log.Data))

	data := strings.Split(string(log.Data), ",")
	op := data[0]
	f.mu.Lock()
	if op == "set" {
		key := data[1]
		value := data[2]
		f.Data[key] = value
	}
	f.mu.Unlock()

	return nil
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	panic("implement me")
}

func (f *Fsm) Restore(closer io.ReadCloser) error {
	panic("implement me")
}

type database map[string]string
