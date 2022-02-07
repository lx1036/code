package raft

import (
	"fmt"
	"io"
	"k8s.io/klog/v2"
)

// SnapshotMeta is for metadata of a snapshot.
type SnapshotMeta struct {
	// ID is opaque to the store, and is used for opening.
	ID string

	// Index and Term store when the snapshot was taken.
	Index uint64
	Term  uint64

	// Configuration and ConfigurationIndex are present in version 1
	// snapshotStore and later.
	Configuration      Configuration
	ConfigurationIndex uint64

	// Size is the size of the snapshot in bytes.
	Size int64
}

// SnapshotStore interface is used to allow for flexible implementations
// of snapshot storage and retrieval. For example, a client could implement
// a shared state store such as S3, allowing new nodes to restore snapshotStore
// without streaming from the leader.
type SnapshotStore interface {
	// Create is used to begin a snapshot at a given index and term, and with
	// the given committed configuration. The version parameter controls
	// which snapshot version to create.
	Create(index, term uint64, configuration Configuration, configurationIndex uint64) (SnapshotSink, error)

	// List is used to list the available snapshotStore in the store.
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
// new snapshotStore of the FSM. It runs in parallel to the FSM and
// main goroutines, so that snapshotStore do not block normal operation.
func (r *Raft) runSnapshots() {
	for {
		select {
		case <-randomTimeout(r.config().SnapshotInterval): // [1s, 2s]
			// Check if we should snapshot
			if !r.shouldSnapshot() {
				continue
			}

			// Trigger a snapshot
			if _, err := r.takeSnapshot(); err != nil {
				klog.Errorf(fmt.Sprintf("failed to take snapshot err:%v", err))
			}

		case future := <-r.userSnapshotCh:
			// User-triggered, run immediately
			id, err := r.takeSnapshot()
			if err != nil {
				klog.Errorf(fmt.Sprintf("failed to take snapshot err:%v", err))
			} else {
				future.opener = func() (*SnapshotMeta, io.ReadCloser, error) {
					return r.snapshotStore.Open(id)
				}
			}
			future.respond(err)

		case <-r.shutdownCh:
			return
		}
	}
}

// shouldSnapshot checks if we meet the conditions to take
// a new snapshot.
func (r *Raft) shouldSnapshot() bool {
	// Check the last snapshot index
	lastSnap, _ := r.getLastSnapshot()

	// Check the last log index
	lastIdx, err := r.logs.LastIndex()
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to get last log index err:%v", err))
		return false
	}

	// Compare the delta to the threshold
	delta := lastIdx - lastSnap
	return delta >= r.config().SnapshotThreshold
}

// takeSnapshot is used to take a new snapshot. This must only be called from
// the snapshot thread, never the main thread. This returns the ID of the new
// snapshot, along with an error.
func (r *Raft) takeSnapshot() (string, error) {
	// Create a request for the FSM to perform a snapshot.
	snapReq := &reqSnapshotFuture{}
	snapReq.init()

	// Wait for dispatch or shutdown.
	select {
	case r.fsmSnapshotCh <- snapReq:
	case <-r.shutdownCh:
		return "", ErrRaftShutdown
	}
	// INFO: Wait until we get a response, 这里是 block，值得借鉴
	if err := snapReq.Error(); err != nil {
		if err != ErrNothingNewToSnapshot {
			err = fmt.Errorf("failed to start snapshot: %v", err)
		}
		return "", err
	}
	defer snapReq.snapshot.Release()

	// INFO: 从 main thread 获取最新 configurations.(committed|committedIndex)，
	//  主要目的是获取 committedIndex 开始到之前做 snapshot
	configReq := &configurationsFuture{}
	configReq.ShutdownCh = r.shutdownCh
	configReq.init()
	select {
	case r.configurationsCh <- configReq:
	case <-r.shutdownCh:
		return "", ErrRaftShutdown
	}
	if err := configReq.Error(); err != nil {
		return "", err
	}
	committed := configReq.configurations.committed
	committedIndex := configReq.configurations.committedIndex

	// We don't support snapshotStore while there's a config change outstanding
	// since the snapshot doesn't have a means to represent this state. This
	// is a little weird because we need the FSM to apply an index that's
	// past the configuration change, even though the FSM itself doesn't see
	// the configuration changes. It should be ok in practice with normal
	// application traffic flowing through the FSM. If there's none of that
	// then it's not crucial that we snapshot, since there's not much going
	// on Raft-wise.
	if snapReq.index < committedIndex {
		return "", fmt.Errorf("cannot take snapshot now, wait until the configuration entry at %v has been applied (have applied %v)",
			committedIndex, snapReq.index)
	}

	// INFO: 所有 logs 中已经 committed 的 pb.LogType_CONFIGURATION log
	klog.Infof(fmt.Sprintf("starting snapshot up to index:%d, the committed configurationIndex is %d in all logs",
		snapReq.index, committedIndex))
	sink, err := r.snapshotStore.Create(snapReq.index, snapReq.term, committed, committedIndex)
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot: %v", err)
	}
	// Try to persist the snapshot.
	if err := snapReq.snapshot.Persist(sink); err != nil {
		sink.Cancel()
		return "", fmt.Errorf("failed to persist snapshot: %v", err)
	}
	// Close and check for error.
	if err := sink.Close(); err != nil {
		return "", fmt.Errorf("failed to close snapshot: %v", err)
	}

	// Update the last stable snapshot info.
	r.setLastSnapshot(snapReq.index, snapReq.term)

	// Compact the logs.
	if err := r.compactLogs(snapReq.index); err != nil {
		return "", err
	}

	klog.Infof(fmt.Sprintf("snapshot complete up to index:%d", snapReq.index))
	return sink.ID(), nil
}

