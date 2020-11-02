package raft

import (
	"log"

)


type raftLog struct {
	// storage contains all stable entries since the last snapshot.
	storage Storage
	
	// unstable contains all unstable entries and snapshot.
	// they will be saved into storage.
	unstable unstable
	
	// committed is the highest log position that is known to be in
	// stable storage on a quorum of nodes.
	committed uint64
	
	// applied is the highest log position that the application has
	// been instructed to apply to its state machine.
	// Invariant: applied <= committed
	applied uint64
	
	// maxNextEntsSize is the maximum number aggregate byte size of the messages
	// returned from calls to nextEnts.
	maxNextEntsSize uint64
	
	logger Logger
}


// newLogWithSize returns a log using the given storage and max
// message size.
func newLogWithSize(storage Storage, logger Logger, maxNextEntsSize uint64) *raftLog {
	if storage == nil {
		log.Panic("storage must not be nil")
	}
	
	rLog := &raftLog{
		storage:         storage,
		logger:          logger,
		maxNextEntsSize: maxNextEntsSize,
	}
	
	firstIndex, err := storage.FirstIndex()
	if err != nil {
		panic(err)
	}
	lastIndex, err := storage.LastIndex()
	if err != nil {
		panic(err)
	}
	rLog.unstable.offset = lastIndex + 1
	rLog.unstable.logger = logger
	// Initialize our committed and applied pointers to the time of the last compaction.
	rLog.committed = firstIndex - 1
	rLog.applied = firstIndex - 1
	
	return rLog
}
