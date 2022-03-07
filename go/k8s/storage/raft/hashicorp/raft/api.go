package raft

import (
	"errors"
	"fmt"
	"io"
	"time"

	pb "k8s-lx1036/k8s/storage/raft/hashicorp/raft/rpc"

	"k8s.io/klog/v2"
)

var (
	ErrRaftShutdown = errors.New("raft is already shutdown")

	ErrCantBootstrap = errors.New("bootstrap only works on new clusters")

	// ErrEnqueueTimeout is returned when a command fails due to a timeout.
	ErrEnqueueTimeout = errors.New("timed out enqueuing operation")

	// ErrLeadershipTransferInProgress is returned when the leader is rejecting
	// client requests because it is attempting to transfer leadership.
	ErrLeadershipTransferInProgress = errors.New("leadership transfer in progress")

	// ErrNotLeader is returned when an operation can't be completed on a
	// follower or candidate node.
	ErrNotLeader = errors.New("node is not the leader")

	// ErrNothingNewToSnapshot is returned when trying to create a snapshot
	// but there's nothing new commited to the FSM since we started.
	ErrNothingNewToSnapshot = errors.New("nothing new to snapshot")

	// ErrAbortedByRestore is returned when a leader fails to commit a log
	// entry because it's been superseded by a user snapshot restore.
	ErrAbortedByRestore = errors.New("snapshot restored while committing log")
)

// Apply is used to apply a command to the FSM in a highly consistent
// manner. This returns a future that can be used to wait on the application.
// An optional timeout can be provided to limit the amount of time we wait
// for the command to be started. This must be run on the leader or it
// will fail.
func (r *Raft) Apply(cmd []byte, timeout time.Duration) ApplyFuture {
	return r.ApplyLog(&pb.Log{Data: cmd}, timeout)
}

// ApplyLog performs Apply but takes in a Log directly. The only values
// currently taken from the submitted Log are Data and Extensions.
func (r *Raft) ApplyLog(log *pb.Log, timeout time.Duration) ApplyFuture {
	var timer <-chan time.Time
	if timeout > 0 {
		timer = time.After(timeout)
	}

	// Create a log future, no index or term yet
	f := &logFuture{
		log: pb.Log{
			Type:       pb.LogType_COMMAND,
			Data:       log.Data,
			Extensions: log.Extensions,
		},
	}
	f.init()

	select {
	case <-timer:
		return errorFuture{ErrEnqueueTimeout}
	case <-r.shutdownCh:
		return errorFuture{ErrRaftShutdown}
	case r.applyCh <- f:
		return f
	}
}

// Restore is used to manually force Raft to consume an external snapshot, such
// as if restoring from a backup. We will use the current Raft configuration,
// not the one from the snapshot, so that we can restore into a new cluster. We
// will also use the higher of the index of the snapshot, or the current index,
// and then add 1 to that, so we force a new state with a hole in the Raft log,
// so that the snapshot will be sent to followers and used for any new joiners.
// This can only be run on the leader, and blocks until the restore is complete
// or an error occurs.
//
// WARNING! This operation has the leader take on the state of the snapshot and
// then sets itself up so that it replicates that to its followers though the
// install snapshot process. This involves a potentially dangerous period where
// the leader commits ahead of its followers, so should only be used for disaster
// recovery into a fresh cluster, and should not be used in normal operations.
func (r *Raft) Restore(meta *SnapshotMeta, reader io.Reader, timeout time.Duration) error {
	var timer <-chan time.Time
	if timeout > 0 {
		timer = time.After(timeout)
	}

	// Perform the restore.
	restore := &userRestoreFuture{
		meta:   meta,
		reader: reader,
	}
	restore.init()
	select {
	case <-timer:
		return ErrEnqueueTimeout
	case <-r.shutdownCh:
		return ErrRaftShutdown
	case r.userRestoreCh <- restore:
		// If the restore is ingested then wait for it to complete.
		if err := restore.Error(); err != nil {
			return err
		}
	}

	// Apply a no-op log entry. Waiting for this allows us to wait until the
	// followers have gotten the restore and replicated at least this new
	// entry, which shows that we've also faulted and installed the
	// snapshot with the contents of the restore.
	noop := &logFuture{
		log: pb.Log{
			Type: pb.LogType_NOOP,
		},
	}
	noop.init()
	select {
	case <-timer:
		return ErrEnqueueTimeout
	case <-r.shutdownCh:
		return ErrRaftShutdown
	case r.applyCh <- noop:
		return noop.Error()
	}
}

