package raft

import pb "go.etcd.io/etcd/raft/v3/raftpb"

func IsResponseMsg(msgt pb.MessageType) bool {
	return msgt == pb.MsgAppResp || msgt == pb.MsgVoteResp || msgt == pb.MsgHeartbeatResp || msgt == pb.MsgUnreachable || msgt == pb.MsgPreVoteResp
}
