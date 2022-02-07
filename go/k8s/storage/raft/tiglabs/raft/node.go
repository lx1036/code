package raft

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"k8s.io/klog/v2"
)

type Config struct {
	LocalID   string
	LocalAddr string

	HeartbeatTimeout time.Duration

	ElectionTimeout time.Duration
}

type Node struct {
	conf atomic.Value
	// Stores our local server ID, used to avoid sending RPCs to ourself
	localID string
	// Stores our local addr
	localAddr string

	rafts map[string]*Raft

	transport Transport

	shutdownCh chan struct{}
}

func NewNode(conf *Config, transport Transport) (*Node, error) {

	node := &Node{
		localID:   conf.LocalID,
		localAddr: conf.LocalAddr,

		transport: transport,

		shutdownCh: make(chan struct{}),
	}

	node.conf.Store(*conf)

	transport.SetHeartbeatHandler()

	go node.run()
}

func (node *Node) config() Config {
	return node.conf.Load().(Config)
}

func (node *Node) run() {
	for {
		select {
		case <-randomTimeout(node.config().HeartbeatTimeout / 10):
			var nodes map[string][]string
			for id, raft := range node.rafts {
				if !raft.isLeader() {
					continue
				}

				peers := raft.getPeers()
				for _, p := range peers {
					nodes[p] = append(nodes[p], id)
				}
			}

			for peer, raftIDs := range nodes {
				node.transport.AppendEntries(peer, raftIDs, request, resp)
			}

		case <-node.shutdownCh:
			return
		}
	}
}

func (node *Node) processHeartbeat(rpc RPC) {
	select {
	case <-node.shutdownCh:
		return
	default:
	}

	// Ensure we are only handling a heartbeat
	switch cmd := rpc.Command.(type) {
	case *AppendEntriesRequest:
		node.appendEntries(rpc, cmd)
	default:
		klog.Error(fmt.Sprintf("expected heartbeat, got command: %+v", rpc.Command))
		rpc.Respond(nil, fmt.Errorf("unexpected command"))
	}
}

func (node *Node) appendEntries(rpc RPC, cmd *AppendEntriesRequest) {
	resp := &AppendEntriesResponse{
		Term:           r.getCurrentTerm(),
		LastLog:        r.getLastIndex(),
		Success:        false,
		NoRetryBackoff: false,
	}

	var rpcErr error
	defer func() {
		rpc.Respond(resp, rpcErr)
	}()

	// Ignore an older term
	if cmd.Term < r.getCurrentTerm() {
		return
	}

	// Everything went well, set success
	resp.Success = true
	// raft paper: "For example, you might reasonably reset a peer’s election timer whenever you receive an AppendEntries or RequestVote RPC."
	r.setLastContact()
	return
}
