package multiraft

import (
	"errors"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/raft/proto"

	"k8s.io/klog/v2"
)

var (
	fatalStopc = make(chan uint64)
)

// NoLeader is a placeholder nodeID used when there is no leader.
const NoLeader uint64 = 0

// RaftServer 只是一个包装Raft对象的类，关键是rafts属性
type RaftServer struct {
	config    *Config
	ticker    *time.Ticker
	heartChan chan *proto.Message
	stopc     chan struct{}
	mu        sync.RWMutex
	rafts     map[uint64]*Raft
}

// INFO: (1)
func NewRaftServer(config *Config) (*RaftServer, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	rs := &RaftServer{
		config:    config,
		ticker:    time.NewTicker(config.TickInterval),
		rafts:     make(map[uint64]*Raft),
		heartChan: make(chan *proto.Message, 512),
		stopc:     make(chan struct{}),
	}
	if transport, err := NewMultiTransport(rs, &config.TransportConfig); err != nil {
		return nil, err
	} else {
		rs.config.transport = transport
	}

	go rs.run()
	return rs, nil
}

// INFO: (2) 创建 RaftServer.rafts
func (rs *RaftServer) CreateRaft(raftConfig *RaftConfig) error {
	var (
		raft *Raft
		err  error
	)

	defer func() {
		if err != nil {
			klog.Error("CreateRaft [%v] failed, error is:\r\n %s", raftConfig.ID, err.Error())
		}
	}()

	if raft, err = newRaft(rs.config, raftConfig); err != nil {
		return err
	}
	if raft == nil {
		return errors.New("CreateRaft return nil, maybe occur panic.")
	}

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, ok := rs.rafts[raftConfig.ID]; ok {
		raft.stop()
		err = ErrRaftExists
		return err
	}

	// 创建rafts实例
	rs.rafts[raftConfig.ID] = raft
	return nil
}

// raftServer.run 主要发送心跳来确认存活
func (rs *RaftServer) run() {
	ticks := 0
	for {
		select {
		case <-rs.stopc:
			return

		case id := <-fatalStopc:
			rs.mu.Lock()
			delete(rs.rafts, id)
			rs.mu.Unlock()

		case m := <-rs.heartChan:
			switch m.Type {
			case proto.ReqMsgHeartBeat:
				rs.handleHeartbeat(m)
			case proto.RespMsgHeartBeat:
				rs.handleHeartbeatResp(m)
			}

		case <-rs.ticker.C:
			ticks++
			if ticks >= rs.config.HeartbeatTick {
				ticks = 0
				rs.sendHeartbeat()
			}

			rs.mu.RLock()
			for _, raft := range rs.rafts {
				raft.tick()
			}
			rs.mu.RUnlock()
		}
	}
}

// INFO: 处理心跳请求
func (rs *RaftServer) handleHeartbeat(message *proto.Message) {
	rs.mu.RLock() // TODO: 这里为何需要加锁
	ctx := proto.DecodeHBContext(message.Context)
	var respCtx proto.HeartbeatContext
	for _, id := range ctx { // id 是每一个字节的 uint64 值
		if raft, ok := rs.rafts[id]; ok {
			raft.receiveMessage(message)
			respCtx = append(respCtx, id)
		}
	}

	rs.mu.RUnlock()

	msg := proto.NewMessage()
	msg.Type = proto.RespMsgHeartBeat
	msg.From = rs.config.NodeID
	msg.To = message.From
	msg.Context = proto.EncodeHBConext(respCtx)
	rs.config.transport.Send(msg)
}

func (rs *RaftServer) handleHeartbeatResp(message *proto.Message) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	ctx := proto.DecodeHBContext(message.Context)
	for _, id := range ctx {
		if raft, ok := rs.rafts[id]; ok {
			raft.receiveMessage(message)
		}
	}
}

// leader给peers发送心跳
func (rs *RaftServer) sendHeartbeat() {
	// key: sendto nodeId; value: range ids
	nodes := make(map[uint64]proto.HeartbeatContext)
	rs.mu.RLock()
	for id, raft := range rs.rafts {
		if !raft.isLeader() {
			continue
		}
		peers := raft.getPeers()
		for _, p := range peers {
			nodes[p] = append(nodes[p], id)
		}
	}
	rs.mu.RUnlock()

	for to, ctx := range nodes {
		if to == rs.config.NodeID {
			continue
		}

		msg := proto.NewMessage()
		msg.Type = proto.ReqMsgHeartBeat
		msg.From = rs.config.NodeID
		msg.To = to
		msg.Context = proto.EncodeHBConext(ctx)
		rs.config.transport.Send(msg)
	}
}

func (rs *RaftServer) receiveMessage(message *proto.Message) {
	if message.Type == proto.ReqMsgHeartBeat || message.Type == proto.RespMsgHeartBeat {
		rs.heartChan <- message
		return
	}

	rs.mu.RLock()
	defer rs.mu.RUnlock()
	raft, ok := rs.rafts[message.ID]
	if ok {
		raft.receiveMessage(message)
	}
}

// ReadIndex read index
func (rs *RaftServer) ReadIndex(id uint64) (future *Future) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	raft, ok := rs.rafts[id]
	future = newFuture()
	if !ok {
		future.respond(nil, ErrRaftNotExists)
		return
	}

	raft.readIndex(future)
	return
}

// Write 向 raft 状态机里提交 cmd
func (rs *RaftServer) Submit(id uint64, cmd []byte) (future *Future) {
	rs.mu.RLock()
	raft, ok := rs.rafts[id]
	rs.mu.RUnlock()

	future = newFuture()
	if !ok {
		future.respond(nil, ErrRaftNotExists)
		return
	}

	raft.propose(cmd, future)
	return
}

func (rs *RaftServer) LeaderTerm(id uint64) (leader, term uint64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	raft, ok := rs.rafts[id]
	if ok {
		return raft.leaderTerm()
	}

	return NoLeader, 0
}
