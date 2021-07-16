package raft

import "time"

// TransportConfig raft server transport config
type TransportConfig struct {
	// HeartbeatAddr is the Heartbeat port.
	// The default value is 3016.
	HeartbeatAddr string
	// ReplicateAddr is the Replation port.
	// The default value is 2015.
	ReplicateAddr string
	// 发送队列大小
	SendBufferSize int
	//复制并发数(node->node)
	MaxReplConcurrency int
	// MaxSnapConcurrency limits the max number of snapshot concurrency.
	// The default value is 10.
	MaxSnapConcurrency int
	// This parameter is required.
	Resolver SocketResolver
}

// Config contains the parameters to start a raft server.
// Default: Do not use lease mechanism.
// NOTE: NodeID and Resolver must be required.Other parameter has default value.
type Config struct {
	TransportConfig
	// NodeID is the identity of the local node. NodeID cannot be 0.
	// This parameter is required.
	NodeID uint64
	// TickInterval is the interval of timer which check heartbeat and election timeout.
	// The default value is 2s.
	TickInterval time.Duration
	// HeartbeatTick is the heartbeat interval. A leader sends heartbeat
	// message to maintain the leadership every heartbeat interval.
	// The default value is 2s.
	HeartbeatTick int
	// ElectionTick is the election timeout. If a follower does not receive any message
	// from the leader of current term during ElectionTick, it will become candidate and start an election.
	// ElectionTick must be greater than HeartbeatTick.
	// We suggest to use ElectionTick = 10 * HeartbeatTick to avoid unnecessary leader switching.
	// The default value is 10s.
	ElectionTick int
	// MaxSizePerMsg limits the max size of each append message.
	// The default value is 1M.
	MaxSizePerMsg uint64
	// MaxInflightMsgs limits the max number of in-flight append messages during optimistic replication phase.
	// The application transportation layer usually has its own sending buffer over TCP/UDP.
	// Setting MaxInflightMsgs to avoid overflowing that sending buffer.
	// The default value is 128.
	MaxInflightMsgs int
	// ReqBufferSize limits the max number of recive request chan buffer.
	// The default value is 1024.
	ReqBufferSize int
	// AppBufferSize limits the max number of apply chan buffer.
	// The default value is 2048.
	AppBufferSize int
	// RetainLogs controls how many logs we leave after truncate.
	// This is used so that we can quickly replay logs on a follower instead of being forced to send an entire snapshot.
	// The default value is 20000.
	RetainLogs uint64
	// LeaseCheck whether to use the lease mechanism.
	// The default value is false.
	LeaseCheck bool
	// ReadOnlyOption specifies how the read only request is processed.
	//
	// ReadOnlySafe guarantees the linearizability of the read only request by
	// communicating with the quorum. It is the default and suggested option.
	//
	// ReadOnlyLeaseBased ensures linearizability of the read only request by
	// relying on the leader lease. It can be affected by clock drift.
	// If the clock drift is unbounded, leader might keep the lease longer than it
	// should (clock can move backward/pause without any bound). ReadIndex is not safe
	// in that case.
	// LeaseCheck MUST be enabled if ReadOnlyOption is ReadOnlyLeaseBased.
	ReadOnlyOption ReadOnlyOption
	transport      Transport
}
