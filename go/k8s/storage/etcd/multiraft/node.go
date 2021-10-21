package multiraft

import (
	"errors"
	"fmt"
	"k8s.io/klog/v2"
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
	defaultTickInterval    = time.Second * 30
	defaultHeartbeatTick   = 2
	defaultElectionTick    = 5
	defaultInflightMsgs    = 128
	defaultSizeReqBuffer   = 2048
	defaultSizeAppBuffer   = 2048
	defaultRetainLogs      = 20000
	defaultSizeSendBuffer  = 10240
	defaultReplConcurrency = 5
	defaultSnapConcurrency = 10
	defaultSizePerMsg      = MB
	defaultHeartbeatAddr   = "127.0.0.1:2020"
	defaultReplicateAddr   = "127.0.0.1:2021"
)

type NodeConfig struct {
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

func DefaultConfig() *NodeConfig {
	conf := &NodeConfig{
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

func (c *NodeConfig) validate() error {
	if c.NodeID == 0 {
		return errors.New("NodeID is required")
	}
	/*if c.TransportConfig.Resolver == nil {
		return errors.New("Resolver is required")
	}*/
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
	/*if c.ReadOnlyOption == ReadOnlyLeaseBased && !c.LeaseCheck {
		return errors.New("LeaseCheck MUST be enabled when use ReadOnlyLeaseBased")
	}*/

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

	config    *NodeConfig
	ticker    *time.Ticker
	heartChan chan *proto.Message
	stopc     chan struct{}
	rafts     map[uint64]*Raft
}

func NewNode(config *NodeConfig) (*Node, error) {
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

// INFO: Node 主要就是处理 node 级别的心跳消息
func (node *Node) run() {
	ticks := 0
	for {
		select {
		case <-node.stopc:
			return

		// 每 60s 每一个 raft 心跳一次，60s * 10 当前 node 心跳一次
		case <-node.ticker.C:
			ticks++
			klog.Infof(fmt.Sprintf("ticks:%d, HeartbeatTick:%d", ticks, node.config.HeartbeatTick))
			if ticks >= node.config.HeartbeatTick {
				ticks = 0
				node.sendHeartbeat()
			}

			for _, raft := range node.rafts {
				leader, term := raft.leaderTerm()
				klog.Infof(fmt.Sprintf("leader:%d, term:%d", leader, term))
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

func (node *Node) CreateRaft(raftConfig *RaftConfig) error {
	r, err := newRaft(node.config, raftConfig)
	if err != nil {
		return err
	}

	node.rafts[raftConfig.ID] = r

	return nil
}

// Propose INFO: raft node 输入接口，一般来自于上层的输入
func (node *Node) Propose(id uint64, cmd []byte) {
	node.mu.RLock()
	raft, ok := node.rafts[id]
	node.mu.RUnlock()

	if !ok {
		return
	}

	raft.propose(cmd)
}

// INFO: leader 给 peers 发送心跳
func (node *Node) sendHeartbeat() {
	// INFO: 每一个 partition 是一个 raft，并且只有 leader node 上的 partition raft 都是 leader
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
		klog.Infof(fmt.Sprintf("to:%d, ctx:%+v", to, ctx))

		if to == node.config.NodeID {
			continue
		}

		msg := proto.NewMessage()
		msg.Type = proto.ReqMsgHeartBeat
		msg.From = node.config.NodeID
		msg.To = to
		msg.Context = proto.EncodeHBConext(ctx)

		klog.Infof(fmt.Sprintf("[node sendHeartbeat]transport send %+v", *msg))
		//node.config.transport.Send(msg)
		node.heartChan <- msg // 本地调试，不走 transport
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

	// TODO: 发心跳时候为何 message 加 term
	msg := proto.NewMessage()
	msg.Type = proto.RespMsgHeartBeat
	msg.From = node.config.NodeID
	msg.To = message.From
	msg.Context = proto.EncodeHBConext(respCtx)
	klog.Infof(fmt.Sprintf("[node handleHeartbeat]transport send %+v", *msg))
	//node.config.transport.Send(msg)
	node.heartChan <- msg // 本地调试，不走 transport
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
