package raft

import (
	"bytes"
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
	commitment                   *commitment
	inflight                     *list.List // list of logFuture in log index order
	replState                    map[ServerID]*followerReplication
	notify                       map[*verifyFuture]struct{}
	stepDown                     chan struct{}
}

// Raft implements a Raft node.
type Raft struct {
	// bootstrapCh is used to attempt an initial bootstrap from outside of
	// the main thread.
	bootstrapCh chan *bootstrapFuture

	raftState
	// Stores our local server ID, used to avoid sending RPCs to ourself
	localID ServerID
	// Stores our local addr
	localAddr ServerAddress
	// Leader is the current cluster leader
	leader     ServerAddress
	leaderLock sync.RWMutex
	// leaderCh is used to notify of leadership changes
	leaderCh chan bool
	// leaderState used only while state is leader
	leaderState leaderState

	// confReloadMu ensures that only one thread can reload config at once since
	// we need to read-modify-write the atomic. It is NOT necessary to hold this
	// for any other operation e.g. reading config using config().
	confReloadMu sync.Mutex
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
	// conf stores the current configuration to use. This is the most recent one
	// provided. All reads of config values should use the config() helper method
	// to read this safely.
	conf atomic.Value

	// applyCh is used to async send logs to the main thread to
	// be committed and applied to the FSM.
	applyCh chan *logFuture
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
	// snapshots is used to store and retrieve snapshots
	snapshots SnapshotStore
	// fsmSnapshotCh is used to trigger a new snapshot being taken
	fsmSnapshotCh chan *reqSnapshotFuture
	// userSnapshotCh is used for user-triggered snapshots
	userSnapshotCh chan *userSnapshotFuture
	// userRestoreCh is used for user-triggered restores of external
	// snapshots
	userRestoreCh chan *userRestoreFuture

	// lastContact is the last time we had contact from the
	// leader node. This can be used to gauge staleness.
	lastContact     time.Time
	lastContactLock sync.RWMutex

	// candidateFromLeadershipTransfer is used to indicate that this server became
	// candidate because the leader tries to transfer leadership. This flag is
	// used in RequestVoteRequest to express that a leadership transfer is going
	// on.
	candidateFromLeadershipTransfer bool
	// leadershipTransferCh is used to start a leadership transfer from outside of
	// the main thread.
	leadershipTransferCh chan *leadershipTransferFuture

	// LogStore provides durable storage for logs
	logs LogStore
	// stable is a StableStore implementation for durable state
	// It provides stable storage for many fields in raftState
	stable StableStore

	// RPC chan comes from the transport layer
	rpcCh <-chan RPC
	// The transport layer we use
	transport Transport

	// verifyCh is used to async send verify futures to the main thread
	// to verify we are still the leader
	verifyCh chan *verifyFuture

	// List of observers and the mutex that protects them. The observers list
	// is indexed by an artificial ID which is used for deregistration.
	observersLock sync.RWMutex
	observers     map[uint64]*Observer

	// Shutdown channel to exit, protected to prevent concurrent exits
	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
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
		bootstrapCh: make(chan *bootstrapFuture),

		localID:               localID,
		localAddr:             localAddr,
		configurationChangeCh: make(chan *configurationChangeFuture),
		configurations:        configurations{},
		configurationsCh:      make(chan *configurationsFuture, 8),
		leaderCh:              make(chan bool, 1),

		applyCh:        applyCh,
		fsm:            fsm,
		fsmMutateCh:    make(chan interface{}, 128),
		fsmSnapshotCh:  make(chan *reqSnapshotFuture),
		snapshots:      snaps,
		userSnapshotCh: make(chan *userSnapshotFuture),
		userRestoreCh:  make(chan *userRestoreFuture),

		logs:   logs,
		stable: stable,

		rpcCh:                transport.Consumer(),
		transport:            transport,
		verifyCh:             make(chan *verifyFuture, 64),
		leadershipTransferCh: make(chan *leadershipTransferFuture, 1),

		observers:  make(map[uint64]*Observer),
		shutdownCh: make(chan struct{}),
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

// BootstrapCluster is equivalent to non-member BootstrapCluster but can be
// called on an un-bootstrapped Raft instance after it has been created. This
// should only be called at the beginning of time for the cluster with an
// identical configuration listing all Voter servers. There is no need to
// bootstrap Nonvoter and Staging servers.
//
// A cluster can only be bootstrapped once from a single participating Voter
// server. Any further attempts to bootstrap will return an error that can be
// safely ignored.
//
// One sane approach is to bootstrap a single server with a configuration
// listing just itself as a Voter, then invoke AddVoter() on it to add other
// servers to the cluster.
func (r *Raft) BootstrapCluster(configuration Configuration) Future {
	bootstrapReq := &bootstrapFuture{}
	bootstrapReq.init()
	bootstrapReq.configuration = configuration
	select {
	case <-r.shutdownCh:
		return errorFuture{ErrRaftShutdown}
	case r.bootstrapCh <- bootstrapReq:
		return bootstrapReq
	}
}

func (r *Raft) run() {
	for {
		// Check if we are doing a shutdown
		select {
		case <-r.shutdownCh:
			// Clear the leader to prevent forwarding
			r.setLeader("")
			return
		default:
		}

		switch r.getState() {
		case Follower:
			r.runFollower()
		case Candidate:
			r.runCandidate()
		case Leader:
			r.runLeader()
		}
	}
}

func (r *Raft) runFollower() {
	klog.Infof(fmt.Sprintf("%s/%s entering follower state in cluster for leader:%s", r.localID, r.localAddr, r.Leader()))

	heartbeatTimer := randomTimeout(r.config().HeartbeatTimeout)
	for r.getState() == Follower {
		select {
		case b := <-r.bootstrapCh:
			b.respond(r.liveBootstrap(b.configuration))

		case <-heartbeatTimer: // 每 [1s, 2s] 一次心跳检查是否要心跳
			// Restart the heartbeat timer
			heartbeatTimeout := r.config().HeartbeatTimeout
			heartbeatTimer = randomTimeout(heartbeatTimeout) // [1s, 2s]

			// INFO: 性能提高: 这里使用 lastContact，如果是正常的 log replicate，也会修改 lastContact，这样在 heartbeatTimeout 内不需要再去心跳检查
			//  本来担心网络抖动会导致几次心跳没成功，会发起 leader election；但是每 HeartbeatTimeout / 10 leader 发起一次心跳，如果
			//  10次心跳都没成功，就必然 ElectionTimeout，则可以发起选举, @see https://github.com/hashicorp/raft/blob/v1.3.3/replication.go#L389-L394
			lastContact := r.LastContact()
			if time.Now().Sub(lastContact) < heartbeatTimeout {
				continue
			}

			// Heartbeat failed! Transition to the candidate state
			lastLeader := r.Leader()
			r.setLeader("")
			if r.configurations.latestIndex == 0 { // INFO: 如果没有 BootstrapCluster 或者 BootstrapCluster 慢于 heartbeatTimeout，提示warning
				klog.Warningf("no known peers because of not BootstrapCluster, aborting election")
			} else if r.configurations.latestIndex == r.configurations.committedIndex &&
				!hasVote(r.configurations.latest, r.localID) {
				klog.Warningf("not part of stable configuration, aborting election")
			} else {
				if hasVote(r.configurations.latest, r.localID) {
					klog.Warningf(fmt.Sprintf("%s/%s heartbeat timeout reached, starting election lastLeader:%s", r.localID, r.localAddr, lastLeader))
					r.setState(Candidate)
					return
				} else {
					klog.Warningf("heartbeat timeout reached, not part of a stable configuration or a non-voter, not triggering a leader election")
				}
			}

		case rpc := <-r.rpcCh:
			r.processRPC(rpc)

		case <-r.shutdownCh:
			return
		}
	}
}

// liveBootstrap attempts to seed an initial configuration for the cluster. See
// the Raft object's member BootstrapCluster for more details. This must only be
// called on the main thread, and only makes sense in the follower state.
func (r *Raft) liveBootstrap(configuration Configuration) error {
	// Use the pre-init API to make the static updates.
	cfg := r.config()
	err := BootstrapCluster(&cfg, r.logs, r.stable, r.snapshots, configuration)
	if err != nil {
		return err
	}

	// Make the configuration live.
	var entry Log
	if err := r.logs.GetLog(1, &entry); err != nil {
		panic(err)
	}
	r.setCurrentTerm(1)
	r.setLastLog(entry.Index, entry.Term)
	return r.processConfigurationLogEntry(&entry)
}

type voteResult struct {
	RequestVoteResponse
	voterID      ServerID
	voterAddress ServerAddress
}

// runCandidate runs the FSM for a candidate.
func (r *Raft) runCandidate() {
	klog.Infof(fmt.Sprintf("%s/%s entering candidate state in term:%d for leader:%s", r.localID, r.localAddr, r.getCurrentTerm()+1, r.Leader()))

	// Start vote for local and peers, and set a timeout
	voteCh := r.startElection()

	// Make sure the leadership transfer flag is reset after each run. Having this
	// flag will set the field LeadershipTransfer in a RequestVoteRequst to true,
	// which will make other servers vote even though they have a leader already.
	// It is important to reset that flag, because this priviledge could be abused
	// otherwise.
	defer func() { r.candidateFromLeadershipTransfer = false }()

	electionTimer := randomTimeout(r.config().ElectionTimeout) // [10s, 20s]
	grantedVotes := 0
	votesNeeded := r.quorumSize()
	klog.Infof(fmt.Sprintf("need %d votes at least", votesNeeded))
	for r.getState() == Candidate {
		select {
		case vote := <-voteCh:
			// Check if the term is greater than ours, bail
			if vote.Term > r.getCurrentTerm() { // INFO: @see raft paper 3.4
				klog.Warningf("newer term discovered, fallback to follower")
				r.setState(Follower)
				r.setCurrentTerm(vote.Term)
				return
			}

			// Check if the vote is granted
			if vote.Granted {
				grantedVotes++
				klog.Infof(fmt.Sprintf("vote granted from %s/%s to %s/%s at term:%d and votes is %d/%d now",
					vote.voterID, vote.voterAddress, r.localID, r.localAddr, vote.Term, grantedVotes, r.totalVoteSize()))
			}
			// Check if we've become the leader
			if grantedVotes >= votesNeeded {
				klog.Infof(fmt.Sprintf("%s/%s election win %d votes", r.localID, r.localAddr, grantedVotes))
				r.setState(Leader)
				r.setLeader(r.localAddr)
				return
			}

		case <-electionTimer:
			// Election failed! Restart the election. We simply return,
			// which will kick us back into runCandidate
			klog.Warningf("Election timeout reached, restarting election")
			return

		case rpc := <-r.rpcCh:
			r.processRPC(rpc)

		case <-r.shutdownCh:
			return
		}
	}
}

func (r *Raft) totalVoteSize() int {
	voters := 0
	for _, server := range r.configurations.latest.Servers {
		if server.Suffrage == Voter {
			voters++
		}
	}
	return voters
}

// quorumSize is used to return the quorum size. This must only be called on
// the main thread.
func (r *Raft) quorumSize() int {
	return r.totalVoteSize()/2 + 1
}

// startElection is used to send a RequestVote RPC to all peers, and vote for
// ourself. This has the side affecting of incrementing the current term. The
// response channel returned is used to wait for all the responses (including a
// vote for ourself). This must only be called from the main thread.
func (r *Raft) startElection() <-chan *voteResult {
	// Create a response channel
	respCh := make(chan *voteResult, len(r.configurations.latest.Servers))

	// Increase the term
	r.setCurrentTerm(r.getCurrentTerm() + 1)

	// Construct the request
	lastIdx, lastTerm := r.getLastEntry()
	req := &RequestVoteRequest{
		Term:               r.getCurrentTerm(),
		Candidate:          r.transport.EncodePeer(r.localID, r.localAddr),
		LastLogIndex:       lastIdx,
		LastLogTerm:        lastTerm,
		LeadershipTransfer: r.candidateFromLeadershipTransfer,
	}

	// Construct a function to ask for a vote
	askPeer := func(peer Server) {
		go func() {
			resp := &voteResult{
				voterID:      peer.ID,
				voterAddress: peer.Address,
			}
			err := r.transport.RequestVote(peer.ID, peer.Address, req, &resp.RequestVoteResponse)
			if err != nil {
				klog.Errorf(fmt.Sprintf("failed to make requestVote RPC for target:%s/%s err:%v", peer.ID, peer.Address, err))
				resp.Term = req.Term
				resp.Granted = false
			}
			respCh <- resp
		}()
	}

	// For each peer, request a vote
	for _, server := range r.configurations.latest.Servers {
		if server.Suffrage == Voter {
			if server.ID == r.localID {
				// Persist a vote for ourselves
				if err := r.persistVote(req.Term, req.Candidate); err != nil {
					klog.Error(fmt.Sprintf("failed to persist vote err:%v", err))
					return nil
				}
				// Include our own vote
				respCh <- &voteResult{
					RequestVoteResponse: RequestVoteResponse{
						Term:    req.Term,
						Granted: true,
					},
					voterID:      r.localID,
					voterAddress: r.localAddr,
				}
			} else {
				askPeer(server)
			}
		}
	}

	return respCh
}

// persistVote is used to persist our vote for safety.
func (r *Raft) persistVote(term uint64, candidate []byte) error {
	if err := r.stable.SetUint64(keyLastVoteTerm, term); err != nil {
		return err
	}
	if err := r.stable.Set(keyLastVoteCand, candidate); err != nil {
		return err
	}
	return nil
}

// runLeader runs the FSM for a leader. Do the setup here and drop into
// the leaderLoop for the hot loop.
func (r *Raft) runLeader() {
	klog.Infof(fmt.Sprintf("%s/%s entering leader state in term:%d for leader:%s", r.localID, r.localAddr, r.getCurrentTerm(), r.Leader()))

	// setup leader state. This is only supposed to be accessed within the leaderloop.
	r.setupLeaderState()

	// Start a replication routine for each peer
	r.startReplication()

	// stepDown is used to track if there is an inflight log that
	// would cause us to lose leadership (specifically a RemovePeer of
	// ourselves). If this is the case, we must not allow any logs to
	// be processed in parallel, otherwise we are basing commit on
	// only a single peer (ourself) and replicating to an undefined set
	// of peers.
	stepDown := false

	for r.getState() == Leader {
		select {
		case rpc := <-r.rpcCh:
			r.processRPC(rpc)

		case newLog := <-r.applyCh:
			if r.getLeadershipTransferInProgress() {
				klog.Warningf(ErrLeadershipTransferInProgress.Error())
				newLog.respond(ErrLeadershipTransferInProgress)
				continue
			}

			// Group commit, gather all the ready commits
			ready := []*logFuture{newLog}
		GroupCommitLoop:
			for i := 0; i < r.config().MaxAppendEntries; i++ {
				select {
				case log := <-r.applyCh:
					ready = append(ready, log)
				default:
					break GroupCommitLoop
				}
			}

			// Dispatch the logs
			if stepDown {
				// we're in the process of stepping down as leader, don't process anything new
				for i := range ready {
					ready[i].respond(ErrNotLeader)
				}
			} else {
				r.dispatchLogs(ready)
			}

		// leader commit logs and apply to fsm
		case <-r.leaderState.commitCh:
			// Process the newly committed entries
			oldCommitIndex := r.getCommitIndex()
			commitIndex := r.leaderState.commitment.getCommitIndex()
			r.setCommitIndex(commitIndex)

			// New configration has been committed, set it as the committed value.
			if r.configurations.latestIndex > oldCommitIndex &&
				r.configurations.latestIndex <= commitIndex {
				r.setCommittedConfiguration(r.configurations.latest, r.configurations.latestIndex)
				if !hasVote(r.configurations.committed, r.localID) {
					stepDown = true
				}
			}

			// Pull all inflight logs that are committed off the queue.
			var groupReady []*list.Element
			var groupFutures = make(map[uint64]*logFuture)
			var lastIdxInGroup uint64
			for e := r.leaderState.inflight.Front(); e != nil; e = e.Next() {
				commitLog := e.Value.(*logFuture)
				idx := commitLog.log.Index
				if idx > commitIndex {
					// Don't go past the committed index
					break
				}

				groupReady = append(groupReady, e)
				groupFutures[idx] = commitLog
				lastIdxInGroup = idx
			}
			// Process the group
			if len(groupReady) != 0 {
				r.applyLogsToFSM(lastIdxInGroup, groupFutures)

				for _, e := range groupReady {
					r.leaderState.inflight.Remove(e)
				}
			}

			if stepDown { // leader 降级
				if r.config().ShutdownOnRemove {
					klog.Info("removed ourself, shutting down")
					r.Shutdown()
				} else {
					klog.Info("removed ourself, transitioning to follower")
					r.setState(Follower)
				}
			}

		// INFO: @see handleStaleTerm()
		case <-r.leaderState.stepDown:
			r.setState(Follower)

		case <-r.shutdownCh:
			return
		}
	}
}

// dispatchLog is called on the leader to push a log to disk, mark it
// as inflight and begin replication of it.
func (r *Raft) dispatchLogs(applyLogs []*logFuture) {
	now := time.Now()
	term := r.getCurrentTerm()
	lastIndex := r.getLastIndex()
	logs := make([]*Log, len(applyLogs))
	for idx, applyLog := range applyLogs {
		applyLog.dispatch = now
		lastIndex++
		applyLog.log.Index = lastIndex
		applyLog.log.Term = term
		applyLog.log.AppendedAt = now
		logs[idx] = &applyLog.log
		r.leaderState.inflight.PushBack(applyLog)
	}

	// push logs to disk(boltdb)
	if err := r.logs.StoreLogs(logs); err != nil {
		klog.Errorf(fmt.Sprintf("failed to store logs to disk err:%v", err))
		for _, applyLog := range applyLogs {
			applyLog.respond(err)
		}
		r.setState(Follower)
		return
	}

	r.leaderState.commitment.match(r.localID, lastIndex)
	// Update the last log since it's on disk now
	r.setLastLog(lastIndex, term)

	// Notify the replicators of the new log
	for _, replication := range r.leaderState.replState {
		select {
		case replication.triggerCh <- struct{}{}:
		default:
		}
	}
}

func (r *Raft) setupLeaderState() {
	r.leaderState.commitCh = make(chan struct{}, 1)
	r.leaderState.commitment = newCommitment(r.leaderState.commitCh,
		r.configurations.latest,
		r.getLastIndex()+1 /* first index that may be committed in this term */)
	r.leaderState.inflight = list.New()
	r.leaderState.replState = make(map[ServerID]*followerReplication)
	r.leaderState.notify = make(map[*verifyFuture]struct{})
	r.leaderState.stepDown = make(chan struct{}, 1)
}

func (r *Raft) setLeadershipTransferInProgress(v bool) {
	if v {
		atomic.StoreInt32(&r.leaderState.leadershipTransferInProgress, 1)
	} else {
		atomic.StoreInt32(&r.leaderState.leadershipTransferInProgress, 0)
	}
}

func (r *Raft) getLeadershipTransferInProgress() bool {
	return atomic.LoadInt32(&r.leaderState.leadershipTransferInProgress) == 1
}

// startReplication will set up state and start asynchronous replication to
// new peers, and stop replication to removed peers. Before removing a peer,
// it'll instruct the replication routines to try to replicate to the current
// index. This must only be called from the main thread.
func (r *Raft) startReplication() {
	inConfig := make(map[ServerID]bool, len(r.configurations.latest.Servers))
	lastIdx := r.getLastIndex()

	// Start replication goroutines that need starting
	for _, server := range r.configurations.latest.Servers {
		if server.ID == r.localID {
			continue
		}

		inConfig[server.ID] = true

		replication, ok := r.leaderState.replState[server.ID]
		if !ok {
			klog.Infof(fmt.Sprintf("leader:%s/%s start a new follower replication for follower:%s/%s",
				r.localID, r.localAddr, server.ID, server.Address))

			replication = &followerReplication{
				currentTerm:         r.getCurrentTerm(),
				nextIndex:           lastIdx + 1,
				peer:                server,
				commitment:          r.leaderState.commitment,
				stepDown:            r.leaderState.stepDown,
				lastContact:         time.Now(),
				stopCh:              make(chan uint64, 1),
				triggerCh:           make(chan struct{}, 1), // buffer channel
				triggerDeferErrorCh: make(chan *deferError, 1),
				notify:              make(map[*verifyFuture]struct{}),
				notifyCh:            make(chan struct{}, 1),
			}
			r.leaderState.replState[server.ID] = replication
			go r.replicate(replication)

			select {
			case replication.triggerCh <- struct{}{}:
			default:
			}

			r.observe(PeerObservation{Peer: server, Removed: false})
		} else {
			replication.peerLock.RLock()
			peer := replication.peer
			replication.peerLock.RUnlock()
			if peer.Address != server.Address {
				klog.Infof(fmt.Sprintf("live change the peer address for %s/%s", server.ID, server.Address))
				replication.peerLock.Lock()
				peer.Address = server.Address
				replication.peerLock.Unlock()
			}
		}
	}

	// Stop replication goroutines that need stopping
	// Before removing a peer, it'll instruct the replication routines to try to replicate to the current index.
	for serverID, repl := range r.leaderState.replState {
		if inConfig[serverID] {
			continue
		}

		// Replicate up to lastIdx and stop
		klog.Infof(fmt.Sprintf("removed peer:%s/%s from leader:%s/%s peers, and stopping replication until up to lastIndex:%d",
			repl.peer.ID, repl.peer.Address, r.localID, r.localAddr, lastIdx))
		repl.stopCh <- lastIdx
		close(repl.stopCh)
		delete(r.leaderState.replState, serverID)
		r.observe(PeerObservation{Peer: repl.peer, Removed: true})
	}
}

func (r *Raft) config() Config {
	return r.conf.Load().(Config)
}

func (r *Raft) setState(state RaftState) {
	oldState := r.raftState.getState()
	r.raftState.setState(state)
	if oldState != state {
		klog.Infof(fmt.Sprintf("switch raft state from %s to %s for %s/%s", oldState, state, r.localID, r.localAddr))
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

// setLeader is used to modify the current leader of the cluster
func (r *Raft) setLeader(leader ServerAddress) {
	r.leaderLock.Lock()
	oldLeader := r.leader
	r.leader = leader
	r.leaderLock.Unlock()
	if oldLeader != leader {
		r.observe(LeaderObservation{Leader: leader})
	}
}

// Leader is used to return the current leader of the cluster.
// It may return empty string if there is no current leader
// or the leader is unknown.
func (r *Raft) Leader() ServerAddress {
	r.leaderLock.RLock()
	leader := r.leader
	r.leaderLock.RUnlock()
	return leader
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

// processRPC is called to handle an incoming RPC request. This must only be
// called from the main thread.
func (r *Raft) processRPC(rpc RPC) {
	switch cmd := rpc.Command.(type) {
	case *AppendEntriesRequest:
		r.appendEntries(rpc, cmd)
	case *RequestVoteRequest:
		r.requestVote(rpc, cmd)
	case *InstallSnapshotRequest:
		r.installSnapshot(rpc, cmd)
	case *TimeoutNowRequest:
		r.timeoutNow(rpc, cmd)
	default:
		klog.Error(fmt.Sprintf("got unexpected command: %+v", rpc.Command))
		rpc.Respond(nil, fmt.Errorf("unexpected command"))
	}
}

/*candidate 之后，获取 grant vote 必要条件：
* 如果已经有 leader 且 leader != candidate，则 reject vote；
* 如果 term 小于 follower term，则 reject term；
* 同 term 时不能给不同 candidate 投票多次，但是可以给同一个 candidate 投票多次，否则 reject term；
* 如果 follower lastLogTerm > candidate lastLogTerm，则 reject term；
* 如果 follower lastLogTerm == candidate lastLogTerm，但是 follower lastLogIndex == candidate lastLogIndex 则 reject vote；
 */
func (r *Raft) requestVote(rpc RPC, req *RequestVoteRequest) {
	resp := &RequestVoteResponse{
		Term:    r.getCurrentTerm(),
		Granted: false,
	}
	var rpcErr error
	defer func() {
		rpc.Respond(resp, rpcErr)
	}()

	// Check if we have an existing leader [who's not the candidate] and also
	// check the LeadershipTransfer flag is set. Usually votes are rejected if
	// there is a known leader. But if the leader initiated a leadership transfer,
	// vote!
	candidate := r.transport.DecodePeer(req.Candidate)
	if leader := r.Leader(); leader != "" && leader != candidate && !req.LeadershipTransfer {
		klog.Warningf(fmt.Sprintf("rejecting vote request since we have a leader:%s from candidate:%s", leader, candidate))
		return
	}

	// Ignore an older term
	if req.Term < r.getCurrentTerm() {
		return
	}

	// Increase the term if we see a newer one
	if req.Term > r.getCurrentTerm() {
		// Ensure transition to follower
		klog.Warningf(fmt.Sprintf("%s/%s received a requestVote from %s with a newer term:%d",
			r.localID, r.localAddr, candidate, req.Term))
		r.setState(Follower)
		r.setCurrentTerm(req.Term)
		resp.Term = req.Term
	}

	// Check if we've voted in this election before
	lastVoteTerm, err := r.stable.GetUint64(keyLastVoteTerm)
	if err != nil && err.Error() != "not found" {
		klog.Errorf(fmt.Sprintf("failed to get last vote term err:%v", err))
		return
	}
	lastVoteCandBytes, err := r.stable.Get(keyLastVoteCand)
	if err != nil && err.Error() != "not found" {
		klog.Errorf(fmt.Sprintf("failed to get last vote candidate err:%v", err))
		return
	}
	if lastVoteTerm == req.Term && lastVoteCandBytes != nil {
		klog.Infof(fmt.Sprintf("duplicate requestVote at term:%d", req.Term))
		if bytes.Compare(lastVoteCandBytes, req.Candidate) == 0 {
			klog.Warningf(fmt.Sprintf("duplicate requestVote from candidate:%s at term:%d", candidate, req.Term))
			resp.Granted = true
		}
		return
	}

	// Reject if their term is older
	lastIdx, lastTerm := r.getLastEntry()
	if lastTerm > req.LastLogTerm {
		klog.Warningf(fmt.Sprintf("rejecting vote request since our lastTerm:%d is greater than candidate:%s term:%d",
			lastTerm, candidate, req.LastLogTerm))
		return
	}
	if lastTerm == req.LastLogTerm && lastIdx > req.LastLogIndex {
		klog.Warningf(fmt.Sprintf("rejecting vote request since our lastIndex:%d is greater than candidate:%s lastIndex:%d",
			lastIdx, candidate, req.LastLogIndex))
		return
	}

	// Persist a vote for safety
	if err = r.persistVote(req.Term, req.Candidate); err != nil {
		klog.Errorf(fmt.Sprintf("failed to persist vote err:%v", err))
		return
	}

	resp.Granted = true
	r.setLastContact()
	return
}

// appendEntries is invoked when we get AppendEntriesRPC call.
func (r *Raft) appendEntries(rpc RPC, cmd *AppendEntriesRequest) {
	resp := &AppendEntriesResponse{
		Term:           r.getCurrentTerm(),
		LastLog:        r.getLastIndex(),
		Success:        false,
		NoRetryBackoff: false,
	}

	var rpcErr error
	defer func() {
		rpc.Respond(resp, rpcErr)
	}()

	// Ignore an older term
	if cmd.Term < r.getCurrentTerm() {
		return
	}

	// Increase the term if we see a newer one, also transition to follower
	// if we ever get an appendEntries call
	if cmd.Term > r.getCurrentTerm() || r.getState() != Follower {
		// Ensure transition to follower
		r.setState(Follower)
		r.setCurrentTerm(cmd.Term)
		resp.Term = cmd.Term
	}

	// Save the current leader
	r.setLeader(r.transport.DecodePeer(cmd.Leader))

	// INFO: 对于 heartbeat AppendEntriesRequest, PrevLogEntry、Entries、LeaderCommitIndex 都是 0
	//  @see https://github.com/hashicorp/raft/blob/v1.3.3/net_transport.go#L587-L592

	// Verify the last log entry, 为何验证 previousLog
	if cmd.PrevLogEntry > 0 {
		lastIdx, lastTerm := r.getLastEntry()
		var prevLogTerm uint64
		if cmd.PrevLogEntry == lastIdx {
			prevLogTerm = lastTerm
		} else {
			var prevLog Log
			if err := r.logs.GetLog(cmd.PrevLogEntry, &prevLog); err != nil {
				klog.Warningf(fmt.Sprintf("failed to get previous log previousLogIndex:%d lastLogIndex:%d error:%v",
					cmd.PrevLogEntry, lastIdx, err))
				resp.NoRetryBackoff = true
				return
			}
			prevLogTerm = prevLog.Term
		}

		if cmd.PrevLogTerm != prevLogTerm {
			klog.Warningf(fmt.Sprintf("previous log term mis-match ours:%d remote:%d", prevLogTerm, cmd.PrevLogTerm))
			resp.NoRetryBackoff = true
			return
		}
	}

	// INFO: (1)store logs in boltdb
	if len(cmd.Entries) > 0 {
		// Delete any conflicting entries, skip any duplicates
		lastLogIdx, _ := r.getLastLog()
		var newEntries []*Log
		for i, entry := range cmd.Entries {
			if entry.Index > lastLogIdx {
				newEntries = cmd.Entries[i:]
				break
			}
			var storeEntry Log
			if err := r.logs.GetLog(entry.Index, &storeEntry); err != nil {
				klog.Warningf(fmt.Sprintf("failed to get log entry index:%d err:%v", entry.Index, err))
				return
			}
			if entry.Term != storeEntry.Term {
				klog.Warningf(fmt.Sprintf("clearing log suffix from:%d to:%d", entry.Index, lastLogIdx))
				if err := r.logs.DeleteRange(entry.Index, lastLogIdx); err != nil {
					klog.Errorf(fmt.Sprintf("failed to clear log suffix err:%v", err))
					return
				}
				if entry.Index <= r.configurations.latestIndex {
					r.setLatestConfiguration(r.configurations.committed, r.configurations.committedIndex)
				}
				newEntries = cmd.Entries[i:]
				break
			}
		}

		if n := len(newEntries); n > 0 {
			// Append the new entries
			if err := r.logs.StoreLogs(newEntries); err != nil {
				klog.Errorf(fmt.Sprintf("failed to append to logs err:%v", err))
				// TODO: leaving r.getLastLog() in the wrong
				// state if there was a truncation above
				return
			}

			// Handle any new configuration changes
			for _, newEntry := range newEntries {
				if err := r.processConfigurationLogEntry(newEntry); err != nil {
					klog.Errorf(fmt.Sprintf("failed to append to logs from index:%d err:%v", newEntry.Index, err))
					rpcErr = err
					return
				}
			}

			// Update the lastLog
			last := newEntries[n-1]
			r.setLastLog(last.Index, last.Term)
		}
	}

	// INFO: (2) apply logs into fsm
	if cmd.LeaderCommitIndex > 0 && cmd.LeaderCommitIndex > r.getCommitIndex() {
		idx := min(cmd.LeaderCommitIndex, r.getLastIndex())
		r.setCommitIndex(idx)
		if r.configurations.latestIndex <= idx {
			r.setCommittedConfiguration(r.configurations.latest, r.configurations.latestIndex)
		}
		r.applyLogsToFSM(idx, nil)
	}

	// Everything went well, set success
	resp.Success = true
	r.setLastContact()
	return
}

// commitTuple is used to send an index that was committed,
// with an optional associated future that should be invoked.
type commitTuple struct {
	log    *Log
	future *logFuture
}

// applyLogsToFSM is used to apply all the committed entries that haven't been
// applied up to the given index limit.
// This can be called from both leaders and followers.
// Followers call this from AppendEntries, for n entries at a time, and always
// pass futures=nil.
// Leaders call this when entries are committed. They pass the futures from any
// inflight logs.
func (r *Raft) applyLogsToFSM(index uint64, futures map[uint64]*logFuture) {
	// Reject logs we've applied already
	lastApplied := r.getLastApplied()
	if index <= lastApplied {
		klog.Warningf(fmt.Sprintf("skipping application of old log index:%d", index))
		return
	}

	applyBatch := func(batch []*commitTuple) {
		select {
		case r.fsmMutateCh <- batch: // INFO: apply batch logs
		case <-r.shutdownCh:
			for _, cl := range batch {
				if cl.future != nil {
					cl.future.respond(ErrRaftShutdown)
				}
			}
		}
	}

	// Store maxAppendEntries for this call in case it ever becomes reloadable. We
	// need to use the same value for all lines here to get the expected result.
	maxAppendEntries := r.config().MaxAppendEntries
	batch := make([]*commitTuple, 0, maxAppendEntries)
	// Apply all the preceding logs
	for idx := lastApplied + 1; idx <= index; idx++ {
		var preparedLog *commitTuple
		// Get the log, either from the future or from our log store
		future, futureOk := futures[idx]
		if futureOk {
			preparedLog = r.prepareLog(&future.log, future)
		} else {
			l := new(Log)
			if err := r.logs.GetLog(idx, l); err != nil {
				klog.Errorf(fmt.Sprintf("failed to get log index:%d error:%v", idx, err))
				panic(err)
			}
			preparedLog = r.prepareLog(l, nil)
		}

		switch {
		case preparedLog != nil:
			// If we have a log ready to send to the FSM add it to the batch.
			// The FSM thread will respond to the future.
			batch = append(batch, preparedLog)

			// If we have filled up a batch, send it to the FSM
			if len(batch) >= maxAppendEntries {
				applyBatch(batch)
				batch = make([]*commitTuple, 0, maxAppendEntries)
			}

		case futureOk:
			// Invoke the future if given.
			future.respond(nil)
		}
	}

	// If there are any remaining logs in the batch apply them
	if len(batch) != 0 {
		applyBatch(batch)
	}

	// Update the lastApplied index and term
	r.setLastApplied(index)
}

// processLog is invoked to process the application of a single committed log entry.
func (r *Raft) prepareLog(l *Log, future *logFuture) *commitTuple {
	switch l.Type {
	case LogBarrier:
		// Barrier is handled by the FSM
		fallthrough // 使用fallthrough强制执行后面的case代码, default 不会执行

	case LogCommand, LogConfiguration:
		return &commitTuple{l, future}

	case LogNoop:
		// Ignore the no-op

	default:
		panic(fmt.Errorf("unrecognized log type: %#v", l))
	}

	return nil
}

// setLastContact is used to set the last contact time to now
func (r *Raft) setLastContact() {
	r.lastContactLock.Lock()
	r.lastContact = time.Now()
	r.lastContactLock.Unlock()
}

// LastContact returns the time of last contact by a leader.
// This only makes sense if we are currently a follower.
func (r *Raft) LastContact() time.Time {
	r.lastContactLock.RLock()
	last := r.lastContact
	r.lastContactLock.RUnlock()
	return last
}

// installSnapshot is invoked when we get a InstallSnapshot RPC call.
// We must be in the follower state for this, since it means we are
// too far behind a leader for log replay. This must only be called
// from the main thread.
func (r *Raft) installSnapshot(rpc RPC, req *InstallSnapshotRequest) {

}

// timeoutNow is what happens when a server receives a TimeoutNowRequest.
func (r *Raft) timeoutNow(rpc RPC, req *TimeoutNowRequest) {

}

// RegisterObserver registers a new observer.
func (r *Raft) RegisterObserver(observer *Observer) {
	r.observersLock.Lock()
	defer r.observersLock.Unlock()
	r.observers[observer.id] = observer
}

// Shutdown is used to stop the Raft background routines.
// This is not a graceful operation. Provides a future that
// can be used to block until all background routines have exited.
func (r *Raft) Shutdown() Future {
	r.shutdownLock.Lock()
	defer r.shutdownLock.Unlock()

	if !r.shutdown {
		close(r.shutdownCh)
		r.shutdown = true
		r.setState(Shutdown)
		return &shutdownFuture{r}
	}

	// avoid closing transport twice
	return &shutdownFuture{nil}
}
