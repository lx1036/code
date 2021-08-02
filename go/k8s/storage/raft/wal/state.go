package wal

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

func NewRaftState(filePath string) (*RaftState, error) {
	state := &RaftState{
		Role:     Follower,
		FilePath: "",
	}
	if err := state.LoadState(filePath); err != nil {
		return nil, err
	}

	return state, nil
}
