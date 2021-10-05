package raft

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/raft/v3/confchange"
	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/raft/v3/tracker"
	"k8s.io/klog/v2"
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

type CampaignType string

// Possible values for CampaignType
const (
	// campaignPreElection represents the first phase of a normal election when
	// Config.PreVote is true.
	campaignPreElection CampaignType = "CampaignPreElection"
	// campaignElection represents a normal (time-based) election (the second phase
	// of the election when Config.PreVote is true).
	campaignElection CampaignType = "CampaignElection"
	// campaignTransfer represents the type of leader transfer
	campaignTransfer CampaignType = "CampaignTransfer"
)

// lockedRand is a small wrapper around rand.Rand to provide
// synchronization among multiple raft groups.
type lockedRand struct {
	mu   sync.Mutex
	rand *rand.Rand
}

func (r *lockedRand) Intn(n int) int {
	r.mu.Lock()
	v := r.rand.Intn(n)
	r.mu.Unlock()
	return v
}

var globalRand = &lockedRand{
	rand: rand.New(rand.NewSource(time.Now().UnixNano())),
}

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

	// PreVote enables the Pre-Vote algorithm described in raft thesis section
	// 9.6. This prevents disruption when a node that has been partitioned away
	// rejoins the cluster.
	PreVote bool
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

type stepFunc func(m pb.Message) error

var ErrProposalDropped = errors.New("raft proposal dropped")

