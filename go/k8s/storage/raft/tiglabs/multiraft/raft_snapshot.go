package multiraft

import (
	"k8s-lx1036/k8s/storage/raft/proto"
	"k8s-lx1036/k8s/storage/raft/util"
)

type snapshotReader struct {
	reader *util.BufferReader
	err    error
}

type snapshotRequest struct {
	respErr
	snapshotReader
	header *proto.Message
}

type snapshotStatus struct {
	respErr
	stopCh chan struct{}
}

func newSnapshotStatus() *snapshotStatus {
	f := &snapshotStatus{
		stopCh: make(chan struct{}),
	}
	f.init()
	return f
}
