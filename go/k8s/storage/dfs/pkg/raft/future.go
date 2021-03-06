package raft

type respErr struct {
	errCh chan error
}

// Future the future
type Future struct {
	respErr
	respCh chan interface{}
}
