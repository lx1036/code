package multiraft

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/raft/proto"

	"k8s.io/klog/v2"
)

func (r *raftFsm) becomeLeader() {

	//lasti := r.raftLog.LastIndex()
	r.step = r.stepLeader
	//r.reset(r.term, lasti, true)
	//r.tick = r.tickHeartbeat
	r.leader = r.nodeConfig.NodeID
	r.state = stateLeader
	r.acks = nil

}

func (r *raftFsm) stepLeader(message *proto.Message) {

	// INFO: replica 表示 leader 能看到 follower 追赶的进度, etcd raft 里用的 progress 对象
	replica, ok := r.replicas[message.From]
	if !ok {
		klog.Warningf(fmt.Sprintf("[raftFsm stepLeader]raftFsm[%v] no progress available for %v", r.id, message.From))
		return
	}

	switch message.Type {

	case proto.RespMsgHeartBeat:
		replica.active = true
		replica.lastActive = time.Now()

		klog.Infof(fmt.Sprintf("[raftFsm stepLeader]RespMsgHeartBeat %+v", *replica))

	}

}

func (r *raftFsm) bcastAppend() {
	for id := range r.replicas {
		if id == r.nodeConfig.NodeID {
			continue
		}

		r.sendAppend(id)
	}
}

func (r *raftFsm) sendAppend(to uint64) {

}
