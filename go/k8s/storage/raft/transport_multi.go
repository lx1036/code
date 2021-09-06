package raft

import (
	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/util"
)

type MultiTransport struct {
	heartbeat *heartbeatTransport
	replicate *replicateTransport
}

func NewMultiTransport(raft *RaftServer, config *TransportConfig) (Transport, error) {
	mt := new(MultiTransport)

	if ht, err := newHeartbeatTransport(raft, config); err != nil {
		return nil, err
	} else {
		mt.heartbeat = ht
	}
	if rt, err := newReplicateTransport(raft, config); err != nil {
		return nil, err
	} else {
		mt.replicate = rt
	}

	mt.heartbeat.start()
	mt.replicate.start()

	return mt, nil
}

func receiveMessage(r *util.BufferReader) (msg *proto.Message, err error) {
	msg = proto.NewMessage()
	if err = msg.Decode(r); err != nil {
		proto.ReturnMessage(msg)
	}
	return
}
