package multiraft

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/raft/proto"

	"k8s.io/klog/v2"
)

func (r *raftFsm) becomeLeader() {
	if r.state == stateFollower {
		klog.Fatalf(fmt.Sprintf("[raft->becomeLeader][%v] invalid transition [follower -> leader].", r.id))
	}

	lastIndex := r.log.LastIndex()
	r.step = r.stepLeader
	r.reset(r.term, lastIndex, true)
	r.tick = r.tickHeartbeat
	r.leader = r.nodeConfig.NodeID
	r.state = stateLeader
	r.acks = nil
	if pr, ok := r.replicas[r.nodeConfig.NodeID]; ok {
		pr.active = true
	}

	ents, err := r.log.entries(r.log.committed+1, noLimit)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("[raft->becomeLeader][%v] unexpected error getting uncommitted entries (%v).", r.id, err))
	}
	nconf := numOfPendingConf(ents)
	if nconf > 1 {
		klog.Fatalf(fmt.Sprintf("[raft->becomeLeader][%v] unexpected double uncommitted config entry.", r.id))
	}
	if nconf == 1 {
		r.pendingConf = true
	}

	r.appendEntry(&proto.Entry{Term: r.term, Index: lastIndex + 1, Data: nil})
	klog.Infof(fmt.Sprintf("raft[%v] became leader at term %d.", r.id, r.term))
}

func (r *raftFsm) tickHeartbeat() {

}

func (r *raftFsm) stepLeader(message *proto.Message) {

	// INFO: replica 表示 leader 能看到 follower 追赶的进度, etcd raft 里用的 progress 对象
	replica, ok := r.replicas[message.From]
	if !ok {
		klog.Warningf(fmt.Sprintf("[raftFsm stepLeader]raftFsm[%v] no progress available for %v", r.id, message.From))
		return
	}

	klog.Info(fmt.Sprintf("message Type %s", message.Type))

	switch message.Type {

	case proto.RespMsgHeartBeat:
		replica.active = true
		replica.lastActive = time.Now()

		klog.Infof(fmt.Sprintf("[raftFsm stepLeader]RespMsgHeartBeat %+v", *replica))

	case proto.LocalMsgProp:
		if _, ok := r.replicas[r.nodeConfig.NodeID]; !ok || len(message.Entries) == 0 {
			return
		}

		r.appendEntry(message.Entries...) // commit 到自己的 raft log 模块中
		r.broadcastAppend()               // 广播给 follower

	}

}

// INFO: leader 先提交到自己的 raft log 中
func (r *raftFsm) appendEntry(entries ...*proto.Entry) {
	r.log.append(entries...)

	// 记录下leader自己 commit raft log 中的 committed index 记录
	r.replicas[r.nodeConfig.NodeID].maybeUpdate(r.log.LastIndex(), r.log.committed)

	//r.maybeCommit()
}

func (r *raftFsm) broadcastAppend() {
	for id := range r.replicas {
		if id == r.nodeConfig.NodeID {
			continue
		}

		r.sendAppend(id)
	}
}

func (r *raftFsm) sendAppend(to uint64) {
	replica := r.replicas[to]
	if replica.isPaused() {
		return
	}

	var (
		term       uint64
		ents       []*proto.Entry
		errt, erre error
		m          *proto.Message
	)
	firstIndex := r.log.firstIndex()
	if replica.next >= firstIndex {
		term, errt = r.log.term(replica.next - 1)
		ents, erre = r.log.entries(replica.next, r.nodeConfig.MaxSizePerMsg)
	}

	if replica.next < firstIndex || errt != nil || erre != nil {
		if !replica.active {
			klog.Infof(fmt.Sprintf("[raft->sendAppend][%v]ignore sending snapshot to %v since it is not recently active.", r.id, to))
			return
		}

	} else {
		m = proto.NewMessage()
		m.Type = proto.ReqMsgAppend
		m.To = to
		m.Index = replica.next - 1
		m.LogTerm = term
		m.Commit = r.log.committed
		m.Entries = append(m.Entries, ents...)
		if n := len(m.Entries); n != 0 {
			switch replica.state {
			case replicaStateReplicate:
				last := m.Entries[n-1].Index
				replica.update(last)
				replica.inflight.add(last)
			case replicaStateProbe:
				replica.pause()
			default:
				klog.Fatalf(fmt.Sprintf("node %x is sending append in unhandled state %s", r.id, replica.state))
			}
		}
	}

	replica.pending = true
	r.send(m)
}
