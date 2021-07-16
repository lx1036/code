package raft

import (
	"net"
	"sync"
)

type heartbeatTransport struct {
	config     *TransportConfig
	raftServer *RaftServer
	listener   net.Listener
	mu         sync.RWMutex
	senders    map[uint64]*transportSender
	stopc      chan struct{}
}

func newHeartbeatTransport(raftServer *RaftServer, config *TransportConfig) (*heartbeatTransport, error) {
	var (
		listener net.Listener
		err      error
	)

	if listener, err = net.Listen("tcp", config.HeartbeatAddr); err != nil {
		return nil, err
	}
	t := &heartbeatTransport{
		config:     config,
		raftServer: raftServer,
		listener:   listener,
		senders:    make(map[uint64]*transportSender),
		stopc:      make(chan struct{}),
	}

	return t, nil
}
