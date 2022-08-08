package myraft

import "time"

type Transport interface {
	State(*Node) (*QueryStateResponse, error)
}

type QueryStateResponse struct {
	State RaftState `json:"state"`
	Peer  Peers     `json:"peers"`
}

type AppendLogRequest struct {
	Term int64 `json:"term"`

	// append log request 需要带上上一个request(index/term)
	PrevLogIndex int64 `json:"prevLogIndex"`
	PrevLogTerm  int64 `json:"prevLogTerm"`

	LogItems []*LogItem `json:"entries"`

	LeaderCommit int64 `json:"leaderCommit"`
}

type AppendLogResponse struct {
	Term       int64 `json:"term"`
	Success    bool  `json:"success"`
	MatchIndex int64 `json:"matchIndex"` // -1: error 0: not match  >0: match
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
