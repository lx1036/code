package config

import (
	"errors"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/sunfs/pkg/raft/proto"
	"k8s-lx1036/k8s/storage/sunfs/pkg/raft/storage"
)

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

// RaftConfig contains the parameters to create a raft.
type RaftConfig struct {
	ID           uint64
	Term         uint64
	Leader       uint64
	Applied      uint64
	Peers        []proto.Peer
	Storage      storage.Storage
	StateMachine StateMachine
}

// validate returns an error if any required elements of the ReplConfig are missing or invalid.
func (c *RaftConfig) validate() error {
	if c.ID == 0 {
		return errors.New("ID is required")
	}
	if len(c.Peers) == 0 {
		return errors.New("Peers is required")
	}
	if c.Storage == nil {
		return errors.New("Storage is required")
	}
	if c.StateMachine == nil {
		return errors.New("StateMachine is required")
	}

	return nil
}

const (
	_ = iota
	// KB killobytes
	KB = 1 << (10 * iota)
	// MB megabytes
	MB
)

const (
	defaultTickInterval    = time.Millisecond * 2000
	defaultHeartbeatTick   = 1
	defaultElectionTick    = 5
	defaultInflightMsgs    = 128
	defaultSizeReqBuffer   = 2048
	defaultSizeAppBuffer   = 2048
	defaultRetainLogs      = 20000
	defaultSizeSendBuffer  = 10240
	defaultReplConcurrency = 5
	defaultSnapConcurrency = 10
	defaultSizePerMsg      = MB
	defaultHeartbeatAddr   = ":3016"
	defaultReplicateAddr   = ":2015"
)

// validate returns an error if any required elements of the Config are missing or invalid.
func (c *Config) validate() error {
	if c.NodeID == 0 {
		return errors.New("NodeID is required")
	}
	if c.TransportConfig.Resolver == nil {
		return errors.New("Resolver is required")
	}
	if c.MaxSizePerMsg > 4*MB {
		return errors.New("MaxSizePerMsg it too high")
	}
	if c.MaxInflightMsgs > 1024 {
		return errors.New("MaxInflightMsgs is too high")
	}
	if c.MaxSnapConcurrency > 256 {
		return errors.New("MaxSnapConcurrency is too high")
	}
	if c.MaxReplConcurrency > 256 {
		return errors.New("MaxReplConcurrency is too high")
	}
	if c.ReadOnlyOption == ReadOnlyLeaseBased && !c.LeaseCheck {
		return errors.New("LeaseCheck MUST be enabled when use ReadOnlyLeaseBased")
	}

	if strings.TrimSpace(c.TransportConfig.HeartbeatAddr) == "" {
		c.TransportConfig.HeartbeatAddr = defaultHeartbeatAddr
	}
	if strings.TrimSpace(c.TransportConfig.ReplicateAddr) == "" {
		c.TransportConfig.ReplicateAddr = defaultReplicateAddr
	}
	if c.TickInterval < 5*time.Millisecond {
		c.TickInterval = defaultTickInterval
	}
	if c.HeartbeatTick <= 0 {
		c.HeartbeatTick = defaultHeartbeatTick
	}
	if c.ElectionTick <= 0 {
		c.ElectionTick = defaultElectionTick
	}
	if c.MaxSizePerMsg <= 0 {
		c.MaxSizePerMsg = defaultSizePerMsg
	}
	if c.MaxInflightMsgs <= 0 {
		c.MaxInflightMsgs = defaultInflightMsgs
	}
	if c.ReqBufferSize <= 0 {
		c.ReqBufferSize = defaultSizeReqBuffer
	}
	if c.AppBufferSize <= 0 {
		c.AppBufferSize = defaultSizeAppBuffer
	}
	if c.MaxSnapConcurrency <= 0 {
		c.MaxSnapConcurrency = defaultSnapConcurrency
	}
	if c.MaxReplConcurrency <= 0 {
		c.MaxReplConcurrency = defaultReplConcurrency
	}
	if c.SendBufferSize <= 0 {
		c.SendBufferSize = defaultSizeSendBuffer
	}
	return nil
}

// DefaultConfig returns a Config with usable defaults.
func DefaultConfig() *Config {
	conf := &Config{
		TickInterval:    defaultTickInterval,
		HeartbeatTick:   defaultHeartbeatTick,
		ElectionTick:    defaultElectionTick,
		MaxSizePerMsg:   defaultSizePerMsg,
		MaxInflightMsgs: defaultInflightMsgs,
		ReqBufferSize:   defaultSizeReqBuffer,
		AppBufferSize:   defaultSizeAppBuffer,
		RetainLogs:      defaultRetainLogs,
		LeaseCheck:      false,
	}
	conf.HeartbeatAddr = defaultHeartbeatAddr
	conf.ReplicateAddr = defaultReplicateAddr
	conf.SendBufferSize = defaultSizeSendBuffer
	conf.MaxReplConcurrency = defaultReplConcurrency
	conf.MaxSnapConcurrency = defaultSnapConcurrency

	return conf
}
