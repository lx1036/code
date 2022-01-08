package raft

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type Server struct {
	id    uint64
	peers []string
	port  int

	// raft info
	lead           uint64 // must use atomic operations to access; keep 64-bit aligned.
	committedIndex uint64 // must use atomic operations to access; keep 64-bit aligned.

	cluster *Cluster

	node *RaftNode
}

func NewServer() *Server {
	return &Server{}
}

func (server *Server) Start(id int, cluster string, port int) {
	rh := &raftReadyHandler{
		getLead:    func() (lead uint64) { return server.getLead() },
		updateLead: func(lead uint64) { server.setLead(lead) },
		updateLeadership: func(newLeader bool) {
		},
		updateCommittedIndex: func(ci uint64) {
			committedIndex := server.getCommittedIndex()
			if ci > committedIndex {
				server.setCommittedIndex(ci)
			}
		},
	}
	peers := strings.Split(cluster, ",")
	server.peers = peers
	server.port = port
	server.node = NewRaftNode(&Config{
		id:    id,
		peers: peers,
	})
	server.node.start(rh)

	server.cluster = NewCluster(server.node)
	server.cluster.start()

	server.startHTTPService(port)
}

func (server *Server) getLeaderAddr() string {
	value := strings.Split(server.peers[int(server.getLead())-1], "//")[1]
	ip := strings.Split(value, ":")[0]
	return fmt.Sprintf("%s:%d", ip, server.port)
}

func (server *Server) setLead(lead uint64) {
	atomic.StoreUint64(&server.lead, lead)
}

func (server *Server) getLead() uint64 {
	return atomic.LoadUint64(&server.lead)
}

func (server *Server) isLeader() bool {
	return server.ID() == server.Lead()
}

func (server *Server) Lead() uint64 {
	return server.getLead()
}

func (server *Server) ID() uint64 {
	return server.id
}

func (server *Server) setCommittedIndex(committedIndex uint64) {
	atomic.StoreUint64(&server.committedIndex, committedIndex)
}

func (server *Server) getCommittedIndex() uint64 {
	return atomic.LoadUint64(&server.committedIndex)
}
