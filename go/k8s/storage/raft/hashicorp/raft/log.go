package raft

import (
	"errors"
	pb "k8s-lx1036/k8s/storage/raft/hashicorp/raft/rpc"
)

var (
	// ErrLogNotFound indicates a given log entry is not available.
	ErrLogNotFound = errors.New("log not found")
)

// LogStore is used to provide an interface for storing
// and retrieving logs in a durable fashion.
type LogStore interface {
	// FirstIndex returns the first index written. 0 for no entries.
	FirstIndex() (uint64, error)

	// LastIndex returns the last index written. 0 for no entries.
	LastIndex() (uint64, error)

	// GetLog gets a log entry at a given index.
	GetLog(index uint64, log *pb.Log) error

	// StoreLog stores a log entry.
	StoreLog(log *pb.Log) error

	// StoreLogs stores multiple log entries.
	StoreLogs(logs []*pb.Log) error

	// DeleteRange deletes a range of log entries. The range is inclusive.
	DeleteRange(min, max uint64) error
}

// StableStore is used to provide stable storage
// of key configurations to ensure safety.
type StableStore interface {
	Set(key []byte, val []byte) error

	// Get returns the value for key, or an empty byte slice if key was not found.
	Get(key []byte) ([]byte, error)

	SetUint64(key []byte, val uint64) error

	// GetUint64 returns the uint64 value for key, or 0 if key was not found.
	GetUint64(key []byte) (uint64, error)
}
