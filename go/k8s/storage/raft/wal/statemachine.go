package wal

type StateMachine interface {
	Apply([]byte) error
}