// Snapshot is used to manually force Raft to take a snapshot. Returns a future
// that can be used to block until complete, and that contains a function that
// can be used to open the snapshot.
func (r *Raft) Snapshot() SnapshotFuture {
	future := &userSnapshotFuture{}
	future.init()
	select {
	case r.userSnapshotCh <- future:
		return future
	case <-r.shutdownCh:
		future.respond(ErrRaftShutdown)
		return future
	}
}

// compactLogs takes the last inclusive index of a snapshot
// and trims the logs that are no longer needed.
func (r *Raft) compactLogs(snapIdx uint64) error {
	// Determine log ranges to compact
	minLog, err := r.logs.FirstIndex()
	if err != nil {
		return fmt.Errorf("failed to get first log index: %v", err)
	}

	// Check if we have enough logs to truncate
	lastLogIdx, _ := r.getLastLog()
	// Use a consistent value for trailingLogs for the duration of this method
	// call to avoid surprising behaviour.
	trailingLogs := r.config().TrailingLogs
	if (lastLogIdx - minLog) <= trailingLogs { // TODO: 这里 hashicorp/raft 貌似有个 bug
		return nil
	}

	// Truncate up to the end of the snapshot, or `TrailingLogs`
	// back from the head, which ever is further back. This ensures
	// at least `TrailingLogs` entries, but does not allow logs
	// after the snapshot to be removed.
	maxLog := min(snapIdx, lastLogIdx-trailingLogs)
	if minLog > maxLog {
		klog.Info("no logs to truncate")
		return nil
	}

	klog.Infof(fmt.Sprintf("compacting logs from %d to %d, trailing %d logs until to lastLogIdx:%d",
		minLog, maxLog, trailingLogs, lastLogIdx))

	// Compact the logs
	if err := r.logs.DeleteRange(minLog, maxLog); err != nil {
		return fmt.Errorf("log compaction failed: %v", err)
	}
	return nil
}

// restoreUserSnapshot is used to manually consume an external snapshot, such
// as if restoring from a backup. We will use the current Raft configuration,
// not the one from the snapshot, so that we can restore into a new cluster. We
// will also use the higher of the index of the snapshot, or the current index,
// and then add 1 to that, so we force a new state with a hole in the Raft log,
// so that the snapshot will be sent to followers and used for any new joiners.
// This can only be run on the leader, and returns a future that can be used to
// block until complete.
func (r *Raft) restoreUserSnapshot(meta *SnapshotMeta, reader io.Reader) error {
	// We don't support snapshots while there's a config change
	// outstanding since the snapshot doesn't have a means to
	// represent this state.
	committedIndex := r.configurations.committedIndex
	latestIndex := r.configurations.latestIndex
	if committedIndex != latestIndex {
		return fmt.Errorf("cannot restore snapshot now, wait until the configuration entry at %v has been applied (have applied %v)",
			latestIndex, committedIndex)
	}

	// Cancel any inflight requests.
	for {
		e := r.leaderState.inflight.Front()
		if e == nil {
			break
		}
		e.Value.(*logFuture).respond(ErrAbortedByRestore)
		r.leaderState.inflight.Remove(e)
	}

	// We will overwrite the snapshot metadata with the current term,
	// an index that's greater than the current index, or the last
	// index in the snapshot. It's important that we leave a hole in
	// the index so we know there's nothing in the Raft log there and
	// replication will fault and send the snapshot.
	term := r.getCurrentTerm()
	lastIndex := r.getLastIndex()
	if meta.Index > lastIndex {
		lastIndex = meta.Index
	}
	lastIndex++

	// Dump the snapshot. Note that we use the latest configuration,
	// not the one that came with the snapshot.
	sink, err := r.snapshotStore.Create(lastIndex, term,
		r.configurations.latest, r.configurations.latestIndex)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}
	n, err := io.Copy(sink, reader)
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("failed to write snapshot: %v", err)
	}
	if n != meta.Size {
		sink.Cancel()
		return fmt.Errorf("failed to write snapshot, size didn't match (%d != %d)", n, meta.Size)
	}
	if err := sink.Close(); err != nil {
		return fmt.Errorf("failed to close snapshot: %v", err)
	}
	klog.Infof(fmt.Sprintf("copied to local snapshot %d bytes", n))

	// Restore the snapshot into the FSM. If this fails we are in a
	// bad state so we panic to take ourselves out.
	restore := &restoreFuture{ID: sink.ID()}
	restore.ShutdownCh = r.shutdownCh
	restore.init()
	select {
	case r.fsmMutateCh <- restore:
	case <-r.shutdownCh:
		return ErrRaftShutdown
	}
	if err := restore.Error(); err != nil {
		panic(fmt.Errorf("failed to restore snapshot: %v", err))
	}

	// We set the last log so it looks like we've stored the empty
	// index we burned. The last applied is set because we made the
	// FSM take the snapshot state, and we store the last snapshot
	// in the stable store since we created a snapshot as part of
	// this process.
	r.setLastLog(lastIndex, term)
	r.setLastApplied(lastIndex)
	r.setLastSnapshot(lastIndex, term)

	klog.Infof(fmt.Sprintf("restored user snapshot from index:%d", latestIndex))
	return nil
}
