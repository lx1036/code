package raft

import (
	"fmt"
	"k8s.io/klog/v2"
	"net"
	"sync"

	"k8s-lx1036/k8s/storage/raft/util"
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

func (t *heartbeatTransport) start() {
	go func() {
		for {
			select {
			case <-t.stopc:
				return
			default:
				conn, err := t.listener.Accept()
				if err != nil {
					continue
				}
				t.handleConn(util.NewConnTimeout(conn))
			}
		}
	}()
}

func (t *heartbeatTransport) handleConn(conn *util.ConnTimeout) {
	go func() {
		defer conn.Close()
		bufRd := util.NewBufferReader(conn, 16*KB)
		for {
			select {
			case <-t.stopc:
				return
			default:
				if msg, err := receiveMessage(bufRd); err != nil {
					return
				} else {
					klog.Infof(fmt.Sprintf("Recive %v from (%v)", msg.ToString(), conn.RemoteAddr()))
					t.raftServer.receiveMessage(msg)
				}
			}
		}
	}()
}
