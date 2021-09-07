package raft

type LogItem struct {
	Term  int64  `json:"term"`
	Index int64  `json:"index"`
	Data  []byte `json:"data"`
}

type RaftLog interface {
	Index(int64) *LogItem

	AppendLog([]*LogItem, ...int64) (int64, error)
}
