package multiraft

import (
	"errors"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/raft/proto"
)

const (
	_ = iota
	// KB killobytes
	KB = 1 << (10 * iota)
	// MB megabytes
	MB = KB << 10
)

const (
	defaultTickInterval    = time.Second * 5
	defaultHeartbeatTick   = 10
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
	//ReadOnlyOption ReadOnlyOption

	// INFO: TCP transport
	transport Transport
}

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

type Node struct {
	mu sync.RWMutex

	config    *Config
	ticker    *time.Ticker
	heartChan chan *proto.Message
	stopc     chan struct{}
	rafts     map[uint64]*Raft
}

func StartNode(config *Config) (*Node, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	node := &Node{
		config:    config,
		ticker:    time.NewTicker(config.TickInterval),
		rafts:     make(map[uint64]*Raft),
		heartChan: make(chan *proto.Message, 512),
		stopc:     make(chan struct{}),
	}
	if transport, err := NewMultiTransport(node, &config.TransportConfig); err != nil {
		return nil, err
	} else {
		node.config.transport = transport
	}

	go node.run()

	return node, nil
}

func (node *Node) run() {
	ticks := 0
	for {
		select {
		case <-node.stopc:
			return

		case <-node.ticker.C:
			ticks++
			if ticks >= node.config.HeartbeatTick {
				ticks = 0
				node.sendHeartbeat()
			}

			for _, raft := range node.rafts {
				raft.tick()
			}

		case message := <-node.heartChan:
			switch message.Type {
			case proto.ReqMsgHeartBeat:
				node.handleHeartbeat(message)
			case proto.RespMsgHeartBeat:
				node.handleHeartbeatResp(message)
			}
		}
	}
}

// leader给peers发送心跳
func (node *Node) sendHeartbeat() {
	// key: sendto nodeId; value: range ids
	nodes := make(map[uint64]proto.HeartbeatContext)
	node.mu.RLock()
	for id, raft := range node.rafts {
		if !raft.isLeader() {
			continue
		}
		peers := raft.getPeers()
		for _, p := range peers {
			nodes[p] = append(nodes[p], id)
		}
	}
	node.mu.RUnlock()

	for to, ctx := range nodes {
		if to == node.config.NodeID {
			continue
		}

		msg := proto.NewMessage()
		msg.Type = proto.ReqMsgHeartBeat
		msg.From = node.config.NodeID
		msg.To = to
		msg.Context = proto.EncodeHBConext(ctx)
		node.config.transport.Send(msg)
	}
}

// INFO: 处理心跳请求
func (node *Node) handleHeartbeat(message *proto.Message) {
	node.mu.RLock() // TODO: 这里为何需要加锁
	ctx := proto.DecodeHBContext(message.Context)
	var respCtx proto.HeartbeatContext
	for _, id := range ctx { // id 是每一个字节的 uint64 值
		if raft, ok := node.rafts[id]; ok {
			raft.receiveMessage(message)
			respCtx = append(respCtx, id)
		}
	}
	node.mu.RUnlock()

	msg := proto.NewMessage()
	msg.Type = proto.RespMsgHeartBeat
	msg.From = node.config.NodeID
	msg.To = message.From
	msg.Context = proto.EncodeHBConext(respCtx)
	node.config.transport.Send(msg)
}

func (node *Node) handleHeartbeatResp(message *proto.Message) {
	node.mu.RLock()
	defer node.mu.RUnlock()

	ctx := proto.DecodeHBContext(message.Context)
	for _, id := range ctx {
		if raft, ok := node.rafts[id]; ok {
			raft.receiveMessage(message)
		}
	}
}

func (node *Node) receiveMessage(message *proto.Message) {
	if message.Type == proto.ReqMsgHeartBeat || message.Type == proto.RespMsgHeartBeat {
		node.heartChan <- message
		return
	}

	node.mu.RLock()
	defer node.mu.RUnlock()
	raft, ok := node.rafts[message.ID]
	if ok {
		raft.receiveMessage(message)
	}
}
