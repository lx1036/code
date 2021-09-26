package raft

import (
	"errors"
	"fmt"
	"k8s.io/klog/v2"
	"math"
	"strings"

	"go.etcd.io/etcd/raft/v3/confchange"
	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/raft/v3/tracker"
)

// None is a placeholder node ID used when there is no leader.
const None uint64 = 0
const noLimit = math.MaxUint64

type StateType uint64

const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
	StatePreCandidate
	numStates
)

type ReadOnlyOption int

const (
	// ReadOnlySafe guarantees the linearizability of the read only request by
	// communicating with the quorum. It is the default and suggested option.
	ReadOnlySafe ReadOnlyOption = iota

	// ReadOnlyLeaseBased ensures linearizability of the read only request by
	// relying on the leader lease. It can be affected by clock drift.
	// If the clock drift is unbounded, leader might keep the lease longer than it
	// should (clock can move backward/pause without any bound). ReadIndex is not safe
	// in that case.
	ReadOnlyLeaseBased
)

// Config contains the parameters to start a raft.
type Config struct {
	// ID is the identity of the local raft. ID cannot be 0.
	ID uint64

	// peers contains the IDs of all nodes (including self) in the raft cluster. It
	// should only be set when starting a new raft cluster. Restarting raft from
	// previous configuration will panic if peers is set. peer is private and only
	// used for testing right now.
	peers []uint64

	// ElectionTick is the number of Node.Tick invocations that must pass between
	// elections. That is, if a follower does not receive any message from the
	// leader of current term before ElectionTick has elapsed, it will become
	// candidate and start an election. ElectionTick must be greater than
	// HeartbeatTick. We suggest ElectionTick = 10 * HeartbeatTick to avoid
	// unnecessary leader switching.
	ElectionTick int
	// HeartbeatTick is the number of Node.Tick invocations that must pass between
	// heartbeats. That is, a leader sends heartbeat messages to maintain its
	// leadership every HeartbeatTick ticks.
	HeartbeatTick int

	// Storage is the storage for raft. raft generates entries and states to be
	// stored in storage. raft reads the persisted entries and states out of
	// Storage when it needs. raft reads out the previous state and configuration
	// out of storage when restarting.
	Storage Storage

	// MaxSizePerMsg limits the max byte size of each append message. Smaller
	// value lowers the raft recovery cost(initial probing and message lost
	// during normal operation). On the other side, it might affect the
	// throughput during normal replication. Note: math.MaxUint64 for unlimited,
	// 0 for at most one entry per message.
	MaxSizePerMsg uint64
	// MaxCommittedSizePerReady limits the size of the committed entries which
	// can be applied.
	MaxCommittedSizePerReady uint64
	// MaxUncommittedEntriesSize limits the aggregate byte size of the
	// uncommitted entries that may be appended to a leader's log. Once this
	// limit is exceeded, proposals will begin to return ErrProposalDropped
	// errors. Note: 0 for no limit.
	MaxUncommittedEntriesSize uint64

	// MaxInflightMsgs limits the max number of in-flight append messages during
	// optimistic replication phase. The application transportation layer usually
	// has its own sending buffer over TCP/UDP. Setting MaxInflightMsgs to avoid
	// overflowing that sending buffer. TODO (xiangli): feedback to application to
	// limit the proposal rate?
	MaxInflightMsgs int

	// ReadOnlyOption specifies how the read only request is processed.
	//
	// ReadOnlySafe guarantees the linearizability of the read only request by
	// communicating with the quorum. It is the default and suggested option.
	//
	// ReadOnlyLeaseBased ensures linearizability of the read only request by
	// relying on the leader lease. It can be affected by clock drift.
	// If the clock drift is unbounded, leader might keep the lease longer than it
	// should (clock can move backward/pause without any bound). ReadIndex is not safe
	// in that case.
	// CheckQuorum MUST be enabled if ReadOnlyOption is ReadOnlyLeaseBased.
	ReadOnlyOption ReadOnlyOption

	// CheckQuorum specifies if the leader should check quorum activity. Leader
	// steps down when quorum is not active for an electionTimeout.
	CheckQuorum bool
}

func (c *Config) validate() error {
	if c.ID == None {
		return errors.New("cannot use none as id")
	}

	if c.HeartbeatTick <= 0 {
		return errors.New("heartbeat tick must be greater than 0")
	}

	if c.ElectionTick <= c.HeartbeatTick {
		return errors.New("election tick must be greater than heartbeat tick")
	}

	if c.Storage == nil {
		return errors.New("storage cannot be nil")
	}

	if c.MaxUncommittedEntriesSize == 0 {
		c.MaxUncommittedEntriesSize = noLimit
	}

	// default MaxCommittedSizePerReady to MaxSizePerMsg because they were
	// previously the same parameter.
	if c.MaxCommittedSizePerReady == 0 {
		c.MaxCommittedSizePerReady = c.MaxSizePerMsg
	}

	if c.MaxInflightMsgs <= 0 {
		return errors.New("max inflight messages must be greater than 0")
	}

	if c.ReadOnlyOption == ReadOnlyLeaseBased && !c.CheckQuorum {
		return errors.New("CheckQuorum must be enabled when ReadOnlyOption is ReadOnlyLeaseBased")
	}

	return nil
}