// INFO: raft struct是raft算法的实现
//  (1) Leader Election: campaign 竞选
//    (1.1)
type raft struct {
	id uint64

	Term uint64
	Vote uint64

	// the log
	raftLog *raftLog

	readStates []ReadState

	maxMsgSize         uint64
	maxUncommittedSize uint64

	state StateType
	// isLearner is true if the local raft node is a learner.
	isLearner bool

	msgs []pb.Message

	// the leader id, 每一个node缓存leader id
	lead uint64
	// leadTransferee is id of the leader transfer target when its value is not zero.
	// Follow the procedure defined in raft thesis 3.10.
	leadTransferee uint64
	// Follower 转发 propose 开关
	disableProposalForwarding bool
	// randomizedElectionTimeout is a random number between
	// [electiontimeout, 2 * electiontimeout - 1]. It gets reset
	// when raft changes its state to follower or candidate.
	randomizedElectionTimeout int

	electionElapsed  int
	electionTimeout  int // 竞选超时时间默认 1000ms
	heartbeatElapsed int // 心跳间隔时间默认 100ms
	heartbeatTimeout int

	progress tracker.ProgressTracker

	checkQuorum bool
	preVote     bool

	// TODO: 做啥的??
	tick func()
	step stepFunc
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

		checkQuorum: config.CheckQuorum,
		preVote:     config.PreVote,

		progress: tracker.MakeProgressTracker(config.MaxInflightMsgs),

		//readOnly:                  newReadOnly(c.ReadOnlyOption),
		//disableProposalForwarding: c.DisableProposalForwarding,
	}

	cfg, progress, err := confchange.Restore(confchange.Changer{
		Tracker:   r.progress,
		LastIndex: raftlog.lastIndex(),
	}, confState)
	if err != nil {
		panic(err)
	}

	if err = confState.Equivalent(r.switchToConfig(cfg, progress)); err != nil {
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
	for _, n := range r.progress.VoterNodes() {
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

func (r *raft) softState() *SoftState {
	return &SoftState{
		Lead:      r.lead,
		RaftState: r.state,
	}
}

func (r *raft) hardState() pb.HardState {
	return pb.HardState{
		Term:   r.Term,
		Vote:   r.Vote,
		Commit: r.raftLog.committed,
	}
}

func (r *raft) hasLeader() bool {
	return r.lead != None
}

// INFO: 加节点, 主要就是修改 progress conf
func (r *raft) applyConfChange(confChange pb.ConfChangeV2) pb.ConfState {
	cfg, progress, err := func() (tracker.Config, tracker.ProgressMap, error) {
		changer := confchange.Changer{
			Tracker:   r.progress,
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

	return r.switchToConfig(cfg, progress)
}

func (r *raft) switchToConfig(cfg tracker.Config, progress tracker.ProgressMap) pb.ConfState {
	r.progress.Config = cfg
	r.progress.Progress = progress
	klog.Infof(fmt.Sprintf("[switchToConfig]raftID:%x switched to configuration %s", r.id, r.progress.Config))

	confState := r.progress.ConfState()
	pr, ok := r.progress.Progress[r.id]
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

// INFO: Follower/PreCandidate/Candidate tick 都是 tickElection() 会在 electionTimeout 之后发起选举
func (r *raft) becomeFollower(term uint64, lead uint64) {
	r.step = r.stepFollower
	r.reset(term)
	r.tick = r.tickElection
	r.lead = lead
	r.state = StateFollower
	klog.Infof("node %x became Follower at term %d", r.id, r.Term)
}

// INFO: pb.Message https://pkg.go.dev/go.etcd.io/etcd/raft/v3#hdr-MessageType
func (r *raft) stepFollower(message pb.Message) error {
	switch message.Type {
	//
	case pb.MsgProp:
		// Follower -> Leader
		if r.lead == None {
			klog.Infof("%x no leader at term %d; dropping proposal", r.id, r.Term)
			return ErrProposalDropped
		} else if r.disableProposalForwarding {
			klog.Infof("%x not forwarding to leader %x at term %d; dropping proposal", r.id, r.lead, r.Term)
			return ErrProposalDropped
		}
		message.To = r.lead
		r.send(message)

	}

	return nil
}

// stepCandidate is shared by StateCandidate and StatePreCandidate; the difference is
// whether they respond to MsgVoteResp or MsgPreVoteResp.
func (r *raft) stepCandidate(message pb.Message) error {

	return nil
}

// INFO: Follower/PreCandidate/Candidate tick 都是 tickElection() 会在 electionTimeout 之后发起选举
func (r *raft) becomeCandidate() {
	if r.state == StateLeader {
		return
	}

	r.step = r.stepCandidate
	r.reset(r.Term + 1) // INFO: 竞选时 term+1
	r.tick = r.tickElection
	r.Vote = r.id
	r.state = StateCandidate
	klog.Infof("node %x became Candidate at term %d", r.id, r.Term)
}

// INFO: Follower/PreCandidate/Candidate tick 都是 tickElection() 会在 electionTimeout 之后发起选举
func (r *raft) becomePreCandidate() {
	if r.state == StateLeader {
		return
	}
	// INFO: PreCandidate 不会增加 term+1, 也不会改变 r.Vote
	r.step = r.stepCandidate
	r.progress.ResetVotes()
	r.tick = r.tickElection
	r.lead = None
	r.state = StatePreCandidate

	klog.Infof("node %x became PreCandidate at term %d", r.id, r.Term)
}

func (r *raft) stepLeader(message pb.Message) error {

	switch message.Type {
	case pb.MsgReadIndex:
		// TODO: 线性一致性读!!!
	case pb.MsgProp: // INFO: 只有 leader 才可以接收写请求
		if len(message.Entries) == 0 {
			return fmt.Errorf("%x stepped empty MsgProp", r.id)
		}
		if r.progress.Progress[r.id] == nil {
			return ErrProposalDropped
		}
		if r.leadTransferee != None {
			klog.Errorf(fmt.Sprintf("%x [term %d] transfer leadership to %x is in progress; dropping proposal", r.id, r.Term, r.leadTransferee))
			return ErrProposalDropped
		}

		for i := range message.Entries {
			entry := &message.Entries[i]
			var cc pb.ConfChangeI
			if entry.Type == pb.EntryConfChangeV2 {
				var ccc pb.ConfChangeV2 // TODO: 不报错么???
				if err := ccc.Unmarshal(entry.Data); err != nil {
					panic(err)
				}
				cc = ccc
			}
			if cc != nil {

			}
		}

		// INFO: 把 entries 先提交给自己
		if !r.appendEntry(message.Entries...) {
			return ErrProposalDropped
		}
		// INFO: 把 entries 异步发给 Follower
		r.bcastAppend()
		return nil
	}

	// All other message types require a progress for m.From (pr).
	pr := r.progress.Progress[message.From]
	if pr == nil {
		klog.Errorf("%x no progress available for %x", r.id, message.From)
		return nil
	}

	return nil
}

func (r *raft) appendEntry(entries ...pb.Entry) (accepted bool) {
	lastIndex := r.raftLog.lastIndex()
	for i := range entries {
		entries[i].Term = r.Term
		entries[i].Index = lastIndex + 1 + uint64(i)
	}

	// TODO:

	// INFO: 追加写append raft log entry, use latest "last" index after truncate/append
	lastIndex = r.raftLog.append(entries...)
	r.progress.Progress[r.id].MaybeUpdate(lastIndex)
	// Regardless of maybeCommit's return, our caller will call bcastAppend.
	r.maybeCommit()
	return true
}

// maybeCommit attempts to advance the commit index. Returns true if
// the commit index changed (in which case the caller should call
// r.bcastAppend).
func (r *raft) maybeCommit() bool {
	mci := r.progress.Committed()
	return r.raftLog.maybeCommit(mci, r.Term)
}

// INFO: leader 把 append log 发送给 follower, r.progress 会记录进度
//  Follower的日志同步进度维护在Progress对象中
func (r *raft) bcastAppend() {
	r.progress.Visit(func(id uint64, _ *tracker.Progress) {
		if id == r.id {
			return
		}
		r.sendAppend(id)
	})
}

// sendAppend sends an append RPC with new entries (if any) and the
// current commit index to the given peer.
// INFO: leader 把 append log 发送给 follower, r.progress 会记录进度
func (r *raft) sendAppend(to uint64) {
	r.maybeSendAppend(to, true)
}

// maybeSendAppend sends an append RPC with new entries to the given peer,
// if necessary. Returns true if a message was sent. The sendIfEmpty
// argument controls whether messages with no entries will be sent
// ("empty" messages are useful to convey updated Commit indexes, but
// are undesirable when we're sending multiple messages in a batch).
func (r *raft) maybeSendAppend(to uint64, sendIfEmpty bool) bool {
	pr := r.progress.Progress[to]
	if pr.IsPaused() {
		return false
	}
	m := pb.Message{}
	m.To = to

	term, errt := r.raftLog.term(pr.Next - 1)
	ents, erre := r.raftLog.entries(pr.Next, r.maxMsgSize)
	if len(ents) == 0 && !sendIfEmpty {
		return false
	}

	if errt != nil || erre != nil { // send snapshot if we failed to get term or entries
		// TODO
	} else {
		m.Type = pb.MsgApp
		m.Index = pr.Next - 1
		m.LogTerm = term
		m.Entries = ents
		m.Commit = r.raftLog.committed
		if n := len(m.Entries); n != 0 {
			switch pr.State {
			// optimistically increase the next when in StateReplicate
			case tracker.StateReplicate:
				last := m.Entries[n-1].Index
				pr.OptimisticUpdate(last)
				pr.Inflights.Add(last)
			case tracker.StateProbe:
				pr.ProbeSent = true
			default:
				klog.Fatalf(fmt.Sprintf("node %x is sending append in unhandled state %s", r.id, pr.State))
			}
		}
	}

	r.send(m)
	return true
}

// INFO: Leader tick 是 tickHeartbeat(), 会发心跳给 Follower/PreCandidate/Candidate
func (r *raft) becomeLeader() {
	if r.state == StateFollower {
		return
	}

	r.step = r.stepLeader
	r.reset(r.Term)
	r.tick = r.tickHeartbeat
	r.lead = r.id
	r.state = StateLeader
	klog.Infof("node %x became Leader at term %d", r.id, r.Term)
}

// INFO: 重置一些关键字段值
func (r *raft) reset(term uint64) {
	if r.Term != term {
		r.Term = term
		r.Vote = None
	}
	r.lead = None

	r.electionElapsed = 0
	r.heartbeatElapsed = 0
	r.resetRandomizedElectionTimeout()

	r.abortLeaderTransfer()

	r.progress.ResetVotes()
	r.progress.Visit(func(id uint64, pr *tracker.Progress) {
		*pr = tracker.Progress{
			Match:     0,
			Next:      r.raftLog.lastIndex() + 1,
			Inflights: tracker.NewInflights(r.progress.MaxInflight),
			IsLearner: pr.IsLearner,
		}
		if id == r.id {
			pr.Match = r.raftLog.lastIndex()
		}
	})

	//r.pendingConfIndex = 0
	//r.uncommittedSize = 0
	//r.readOnly = newReadOnly(r.readOnly.option)
}

func (r *raft) abortLeaderTransfer() {
	r.leadTransferee = None
}

// Raft 为了优化选票被瓜分导致选举失败的问题，引入了随机数，每个节点等待发起选举的时间点不一致，优雅的解决了潜在的竞选活锁，同时易于理解
func (r *raft) resetRandomizedElectionTimeout() {
	r.randomizedElectionTimeout = r.electionTimeout + globalRand.Intn(r.electionTimeout)
}

// send schedules persisting state to a stable storage and AFTER that
// sending the message (as part of next Ready message processing).
func (r *raft) send(message pb.Message) {
	if message.Type == pb.MsgVote || message.Type == pb.MsgVoteResp || message.Type == pb.MsgPreVote || message.Type == pb.MsgPreVoteResp {
		if message.Term == 0 {
			klog.Fatalf(fmt.Sprintf("term should be set when sending %s", message.Type))
		}
	} else {
		if message.Term != 0 {
			klog.Fatalf(fmt.Sprintf("term should not be set when sending %s (was %d)", message.Type, message.Term))
		}

		if message.Type != pb.MsgProp && message.Type != pb.MsgReadIndex {
			message.Term = r.Term
		}
	}

	// INFO: @see newReady()
	r.msgs = append(r.msgs, message)
}

// tickElection INFO: r.electionTimeout 过期之后，Follower/Candidate 发起选举
func (r *raft) tickElection() {
	r.electionElapsed++
	if r.promotable() && r.pastElectionTimeout() {
		r.electionElapsed = 0
		if err := r.Step(pb.Message{From: r.id, Type: pb.MsgHup}); err != nil {
			klog.Errorf("error occurred during election: %v", err)
		}
	}
}

// INFO: 在 r.heartbeatTimeout 之后, leader 发 MsgBeat 给 Follower/PreCandidate/Candidate
func (r *raft) tickHeartbeat() {
	r.heartbeatElapsed++
	r.electionElapsed++

	if r.state != StateLeader {
		return
	}

	// TODO: election

	if r.heartbeatElapsed >= r.heartbeatTimeout {
		r.heartbeatElapsed = 0 // 重置
		// INFO: Leader 发送心跳
		if err := r.Step(pb.Message{From: r.id, Type: pb.MsgBeat}); err != nil {
			klog.Errorf(fmt.Sprintf("error occurred during checking sending heartbeat: %v", err))
		}
	}
}

func (r *raft) Step(message pb.Message) error {
	// Handle the message term, which may result in our stepping down to a follower.
	switch {
	case message.Term == 0:
		// local message
	case message.Term > r.Term: // INFO: 如果 message term 大于当前节点的 term
		if message.Type == pb.MsgVote || message.Type == pb.MsgPreVote {
			force := bytes.Equal(message.Context, []byte(campaignTransfer))
			inLease := r.checkQuorum && r.lead != None && r.electionElapsed < r.electionTimeout
			if !force && inLease {
				// If a server receives a RequestVote request within the minimum election timeout
				// of hearing from a current leader, it does not update its term or grant its vote
				klog.Infof("%x [logterm: %d, index: %d, vote: %x] ignored %s from %x [logterm: %d, index: %d] at term %d: lease is not expired (remaining ticks: %d)",
					r.id, r.raftLog.lastTerm(), r.raftLog.lastIndex(), r.Vote, message.Type, message.From, message.LogTerm, message.Index, r.Term, r.electionTimeout-r.electionElapsed)
				return nil
			}
		}

		switch {
		case message.Type == pb.MsgPreVote:
			// Never change our term in response to a PreVote
		case message.Type == pb.MsgPreVoteResp && !message.Reject:
		default:
			klog.Infof("%x [term: %d] received a %s message with higher term from %x [term: %d]",
				r.id, r.Term, message.Type, message.From, message.Term)
			if message.Type == pb.MsgApp || message.Type == pb.MsgHeartbeat || message.Type == pb.MsgSnap {
				r.becomeFollower(message.Term, message.From)
			} else {
				r.becomeFollower(message.Term, None)
			}
		}
	case message.Term < r.Term: // INFO: 如果 message term 小于当前节点的 term
		if (r.checkQuorum || r.preVote) && (message.Type == pb.MsgHeartbeat || message.Type == pb.MsgApp) {
			r.send(pb.Message{To: message.From, Type: pb.MsgAppResp})
		} else if message.Type == pb.MsgPreVote {
			klog.Infof(fmt.Sprintf("[raft Step]%x [logterm: %d, index: %d, vote: %x] rejected %s from %x [logterm: %d, index: %d] at term %d",
				r.id, r.raftLog.lastTerm(), r.raftLog.lastIndex(), r.Vote, message.Type, message.From, message.LogTerm, message.Index, r.Term))
		} else {
			klog.Infof(fmt.Sprintf("%x [term: %d] ignored a %s message with lower term from %x [term: %d]",
				r.id, r.Term, message.Type, message.From, message.Term))
		}

		return nil
	}

	switch message.Type {
	case pb.MsgHup:
		if r.preVote {
			r.hup(campaignPreElection)
		} else {
			r.hup(campaignElection)
		}

	case pb.MsgVote, pb.MsgPreVote:
		// TODO:

	default:
		err := r.step(message)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *raft) hup(t CampaignType) {
	// TODO:
}

// pastElectionTimeout returns true iff r.electionElapsed is greater
// than or equal to the randomized election timeout in
// [electiontimeout, 2 * electiontimeout - 1].
func (r *raft) pastElectionTimeout() bool {
	return r.electionElapsed >= r.randomizedElectionTimeout
}

// INFO: state machine 是否可以 promoted to be leader
func (r *raft) promotable() bool {
	pr := r.progress.Progress[r.id]
	return pr != nil && !pr.IsLearner && !r.raftLog.hasPendingSnapshot()
}

// Ready encapsulates the entries and messages that are ready to read,
// be saved to stable storage, committed or sent to other peers.
// All fields in Ready are read-only.
type Ready struct {
	Entries []pb.Entry

	pb.HardState

	*SoftState

	// ReadStates can be used for node to serve linearizable read requests locally
	// when its applied index is greater than the index in ReadState.
	// Note that the readState will be returned when raft receives msgReadIndex.
	// The returned is only valid for the request that requested to read.
	ReadStates []ReadState

	// INFO: 已经提交到 state-machine 中的 []pb.Entry
	CommittedEntries []pb.Entry

	Snapshot pb.Snapshot

	Messages []pb.Message

	MustSync bool
}

// MustSync returns true if the hard state and count of Raft entries indicate
// that a synchronous write to persistent storage is required.
func MustSync(st, prevst pb.HardState, entsnum int) bool {
	// Persistent state on all servers:
	// (Updated on stable storage before responding to RPCs)
	// currentTerm
	// votedFor
	// log entries[]
	return entsnum != 0 || st.Vote != prevst.Vote || st.Term != prevst.Term
}
func (r *raft) newReady(prevSoftSt *SoftState, prevHardSt pb.HardState) Ready {
	ready := Ready{
		Entries:          r.raftLog.unstableEntries(),
		CommittedEntries: r.raftLog.nextEnts(),
		Messages:         r.msgs,
	}
	if softSt := r.softState(); !softSt.equal(prevSoftSt) {
		ready.SoftState = softSt
	}
	if hardSt := r.hardState(); !isHardStateEqual(hardSt, prevHardSt) {
		ready.HardState = hardSt
	}
	if r.raftLog.unstable.snapshot != nil {
		ready.Snapshot = *r.raftLog.unstable.snapshot
	}
	if len(r.readStates) != 0 {
		ready.ReadStates = r.readStates
	}
	ready.MustSync = MustSync(r.hardState(), prevHardSt, len(ready.Entries))

	return ready
}
