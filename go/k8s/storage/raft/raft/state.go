// INFO: raft state 持久化

package raft

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type Role uint8

const (
	Follower Role = iota
	Candidate
	Leader
)

type RaftState struct {
	sync.RWMutex

	Role     Role `json:"role"`
	FilePath string

	// 这几个字段会被持久化
	PersistentState

	CommitIndex int64 `json:"commitIndex"`
}

type PersistentState struct {
	CurrentTerm int64 `json:"currentTerm"`
	VotedFor    int64 `json:"votedFor"`
}

func (state *RaftState) LoadState(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil || os.IsExist(err) {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		var persistentState PersistentState
		err = json.Unmarshal(content, &persistentState)
		if err != nil {
			return err
		}
		state.CurrentTerm = persistentState.CurrentTerm
		state.VotedFor = persistentState.VotedFor
	} else {
		state.CurrentTerm = 1
		state.VotedFor = 0
	}

	return nil
}

// SaveState 持久化到 raft/state.json
func (state *RaftState) SaveState() error {
	persistentState := &PersistentState{
		CurrentTerm: state.CurrentTerm,
		VotedFor:    state.VotedFor,
	}
	data, err := json.Marshal(persistentState)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(state.FilePath, data, 0644)

	return err
}

func NewRaftState(filePath string) (*RaftState, error) {
	state := &RaftState{
		Role:     Follower,
		FilePath: filePath,
	}
	if err := state.LoadState(filePath); err != nil {
		return nil, err
	}

	return state, nil
}
