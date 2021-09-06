package raft

import (
	"fmt"
	"k8s.io/klog/v2"
	"net"
	"runtime"
	"sync"

	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/util"
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

func (t *replicateTransport) start() {
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

func (t *replicateTransport) handleConn(conn *util.ConnTimeout) {
	go func() {
		defer conn.Close()
		loopCount := 0
		bufRd := util.NewBufferReader(conn, 16*KB)
		for {
			loopCount = loopCount + 1
			if loopCount > 16 {
				loopCount = 0
				runtime.Gosched()
			}

			select {
			case <-t.stopc:
				return
			default:
				if msg, err := receiveMessage(bufRd); err != nil {
					return
				} else {
					klog.Infof(fmt.Sprintf("Recive %v from (%v)", msg.ToString(), conn.RemoteAddr()))
					if msg.Type == proto.ReqMsgSnapShot {
						/*if err := t.handleSnapshot(msg, conn, bufRd); err != nil {
							return
						}*/
					} else {
						t.raftServer.receiveMessage(msg)
					}
				}
			}
		}
	}()
}
