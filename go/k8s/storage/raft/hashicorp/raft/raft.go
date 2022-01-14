package raft

import (
	"container/list"
	"fmt"
	"k8s.io/klog/v2"
	"sync"
	"sync/atomic"
	"time"
)

var (
	keyCurrentTerm  = []byte("CurrentTerm")
	keyLastVoteTerm = []byte("LastVoteTerm")
	keyLastVoteCand = []byte("LastVoteCand")
)

// leaderState is state that is used while we are a leader.
type leaderState struct {
	leadershipTransferInProgress int32 // indicates that a leadership transfer is in progress.
	commitCh                     chan struct{}
	//commitment                   *commitment
	inflight *list.List // list of logFuture in log index order
	//replState                    map[ServerID]*followerReplication
	//notify                       map[*verifyFuture]struct{}
	stepDown chan struct{}
}

// Raft implements a Raft node.
type Raft struct {
	raftState

	// applyCh is used to async send logs to the main thread to
	// be committed and applied to the FSM.
	applyCh chan *logFuture

	// conf stores the current configuration to use. This is the most recent one
	// provided. All reads of config values should use the config() helper method
	// to read this safely.
	conf atomic.Value

	// confReloadMu ensures that only one thread can reload config at once since
	// we need to read-modify-write the atomic. It is NOT necessary to hold this
	// for any other operation e.g. reading config using config().
	confReloadMu sync.Mutex

	// FSM is the client state machine to apply commands to
	fsm FSM

	// fsmMutateCh is used to send state-changing updates to the FSM. This
	// receives pointers to commitTuple structures when applying logs or
	// pointers to restoreFuture structures when restoring a snapshot. We
	// need control over the order of these operations when doing user
	// restores so that we finish applying any old log applies before we
	// take a user snapshot on the leader, otherwise we might restore the
	// snapshot and apply old logs to it that were in the pipe.
	fsmMutateCh chan interface{}

	// fsmSnapshotCh is used to trigger a new snapshot being taken
	//fsmSnapshotCh chan *reqSnapshotFuture

	// lastContact is the last time we had contact from the
	// leader node. This can be used to gauge staleness.
	lastContact     time.Time
	lastContactLock sync.RWMutex

	// Leader is the current cluster leader
	leader     ServerAddress
	leaderLock sync.RWMutex

	// leaderCh is used to notify of leadership changes
	leaderCh chan bool

	// leaderState used only while state is leader
	leaderState leaderState

	// candidateFromLeadershipTransfer is used to indicate that this server became
	// candidate because the leader tries to transfer leadership. This flag is
	// used in RequestVoteRequest to express that a leadership transfer is going
	// on.
	candidateFromLeadershipTransfer bool

	// Stores our local server ID, used to avoid sending RPCs to ourself
	localID ServerID

	// Stores our local addr
	localAddr ServerAddress

	// LogStore provides durable storage for logs
	logs LogStore
	// stable is a StableStore implementation for durable state
	// It provides stable storage for many fields in raftState
	stable StableStore

	// Used to request the leader to make configuration changes.
	configurationChangeCh chan *configurationChangeFuture
	// Tracks the latest configuration and latest committed configuration from
	// the log/snapshot.
	configurations configurations
	// Holds a copy of the latest configuration which can be read
	// independent of main loop.
	latestConfiguration atomic.Value
	// configurationsCh is used to get the configuration data safely from
	// outside of the main thread.
	configurationsCh chan *configurationsFuture

	// RPC chan comes from the transport layer
	rpcCh <-chan RPC

	// Shutdown channel to exit, protected to prevent concurrent exits
	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex

	// snapshots is used to store and retrieve snapshots
	snapshots SnapshotStore

	// userSnapshotCh is used for user-triggered snapshots
	//userSnapshotCh chan *userSnapshotFuture

	// userRestoreCh is used for user-triggered restores of external
	// snapshots
	//userRestoreCh chan *userRestoreFuture

	// The transport layer we use
	trans Transport

	// verifyCh is used to async send verify futures to the main thread
	// to verify we are still the leader
	verifyCh chan *verifyFuture

	// bootstrapCh is used to attempt an initial bootstrap from outside of
	// the main thread.
	bootstrapCh chan *bootstrapFuture

	// List of observers and the mutex that protects them. The observers list
	// is indexed by an artificial ID which is used for deregistration.
	observersLock sync.RWMutex
	//observers     map[uint64]*Observer

	// leadershipTransferCh is used to start a leadership transfer from outside of
	// the main thread.
	//leadershipTransferCh chan *leadershipTransferFuture
}

