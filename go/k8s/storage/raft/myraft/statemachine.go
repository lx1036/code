package myraft

type StateMachine interface {
	Apply([]byte) error
}
