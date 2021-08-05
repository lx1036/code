// INFO: https://github.com/wereliang/raft

// INFO: raft博士作者论文中文版：https://github.com/maemual/raft-zh_cn/blob/master/raft-zh_cn.md

// INFO:
//  分布式一致性算法 Raft: https://zhuanlan.zhihu.com/p/383555591
//  一文搞懂Raft算法: https://www.cnblogs.com/xybaby/p/10124083.html

package wal

import (
	"fmt"
	"k8s.io/klog/v2"
	"sync"
)

type Raft struct {
	config *Config
	state  *RaftState

	processors map[Role]Processor
	processor  Processor

	wg  sync.WaitGroup
	mux sync.Mutex

	notifyC chan *raftEvent // processor -> raft
	applyC  chan *raftEvent // processor -> raft -> application state

	transport Transport

	stateMachine StateMachine

	raftLog RaftLog
}

func (raft *Raft) Start() error {
	go raft.raftLoop()
	return nil
}

func (raft *Raft) raftLoop() {
	raft.become(Follower)
	raft.wg.Add(3)
	defer raft.wg.Done()
	go func() {
		defer raft.wg.Done()
		raft.notifyLoop()
	}()
	go func() {
		defer raft.wg.Done()
		raft.applyLoop()
	}()
}

func (raft *Raft) notifyLoop() {
	for {
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

func (raft *Raft) become(role Role) {
}

func (raft *Raft) applyLoop() {
	for {
		select {
		case event := <-raft.applyC:
			klog.Infof(fmt.Sprintf("[applyLoop]event %+v", event))
			raft.apply()
		}
	}
}

func (raft *Raft) apply() {

}

// raft = config + transport + stateMachine
func NewRaft(config *Config, transport Transport, stateMachine StateMachine) (*Raft, error) {

	state, err := NewRaftState("raft/state.json")
	if err != nil {
		return nil, err
	}

	raftLog, err := NewStorage("raft")
	if err != nil {
		return nil, err
	}
	raft := &Raft{
		config:     config,
		state:      state,
		processors: nil,
		processor:  nil,
		notifyC:    make(chan *raftEvent, 256),
		applyC:     nil,

		transport: transport,

		stateMachine: nil,

		raftLog: raftLog,
	}

	commonProcessor := &FollowerProcessor{
		raftLog: raft.raftLog,
		state:   raft.state,
		notifyC: raft.notifyC,
	}
	raft.processors[Follower] = NewProcessor(Follower, commonProcessor)
	raft.processors[Candidate] = NewProcessor(Candidate, commonProcessor)
	raft.processors[Leader] = NewProcessor(Leader, commonProcessor)

	return raft, err
}