func (r *Raft) State() RaftState {
	return r.getState()
}

// LeaderCh is used to get a channel which delivers signals on acquiring or
// losing leadership. It sends true if we become the leader, and false if we
// lose it.
//
// Receivers can expect to receive a notification only if leadership
// transition has occured.
//
// If receivers aren't ready for the signal, signals may drop and only the
// latest leadership transition. For example, if a receiver receives subsequent
// `true` values, they may deduce that leadership was lost and regained while
// the receiver was processing first leadership transition.
func (r *Raft) LeaderCh() <-chan bool {
	return r.leaderCh
}

// BootstrapCluster initializes a server's storage with the given cluster
// configuration. This should only be called at the beginning of time for the
// cluster with an identical configuration listing all Voter servers. There is
// no need to bootstrap Nonvoter and Staging servers.
//
// A cluster can only be bootstrapped once from a single participating Voter
// server. Any further attempts to bootstrap will return an error that can be
// safely ignored.
//
// One approach is to bootstrap a single server with a configuration
// listing just itself as a Voter, then invoke AddVoter() on it to add other
// servers to the cluster.
func BootstrapCluster(conf *Config, logs LogStore, stable StableStore, snaps SnapshotStore, configuration Configuration) error {
	// Validate the Raft server config.
	if err := ValidateConfig(conf); err != nil {
		return err
	}

	// Sanity check the Raft peer configuration.
	if err := checkConfiguration(configuration); err != nil {
		return err
	}

	// Make sure the cluster is in a clean state.
	hasState, err := HasExistingState(logs, stable, snaps)
	if err != nil {
		return fmt.Errorf("failed to check for existing state: %v", err)
	}
	if hasState {
		return ErrCantBootstrap
	}

	// Set current term to 1.
	if err := stable.SetUint64(keyCurrentTerm, 1); err != nil {
		return fmt.Errorf("failed to save current term: %v", err)
	}
	// Append configuration entry to log.
	entry := &pb.Log{
		Index: 1,
		Term:  1,
		Type:  pb.LogType_CONFIGURATION,
		Data:  EncodeConfiguration(configuration),
	}
	if err := logs.StoreLog(entry); err != nil {
		return fmt.Errorf("failed to append configuration entry to log: %v", err)
	}

	return nil
}

