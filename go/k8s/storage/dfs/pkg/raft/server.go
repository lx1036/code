package raft

import (
	"sync"
	"time"
)

type RaftServer struct {
	config *Config
	ticker *time.Ticker
	heartc chan *proto.Message
	stopc  chan struct{}
	mu     sync.RWMutex
	rafts  map[uint64]*raft
}
