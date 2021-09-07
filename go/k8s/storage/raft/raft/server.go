package raft

import "fmt"

type Server struct {
	raft *Raft

	nodes []*Node

	currentNode *Node
}

func (server *Server) Start() error {

	err := server.raft.Start()

	return err
}

func (server *Server) Leader() (*Node, error) {
	// 如果当前节点是 Leader，直接返回当前 Node
	if server.raft.state.Role == Leader {
		return server.currentNode, nil
	}

	// http 查询 peers 哪个是 Leader
	peers := server.raft.config.GetPeers()

	for _, peer := range peers {
		response, err := server.raft.transport.State(peer.Node)
		if err != nil {
			continue
		}

		if response.State.Role == Leader {
			return peer.Node, nil
		}
	}

	return nil, fmt.Errorf("no leader")
}

func (server *Server) Role() Role {
	return server.raft.state.Role
}

func NewServer(config *Config, stateMachine StateMachine, nodes []*Node, currentNode *Node) (*Server, error) {

	httpTransport := NewHttpTransport()
	raft, err := NewRaft(config, httpTransport, stateMachine)
	if err != nil {
		return nil, err
	}

	server := &Server{
		raft:        raft,
		nodes:       nodes,
		currentNode: currentNode,
	}

	return server, nil
}