// NewRaft is used to construct a new Raft node. It takes a configuration, as well
// as implementations of various interfaces that are required. If we have any
// old state, such as snapshots, logs, peers, etc, all those will be restored
// when creating the Raft node.
func NewRaft(conf *Config, fsm FSM, logs LogStore, stable StableStore, snaps SnapshotStore, transport Transport) (*Raft, error) {
	// Validate the configuration.
	if err := ValidateConfig(conf); err != nil {
		return nil, err
	}

	// Try to restore the current term.
	currentTerm, err := stable.GetUint64(keyCurrentTerm)
	if err != nil && err.Error() != "not found" {
		return nil, fmt.Errorf("failed to load current term: %v", err)
	}

	// Read the index of the last log entry.
	lastIndex, err := logs.LastIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to find last log: %v", err)
	}

	// Get the last log entry.
	var lastLog Log
	if lastIndex > 0 {
		if err = logs.GetLog(lastIndex, &lastLog); err != nil {
			return nil, fmt.Errorf("failed to get last log at index %d: %v", lastIndex, err)
		}
	}

	// Buffer applyCh to MaxAppendEntries if the option is enabled
	applyCh := make(chan *logFuture)
	if conf.BatchApplyCh {
		applyCh = make(chan *logFuture, conf.MaxAppendEntries)
	}

	localAddr := transport.LocalAddr()
	localID := conf.LocalID
	r := &Raft{
		applyCh:     applyCh,
		fsm:         fsm,
		fsmMutateCh: make(chan interface{}, 128),
		//fsmSnapshotCh:         make(chan *reqSnapshotFuture),
		leaderCh:              make(chan bool, 1),
		localID:               localID,
		localAddr:             localAddr,
		logs:                  logs,
		configurationChangeCh: make(chan *configurationChangeFuture),
		configurations:        configurations{},
		rpcCh:                 transport.Consumer(),
		snapshots:             snaps,
		//userSnapshotCh:        make(chan *userSnapshotFuture),
		//userRestoreCh:         make(chan *userRestoreFuture),
		shutdownCh:       make(chan struct{}),
		stable:           stable,
		trans:            transport,
		verifyCh:         make(chan *verifyFuture, 64),
		configurationsCh: make(chan *configurationsFuture, 8),
		bootstrapCh:      make(chan *bootstrapFuture),
		//observers:             make(map[uint64]*Observer),
		//leadershipTransferCh:  make(chan *leadershipTransferFuture, 1),
	}

	r.conf.Store(*conf)

	// Initialize as a follower.
	r.setState(Follower)
	// Restore the current term and the last log.
	r.setCurrentTerm(currentTerm)
	r.setLastLog(lastLog.Index, lastLog.Term)

	// Attempt to restore a snapshot if there are any.
	if err := r.restoreSnapshot(); err != nil {
		return nil, err
	}

	// Scan through the log for any configuration change entries in [snapshotIndex + 1, lastLogIndex]
	snapshotIndex, _ := r.getLastSnapshot()
	for index := snapshotIndex + 1; index <= lastLog.Index; index++ {
		var entry Log
		if err := r.logs.GetLog(index, &entry); err != nil {
			klog.Error(fmt.Sprintf("failed to get log for index:%d err:%v", index, err))
			panic(err)
		}
		if err := r.processConfigurationLogEntry(&entry); err != nil {
			return nil, err
		}
	}

	klog.Infof(fmt.Sprintf("initial configuration for latestIndex:%d in servers:%v",
		r.configurations.latestIndex, r.configurations.latest.Servers))

	// Setup a heartbeat fast-path to avoid head-of-line
	// blocking where possible. It MUST be safe for this
	// to be called concurrently with a blocking RPC.
	transport.SetHeartbeatHandler(r.processHeartbeat)

	if conf.skipStartup {
		return r, nil
	}

	go r.run()
	go r.runFSM()
	go r.runSnapshots()

	return r, nil
}

func (r *Raft) run() {
	for {
		switch r.getState() {
		case Follower:
			r.runFollower()
		}
	}
}

func (r *Raft) runFollower() {
	for r.getState() == Follower {
		select {}
	}
}

func (r *Raft) setState(state RaftState) {
	oldState := r.raftState.getState()
	r.raftState.setState(state)
	if oldState != state {
		klog.Infof(fmt.Sprintf("swich raft state from %s to %s", oldState, state))
	}
}

// setCurrentTerm is used to set the current term in a durable manner.
func (r *Raft) setCurrentTerm(t uint64) {
	// Persist to disk first
	if err := r.stable.SetUint64(keyCurrentTerm, t); err != nil {
		panic(fmt.Errorf("failed to save current term: %v", err))
	}
	r.raftState.setCurrentTerm(t)
}

// processConfigurationLogEntry takes a log entry and updates the latest
// configuration if the entry results in a new configuration. This must only be
// called from the main thread, or from NewRaft() before any threads have begun.
func (r *Raft) processConfigurationLogEntry(entry *Log) error {
	switch entry.Type {
	case LogConfiguration:
		r.setCommittedConfiguration(r.configurations.latest, r.configurations.latestIndex)
		r.setLatestConfiguration(DecodeConfiguration(entry.Data), entry.Index)

	}

	return nil
}

// setCommittedConfiguration stores the committed configuration.
func (r *Raft) setCommittedConfiguration(c Configuration, i uint64) {
	r.configurations.committed = c
	r.configurations.committedIndex = i
}

// setLatestConfiguration stores the latest configuration and updates a copy of it.
func (r *Raft) setLatestConfiguration(c Configuration, i uint64) {
	r.configurations.latest = c
	r.configurations.latestIndex = i
	r.latestConfiguration.Store(c.Clone())
}

// restoreSnapshot attempts to restore the latest snapshots, and fails if none
// of them can be restored. This is called at initialization time, and is
// completely unsafe to call at any other time.
func (r *Raft) restoreSnapshot() error {
	return nil
}

// processHeartbeat is a special handler used just for heartbeat requests
// so that they can be fast-pathed if a transport supports it. This must only
// be called from the main thread.
func (r *Raft) processHeartbeat(rpc RPC) {
	// Check if we are shutdown, just ignore the RPC
	select {
	case <-r.shutdownCh:
		return
	default:
	}

	// Ensure we are only handling a heartbeat
	switch cmd := rpc.Command.(type) {
	case *AppendEntriesRequest:
		r.appendEntries(rpc, cmd)
	default:
		klog.Error(fmt.Sprintf("expected heartbeat, got command: %+v", rpc.Command))
		rpc.Respond(nil, fmt.Errorf("unexpected command"))
	}
}

// appendEntries is invoked when we get an append entries RPC call. This must
// only be called from the main thread.
func (r *Raft) appendEntries(rpc RPC, cmd *AppendEntriesRequest) {

}
