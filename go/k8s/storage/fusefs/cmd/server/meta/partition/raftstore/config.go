package raftstore

import (
	"fmt"

	"github.com/tiglabs/raft/proto"
)

// Constants for network port definition.
const (
	DefaultHeartbeatPort     = 5901
	DefaultReplicaPort       = 5902
	DefaultNumOfLogsToRetain = 20000
	DefaultTickInterval      = 300
	DefaultElectionTick      = 3
)

// Config defines the configuration properties for the raft store.
type Config struct {
	NodeID            uint64 // Identity of raft server instance.
	RaftPath          string // Path of raft logs
	IPAddr            string // IP address
	HeartbeatPort     int
	ReplicaPort       int
	NumOfLogsToRetain uint64 // number of logs to be kept after truncation. The default value is 20000.

	// TickInterval is the interval of timer which check heartbeat and election timeout.
	// The default value is 300,unit is millisecond.
	TickInterval int

	// RecvBufSize is the size of raft receive buffer channel.
	// The default value is 2048.
	RecvBufSize int

	// ElectionTick is the election timeout. If a follower does not receive any message
	// from the leader of current term during ElectionTick, it will become candidate and start an election.
	// ElectionTick must be greater than HeartbeatTick.
	// We suggest to use ElectionTick = 10 * HeartbeatTick to avoid unnecessary leader switching.
	// The default value is 1s.
	ElectionTick int
}

// PeerAddress defines the set of addresses that will be used by the peers.
type PeerAddress struct {
	proto.Peer
	Address       string
	HeartbeatPort int
	ReplicaPort   int
}

// PartitionConfig defines the configuration properties for the partitions.
type PartitionConfig struct {
	ID      uint64
	Applied uint64
	Leader  uint64
	Term    uint64
	Peers   []PeerAddress
	SM      PartitionFsm
	WalPath string
}

func (p PeerAddress) String() string {
	return fmt.Sprintf(`"nodeID":"%v","peerID":"%v","priority":"%v","type":"%v","heartbeatPort":"%v","ReplicaPort":"%v"`,
		p.ID, p.PeerID, p.Priority, p.Type.String(), p.HeartbeatPort, p.ReplicaPort)
}
