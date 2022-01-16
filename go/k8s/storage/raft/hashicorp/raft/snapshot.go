package raft

import (
	"io"
)

// SnapshotMeta is for metadata of a snapshot.
type SnapshotMeta struct {
	// ID is opaque to the store, and is used for opening.
	ID string

	// Index and Term store when the snapshot was taken.
	Index uint64
	Term  uint64

	// Configuration and ConfigurationIndex are present in version 1
	// snapshots and later.
	Configuration      Configuration
	ConfigurationIndex uint64

	// Size is the size of the snapshot in bytes.
	Size int64
}

// SnapshotStore interface is used to allow for flexible implementations
// of snapshot storage and retrieval. For example, a client could implement
// a shared state store such as S3, allowing new nodes to restore snapshots
// without streaming from the leader.
type SnapshotStore interface {
	// Create is used to begin a snapshot at a given index and term, and with
	// the given committed configuration. The version parameter controls
	// which snapshot version to create.
	Create(index, term uint64, configuration Configuration,
		configurationIndex uint64, trans Transport) (SnapshotSink, error)

	// List is used to list the available snapshots in the store.
	// It should return then in descending order, with the highest index first.
	List() ([]*SnapshotMeta, error)

	// Open takes a snapshot ID and provides a ReadCloser. Once close is
	// called it is assumed the snapshot is no longer needed.
	Open(id string) (*SnapshotMeta, io.ReadCloser, error)
}

// SnapshotSink is returned by StartSnapshot. The FSM will Write state
// to the sink and call Close on completion. On error, Cancel will be invoked.
type SnapshotSink interface {
	io.WriteCloser
	ID() string
	Cancel() error
}

// runSnapshots is a long running goroutine used to manage taking
// new snapshots of the FSM. It runs in parallel to the FSM and
// main goroutines, so that snapshots do not block normal operation.
func (r *Raft) runSnapshots() {

}
