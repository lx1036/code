package multiraft

import (
	"fmt"
	"net"
	"runtime"
	"sync"

	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/util"

	"k8s.io/klog/v2"
)

type SocketType byte

const (
	HeartBeat SocketType = 0
	Replicate SocketType = 1
)

// Transport raft server transport
type Transport interface {
	Send(m *proto.Message)
	//SendSnapshot(m *proto.Message, rs *snapshotStatus)
	Stop()
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
	//Resolver SocketResolver
}

type MultiTransport struct {
	heartbeat *heartbeatTransport
	replicate *replicateTransport
}

func NewMultiTransport(node *Node, config *TransportConfig) (Transport, error) {
	mt := new(MultiTransport)

	if ht, err := newHeartbeatTransport(node, config); err != nil {
		return nil, err
	} else {
		mt.heartbeat = ht
	}
	if rt, err := newReplicateTransport(node, config); err != nil {
		return nil, err
	} else {
		mt.replicate = rt
	}

	mt.heartbeat.start()
	mt.replicate.start()

	return mt, nil
}

func (t *MultiTransport) Send(m *proto.Message) {
	// if m.IsElectionMsg() {
	if m.IsHeartbeatMsg() {
		t.heartbeat.send(m)
	} else {
		t.replicate.send(m)
	}
}

func (t *MultiTransport) Stop() {
	t.heartbeat.stop()
	t.replicate.stop()
}

type heartbeatTransport struct {
	mu sync.RWMutex

	config   *TransportConfig
	node     *Node
	listener net.Listener
	senders  map[uint64]*transportSender
	stopc    chan struct{}
}

func newHeartbeatTransport(node *Node, config *TransportConfig) (*heartbeatTransport, error) {
	var (
		listener net.Listener
		err      error
	)

	if listener, err = net.Listen("tcp", config.HeartbeatAddr); err != nil {
		return nil, err
	}
	t := &heartbeatTransport{
		config:   config,
		node:     node,
		listener: listener,
		senders:  make(map[uint64]*transportSender),
		stopc:    make(chan struct{}),
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
				if msg, err := convertMessage(bufRd); err != nil {
					return
				} else {
					klog.Infof(fmt.Sprintf("Receive %v from (%v)", msg.ToString(), conn.RemoteAddr()))
					t.node.receiveMessage(msg)
				}
			}
		}
	}()
}

func (t *heartbeatTransport) getSender(nodeId uint64) *transportSender {
	t.mu.RLock()
	sender, ok := t.senders[nodeId]
	t.mu.RUnlock()
	if ok {
		return sender
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if sender, ok = t.senders[nodeId]; !ok {
		sender = newTransportSender(nodeId, 1, 64, HeartBeat, t.config.Resolver)
		t.senders[nodeId] = sender
	}
	return sender
}

func (t *heartbeatTransport) send(msg *proto.Message) {
	s := t.getSender(msg.To)
	s.send(msg)
}

func (t *heartbeatTransport) stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	select {
	case <-t.stopc:
		return
	default:
		close(t.stopc)
		t.listener.Close()
		for _, s := range t.senders {
			s.stop()
		}
	}
}

type replicateTransport struct {
	config      *TransportConfig
	node        *Node
	listener    net.Listener
	curSnapshot int32
	mu          sync.RWMutex
	senders     map[uint64]*transportSender
	stopc       chan struct{}
}

func newReplicateTransport(node *Node, config *TransportConfig) (*replicateTransport, error) {
	var (
		listener net.Listener
		err      error
	)

	if listener, err = net.Listen("tcp", config.ReplicateAddr); err != nil {
		return nil, err
	}
	t := &replicateTransport{
		config:   config,
		node:     node,
		listener: listener,
		senders:  make(map[uint64]*transportSender),
		stopc:    make(chan struct{}),
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
				if msg, err := convertMessage(bufRd); err != nil {
					return
				} else {
					klog.Infof(fmt.Sprintf("Recive %v from (%v)", msg.ToString(), conn.RemoteAddr()))
					if msg.Type == proto.ReqMsgSnapShot {
						/*if err := t.handleSnapshot(msg, conn, bufRd); err != nil {
							return
						}*/
					} else {
						t.node.receiveMessage(msg)
					}
				}
			}
		}
	}()
}

func (t *replicateTransport) send(m *proto.Message) {
	s := t.getSender(m.To)
	s.send(m)
}

func (t *replicateTransport) getSender(nodeId uint64) *transportSender {
	t.mu.RLock()
	sender, ok := t.senders[nodeId]
	t.mu.RUnlock()
	if ok {
		return sender
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if sender, ok = t.senders[nodeId]; !ok {
		sender = newTransportSender(nodeId, uint64(t.config.MaxReplConcurrency), t.config.SendBufferSize, Replicate, t.config.Resolver)
		t.senders[nodeId] = sender
	}
	return sender
}

func (t *replicateTransport) stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	select {
	case <-t.stopc:
		return
	default:
		close(t.stopc)
		t.listener.Close()
		for _, s := range t.senders {
			s.stop()
		}
	}
}

type transportSender struct {
	nodeID      uint64
	concurrency uint64
	senderType  SocketType
	resolver    SocketResolver
	inputc      []chan *proto.Message
	send        func(msg *proto.Message)
	mu          sync.Mutex
	stopc       chan struct{}
}

func newTransportSender(nodeID, concurrency uint64, buffSize int, senderType SocketType, resolver SocketResolver) *transportSender {
	sender := &transportSender{
		nodeID:      nodeID,
		concurrency: concurrency,
		senderType:  senderType,
		resolver:    resolver,
		inputc:      make([]chan *proto.Message, concurrency),
		stopc:       make(chan struct{}),
	}
	for i := uint64(0); i < concurrency; i++ {
		sender.inputc[i] = make(chan *proto.Message, buffSize)
		sender.loopSend(sender.inputc[i])
	}

	if (concurrency & (concurrency - 1)) == 0 {
		sender.send = func(msg *proto.Message) {
			idx := 0
			if concurrency > 1 {
				idx = int(msg.ID&concurrency - 1)
			}
			sender.inputc[idx] <- msg
		}
	} else {
		sender.send = func(msg *proto.Message) {
			idx := 0
			if concurrency > 1 {
				idx = int(msg.ID % concurrency)
			}
			sender.inputc[idx] <- msg
		}
	}

	return sender
}

func (s *transportSender) loopSend(message chan *proto.Message) {
	go func() {
		for {
			select {
			case <-s.stopc:
				return

			default:

			}
		}
	}()
}

func (s *transportSender) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.stopc:
		return
	default:
		close(s.stopc)
	}
}

func convertMessage(r *util.BufferReader) (msg *proto.Message, err error) {
	msg = proto.NewMessage()
	if err = msg.Decode(r); err != nil {
		proto.ReturnMessage(msg)
	}
	return
}
