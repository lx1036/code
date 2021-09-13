package raft

type StateMachine interface {
	Apply([]byte) error
}