// raft struct是raft算法的实现
type raft struct {
	id uint64

	Term uint64
	Vote uint64

	// the log
	raftLog *raftLog

	maxMsgSize         uint64
	maxUncommittedSize uint64

	state StateType
	// isLearner is true if the local raft node is a learner.
	isLearner bool

	msgs []pb.Message

	// the leader id
	lead uint64

	electionElapsed  int
	electionTimeout  int
	heartbeatElapsed int
	heartbeatTimeout int

	prs tracker.ProgressTracker
}

func newRaft(config *Config) *raft {
	if err := config.validate(); err != nil {
		panic(err.Error())
	}

	raftlog := newLogWithSize(config.Storage, config.MaxCommittedSizePerReady)
	hardState, confState, err := config.Storage.InitialState()
	if err != nil {
		panic(err)
	}

	r := &raft{
		id:                 config.ID,
		lead:               None,
		isLearner:          false,
		raftLog:            raftlog,
		maxMsgSize:         config.MaxSizePerMsg,
		maxUncommittedSize: config.MaxUncommittedEntriesSize,
		electionTimeout:    config.ElectionTick,
		heartbeatTimeout:   config.HeartbeatTick,

		prs: tracker.MakeProgressTracker(config.MaxInflightMsgs),

		//checkQuorum:               c.CheckQuorum,
		//preVote:                   c.PreVote,
		//readOnly:                  newReadOnly(c.ReadOnlyOption),
		//disableProposalForwarding: c.DisableProposalForwarding,
	}

	cfg, prs, err := confchange.Restore(confchange.Changer{
		Tracker:   r.prs,
		LastIndex: raftlog.lastIndex(),
	}, confState)
	if err != nil {
		panic(err)
	}

	if err = confState.Equivalent(r.switchToConfig(cfg, prs)); err != nil {
		panic(err)
	}

	if !IsEmptyHardState(hardState) {
		r.loadState(hardState)
	}
	/*if config.Applied > 0 {
		raftlog.appliedTo(config.Applied)
	}*/

	r.becomeFollower(r.Term, None)

	var nodesStrs []string
	for _, n := range r.prs.VoterNodes() {
		nodesStrs = append(nodesStrs, fmt.Sprintf("%x", n))
	}
	klog.Infof("newRaft %x [peers: [%s], term: %d, commit: %d, applied: %d, lastindex: %d, lastterm: %d]",
		r.id, strings.Join(nodesStrs, ","), r.Term, r.raftLog.committed, r.raftLog.applied, r.raftLog.lastIndex(),
		0, //r.raftLog.lastTerm(),
	)

	return r
}

func (r *raft) loadState(state pb.HardState) {
	if state.Commit < r.raftLog.committed || state.Commit > r.raftLog.lastIndex() {
		klog.Fatalf(fmt.Sprintf("[loadState]%x state.commit %d is out of range [%d, %d]", r.id, state.Commit,
			r.raftLog.committed, r.raftLog.lastIndex()))
	}

	r.raftLog.committed = state.Commit
	r.Term = state.Term
	r.Vote = state.Vote
}

// INFO: 加节点
func (r *raft) applyConfChange(confChange pb.ConfChangeV2) pb.ConfState {
	cfg, prs, err := func() (tracker.Config, tracker.ProgressMap, error) {
		changer := confchange.Changer{
			Tracker:   r.prs,
			LastIndex: r.raftLog.lastIndex(),
		}
		if confChange.LeaveJoint() {
			return changer.LeaveJoint()
		} else if autoLeave, ok := confChange.EnterJoint(); ok {
			return changer.EnterJoint(autoLeave, confChange.Changes...)
		}
		return changer.Simple(confChange.Changes...)
	}()
	if err != nil {
		// TODO(tbg): return the error to the caller.
		panic(err)
	}

	return r.switchToConfig(cfg, prs)
}

func (r *raft) switchToConfig(cfg tracker.Config, prs tracker.ProgressMap) pb.ConfState {
	r.prs.Config = cfg
	r.prs.Progress = prs
	klog.Infof(fmt.Sprintf("[switchToConfig]%x switched to configuration %s", r.id, r.prs.Config))

	confState := r.prs.ConfState()
	pr, ok := r.prs.Progress[r.id]
	r.isLearner = ok && pr.IsLearner
	if (!ok || r.isLearner) && r.state == StateLeader {
		return confState
	}

	if r.state != StateLeader || len(confState.Voters) == 0 {
		return confState
	}

	// TODO: r.maybeCommit()

	return confState
}

func (r *raft) becomeFollower(term uint64, lead uint64) {
	//r.step = stepFollower
	//r.reset(term)
	//r.tick = r.tickElection
	r.lead = lead
	r.state = StateFollower
	klog.Infof("%x became follower at term %d", r.id, r.Term)
}

// INFO: state machine 是否可以 promoted to be leader
func (r *raft) promotable() bool {
	pr := r.prs.Progress[r.id]
	return pr != nil && !pr.IsLearner && !r.raftLog.hasPendingSnapshot()
}
