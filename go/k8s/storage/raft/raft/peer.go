package raft

type Peers []*Peer

type Peer struct {
	*Node
}
