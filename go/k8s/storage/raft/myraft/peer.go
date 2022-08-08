package myraft

type Peers []*Peer

type Peer struct {
	*Node
}
