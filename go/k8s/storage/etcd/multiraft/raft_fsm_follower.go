package multiraft

import (
	"k8s-lx1036/k8s/storage/raft/proto"
)

func (r *raftFsm) becomeFollower(term, lead uint64) {

	r.step = r.stepFollower
	r.leader = lead
	r.state = stateFollower

}

func (r *raftFsm) stepFollower(message *proto.Message) {

}
