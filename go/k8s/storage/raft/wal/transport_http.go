package wal

import "time"

type Transport interface {
	State(*Node) (*QueryStateResponse, error)
}

type QueryStateResponse struct {
	State RaftState `json:"state"`
	Peer  Peers     `json:"peers"`
}

type HttpTransport struct {
	timeout time.Duration
}

func (transport *HttpTransport) State(node *Node) (*QueryStateResponse, error) {

}

func NewHttpTransport() Transport {
	return &HttpTransport{
		timeout: 3 * time.Second,
	}
}
