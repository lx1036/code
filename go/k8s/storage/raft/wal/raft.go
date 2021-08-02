


// INFO: https://github.com/wereliang/raft


package wal

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"
)

type Raft struct {
	state *RaftState
	
	processors map[Role]Processor
	processor Processor
	
	wg   sync.WaitGroup
	
	
	notifyC chan *raftEvent // processor -> raft
	applyC chan *raftEvent // processor -> raft -> application state
	
	stateMachine StateMachine
	
}

func (raft *Raft) Start() error {
	go raft.raftLoop()
	return nil
}

func (raft *Raft) raftLoop()  {
	
	raft.become(Follower)
	raft.wg.Add(3)
	defer raft.wg.Done()
	go func() { defer raft.wg.Done(); raft.notifyLoop() }()
	go func() { defer raft.wg.Done(); raft.applyLoop() }()
}

func (raft *Raft) notifyLoop()  {
	for  {
		select {
		case event := <-raft.notifyC:
			switch event.name {
			case EventNotifySwitchRole:
				role := event.data.(Role)
				raft.become(role)
			}
		}
	}
}

func (raft *Raft) become(role Role)  {
	
	
	
}

func (raft *Raft) applyLoop()  {
	for  {
		select {
		case event := <-raft.applyC:
			klog.Infof(fmt.Sprintf("[applyLoop]event %+v", event))
			raft.apply()
		}
	}
}

func (raft *Raft) apply()  {
	
}

func NewRaft() (*Raft, error) {




}
