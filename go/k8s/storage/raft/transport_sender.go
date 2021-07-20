package raft

import (
	"sync"

	"k8s-lx1036/k8s/storage/raft/proto"
)

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
