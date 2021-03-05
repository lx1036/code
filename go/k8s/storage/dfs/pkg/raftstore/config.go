package raftstore

import "github.com/tiglabs/raft/proto"

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
