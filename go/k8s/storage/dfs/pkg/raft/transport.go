package raft

// Transport raft server transport
type Transport interface {
	Send(m *proto.Message)
	SendSnapshot(m *proto.Message, rs *snapshotStatus)
	Stop()
}