// RecoverCluster is used to manually force a new configuration in order to
// recover from a loss of quorum where the current configuration cannot be
// restored, such as when several servers die at the same time. This works by
// reading all the current state for this server, creating a snapshot with the
// supplied configuration, and then truncating the Raft log. This is the only
// safe way to force a given configuration without actually altering the log to
// insert any new entries, which could cause conflicts with other servers with
// different state.
//
// WARNING! This operation implicitly commits all entries in the Raft log, so
// in general this is an extremely unsafe operation. If you've lost your other
// servers and are performing a manual recovery, then you've also lost the
// commit information, so this is likely the best you can do, but you should be
// aware that calling this can cause Raft log entries that were in the process
// of being replicated but not yet be committed to be committed.
//
// Note the FSM passed here is used for the snapshot operations and will be
// left in a state that should not be used by the application. Be sure to
// discard this FSM and any associated state and provide a fresh one when
// calling NewRaft later.
//
// A typical way to recover the cluster is to shut down all servers and then
// run RecoverCluster on every server using an identical configuration. When
// the cluster is then restarted, and election should occur and then Raft will
// resume normal operation. If it's desired to make a particular server the
// leader, this can be used to inject a new configuration with that server as
// the sole voter, and then join up other new clean-state peer servers using
// the usual APIs in order to bring the cluster back into a known state.
func RecoverCluster(conf *Config, fsm FSM, logs LogStore, stable StableStore,
	snaps SnapshotStore, trans Transport, configuration Configuration) error {
	// Validate the Raft server config.
	if err := ValidateConfig(conf); err != nil {
		return err
	}

	// Sanity check the Raft peer configuration.
	if err := checkConfiguration(configuration); err != nil {
		return err
	}

	// Make sure the cluster is in a clean state.
	hasState, err := HasExistingState(logs, stable, snaps)
	if err != nil {
		return fmt.Errorf("failed to check for existing state: %v", err)
	}
	if !hasState {
		return fmt.Errorf("refused to recover cluster with no initial state, this is probably an operator error")
	}

	// Attempt to restore any snapshotStore we find, newest to oldest.
	var (
		snapshotIndex uint64
		snapshotTerm  uint64
	)
	snapshots, err := snaps.List()
	if err != nil {
		klog.Errorf(fmt.Sprintf("failed to list snapshotStore err:%v", err))
		return err
	}
	// Try to load in order of newest to oldest
	for _, snapshot := range snapshots {
		if !conf.NoSnapshotRestoreOnStart {
			_, source, err := snaps.Open(snapshot.ID)
			if err != nil {
				klog.Errorf(fmt.Sprintf("failed to open snapshot id:%s err:%v", snapshot.ID, err))
				continue
			}

			if err := fsm.Restore(source); err != nil {
				source.Close()
				klog.Errorf(fmt.Sprintf("failed to restore snapshot id:%s err:%v", snapshot.ID, err))
				continue
			}

			source.Close()
			klog.Infof(fmt.Sprintf("restored from snapshot id:%s", snapshot.ID))
		}

		snapshotIndex = snapshot.Index
		snapshotTerm = snapshot.Term
		break
	}
	if len(snapshots) > 0 && (snapshotIndex == 0 || snapshotTerm == 0) {
		return fmt.Errorf("failed to restore any of the available snapshotStore")
	}

	// The snapshot information is the best known end point for the data
	// until we play back the Raft log entries.
	lastIndex := snapshotIndex
	lastTerm := snapshotTerm
	// Apply any Raft log entries from the snapshot index to last log index.
	lastLogIndex, err := logs.LastIndex()
	if err != nil {
		return fmt.Errorf("failed to find last log: %v", err)
	}
	for index := snapshotIndex + 1; index <= lastLogIndex; index++ {
		var entry pb.Log
		if err = logs.GetLog(index, &entry); err != nil {
			return fmt.Errorf("failed to get log at index %d: %v", index, err)
		}
		if entry.Type == pb.LogType_COMMAND {
			_ = fsm.Apply(&entry)
		}

		lastIndex = entry.Index
		lastTerm = entry.Term
	}

	if lastIndex != lastLogIndex {
		klog.Fatalf(fmt.Sprintf("lastIndex:%d should be equal to lastLogIndex:%d", lastIndex, lastLogIndex))
	}

	// snapshot fsm
	snapshot, err := fsm.Snapshot()
	if err != nil {
		return fmt.Errorf("failed to snapshot FSM: %v", err)
	}
	sink, err := snaps.Create(lastIndex, lastTerm, configuration, 1)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}
	if err = snapshot.Persist(sink); err != nil {
		return fmt.Errorf("failed to persist snapshot: %v", err)
	}
	if err = sink.Close(); err != nil {
		return fmt.Errorf("failed to finalize snapshot: %v", err)
	}
	// compact logs
	firstLogIndex, err := logs.FirstIndex()
	if err != nil {
		return fmt.Errorf("failed to get first log index: %v", err)
	}
	if err := logs.DeleteRange(firstLogIndex, lastLogIndex); err != nil {
		return fmt.Errorf("log compaction failed: %v", err)
	}

	return nil
}

// HasExistingState returns true if the server has any existing state (logs,
// knowledge of a current term, or any snapshotStore).
func HasExistingState(logs LogStore, stable StableStore, snaps SnapshotStore) (bool, error) {
	// Make sure we don't have a current term.
	currentTerm, err := stable.GetUint64(keyCurrentTerm)
	if err == nil {
		if currentTerm > 0 {
			return true, nil
		}
	} else {
		if err.Error() != "not found" {
			return false, fmt.Errorf("failed to read current term: %v", err)
		}
	}

	// Make sure we have an empty log.
	lastIndex, err := logs.LastIndex()
	if err != nil {
		return false, fmt.Errorf("failed to get last log index: %v", err)
	}
	if lastIndex > 0 {
		return true, nil
	}

	// Make sure we have no snapshotStore
	snapshots, err := snaps.List()
	if err != nil {
		return false, fmt.Errorf("failed to list snapshotStore: %v", err)
	}
	if len(snapshots) > 0 {
		return true, nil
	}

	return false, nil
}
