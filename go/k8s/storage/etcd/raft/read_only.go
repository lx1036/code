package raft

// ReadState INFO: 线性一致性读, 返回给 read-only node 该 ReadState
type ReadState struct {
	Index      uint64
	RequestCtx []byte
}
