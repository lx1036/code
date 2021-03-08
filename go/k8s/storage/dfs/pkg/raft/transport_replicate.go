package raft

import (
	"net"
	"sync"
)

type replicateTransport struct {
	config      *TransportConfig
	raftServer  *RaftServer
	listener    net.Listener
	curSnapshot int32
	mu          sync.RWMutex
	senders     map[uint64]*transportSender
	stopc       chan struct{}
}

func newReplicateTransport(raftServer *RaftServer, config *TransportConfig) (*replicateTransport, error) {
	var (
		listener net.Listener
		err      error
	)

	if listener, err = net.Listen("tcp", config.ReplicateAddr); err != nil {
		return nil, err
	}
	t := &replicateTransport{
		config:     config,
		raftServer: raftServer,
		listener:   listener,
		senders:    make(map[uint64]*transportSender),
		stopc:      make(chan struct{}),
	}
	return t, nil
}
