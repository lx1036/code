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

	"k8s-lx1036/k8s/storage/etcd/raft/confchange"
	"k8s-lx1036/k8s/storage/etcd/raft/tracker"

	"go.etcd.io/etcd/raft/v3/quorum"
	pb "go.etcd.io/etcd/raft/v3/raftpb"
	"k8s.io/klog/v2"
)

// INFO:
//  etcd-线性一致性(linearizability): https://zhuanlan.zhihu.com/p/386335327
//  线性一致性和 Raft: https://pingcap.com/zh/blog/linearizability-and-raft

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
	id uint64 // INFO: 对于单 node, raft id 就是 node id

	Term uint64
	Vote uint64

	// the log
	raftLog *raftLog

	readOnly   *readOnly
	readStates []ReadState

	maxMsgSize         uint64
	maxUncommittedSize uint64

	state StateType
	// isLearner is true if the local raft node is a learner.
	isLearner bool

	msgs []pb.Message

	// the leader id, 每一个node缓存leader id
	lead uint64
	// an estimate of the size of the uncommitted tail of the Raft log. Used to
	// prevent unbounded log growth. Only maintained by the leader. Reset on
	// term changes.
	uncommittedSize uint64
	// Only one conf change may be pending (in the log, but not yet
	// applied) at a time. This is enforced via pendingConfIndex, which
	// is set to a value >= the log index of the latest pending
	// configuration change (if any). Config changes are only allowed to
	// be proposed if the leader's applied index is greater than this
	// value.
	pendingConfIndex uint64

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

	// INFO: 这里不同角色 raft node，tick() 函数不一样：对于 Leader node, tick() 就是 tickHeartbeat；对于 Follower/PreCandidate/Candidate，tick() 就是 tickElection
	tick func()
	step stepFunc

	// pendingReadIndexMessages is used to store messages of type MsgReadIndex
	// that can't be answered as new leader didn't committed any log in
	// current term. Those will be handled as fast as first log is committed in
	// current term.
	pendingReadIndexMessages []pb.Message
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

		readOnly: newReadOnly(config.ReadOnlyOption),

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

// switchToConfig reconfigures this node to use the provided configuration.
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
	// INFO: 这里不同角色 raft node，tick() 函数不一样：对于 Leader node, tick() 就是 tickHeartbeat；对于 Follower/PreCandidate/Candidate，tick() 就是 tickElection
	r.tick = r.tickElection
	r.lead = lead
	r.state = StateFollower
	klog.Infof("node %x became Follower at term %d", r.id, r.Term)
}

// INFO: Follower/PreCandidate/Candidate tick 都是 tickElection() 会在 electionTimeout 之后发起选举
func (r *raft) becomePreCandidate() {
	if r.state == StateLeader {
		return
	}
	// INFO: PreCandidate 不会增加 term+1, 也不会改变 r.Vote
	r.step = r.stepCandidate
	r.progress.ResetVotes()
	// INFO: 这里不同角色 raft node，tick() 函数不一样：对于 Leader node, tick() 就是 tickHeartbeat；对于 Follower/PreCandidate/Candidate，tick() 就是 tickElection
	r.tick = r.tickElection
	r.lead = None
	r.state = StatePreCandidate

	klog.Infof("node %x became PreCandidate at term %d", r.id, r.Term)
}

// INFO: Follower/PreCandidate/Candidate tick 都是 tickElection() 会在 electionTimeout 之后发起选举
func (r *raft) becomeCandidate() {
	if r.state == StateLeader {
		return
	}

	r.step = r.stepCandidate
	r.reset(r.Term + 1) // INFO: 竞选时 term+1
	// INFO: 这里不同角色 raft node，tick() 函数不一样：对于 Leader node, tick() 就是 tickHeartbeat；对于 Follower/PreCandidate/Candidate，tick() 就是 tickElection
	r.tick = r.tickElection
	r.Vote = r.id
	r.state = StateCandidate
	klog.Infof("node %x became Candidate at term %d", r.id, r.Term)
}

// INFO: Leader tick 是 tickHeartbeat(), 会发心跳给 Follower/PreCandidate/Candidate
func (r *raft) becomeLeader() {
	if r.state == StateFollower {
		return
	}

	r.step = r.stepLeader
	r.reset(r.Term)
	// INFO: 这里不同角色 raft node，tick() 函数不一样：对于 Leader node, tick() 就是 tickHeartbeat；对于 Follower/PreCandidate/Candidate，tick() 就是 tickElection
	r.tick = r.tickHeartbeat
	r.lead = r.id
	r.state = StateLeader
	klog.Infof("node %x became Leader at term %d", r.id, r.Term)
}

// INFO: pb.Message https://pkg.go.dev/go.etcd.io/etcd/raft/v3#hdr-MessageType
func (r *raft) stepFollower(message pb.Message) error {
	switch message.Type {
	case pb.MsgHeartbeat: // 收到来自 leader 的 MsgHeartbeat
		r.electionElapsed = 0
		r.lead = message.From
		r.handleHeartbeat(message)

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

	// INFO: Follower 的线性一致性读处理逻辑
	case pb.MsgReadIndex:
		if r.lead == None {
			klog.Infof(fmt.Sprintf("%x no leader at term %d; dropping index reading msg", r.id, r.Term))
			return nil
		}

		// INFO: follower 转发 MsgReadIndex 给 leader，获取 leader 中状态机已经应用的 appliedIndex。只有 leader 处理 MsgReadIndex 请求
		message.To = r.lead
		r.send(message)
	case pb.MsgReadIndexResp:
		if len(message.Entries) != 1 {
			klog.Errorf(fmt.Sprintf("%x invalid format of MsgReadIndexResp from %x, entries count: %d", r.id, message.From, len(message.Entries)))
			return nil
		}
		r.readStates = append(r.readStates, ReadState{
			Index:      message.Index,
			RequestCtx: message.Entries[0].Data,
		})
	}

	return nil
}

// stepCandidate is shared by StateCandidate and StatePreCandidate; the difference is
// whether they respond to MsgVoteResp or MsgPreVoteResp.
func (r *raft) stepCandidate(message pb.Message) error {

	switch message.Type {
	case pb.MsgHeartbeat:

	}

	return nil
}

// INFO: leader 节点接收到了用户提交的 Put/Delete propose, 见 stepLeader()
func (r *raft) stepLeader(message pb.Message) error {
	switch message.Type {
	case pb.MsgBeat:
		r.bcastHeartbeat()
		return nil

	// INFO: Leader 的线性一致性读处理逻辑
	case pb.MsgReadIndex:
		// INFO: 线性一致性读!!!
		// only one leader member in cluster
		if r.progress.IsSingleton() {
			/*if resp := r.responseToReadIndexReq(message, r.raftLog.committed); resp.To != None {
				r.send(resp)
			}*/

			return nil
		}

		// Postpone read only request when this leader has not committed
		// any log entry at its term.
		if !r.committedEntryInCurrentTerm() {
			r.pendingReadIndexMessages = append(r.pendingReadIndexMessages, message)
			return nil
		}

		//sendMsgReadIndexResponse(r, message)
		return nil

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

		// INFO: 如果 entry 是 ConfChange type
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

		// INFO: 把 entries 先提交给自己 raft log 模块，这时 entries 存储在 memory storage 中
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

// INFO: 如果 committed term 已经追赶到了当前的 term
func (r *raft) committedEntryInCurrentTerm() bool {
	return r.raftLog.zeroTermOnErrCompacted(r.raftLog.term(r.raftLog.committed)) == r.Term
}

// INFO: 把 entries 先提交给自己 raft log 模块，这时 entries 存储在 memory storage 中
func (r *raft) appendEntry(entries ...pb.Entry) (accepted bool) {
	lastIndex := r.raftLog.lastIndex()
	for i := range entries { // leader 需要更新下 entry Term 和 Index
		entries[i].Term = r.Term
		entries[i].Index = lastIndex + 1 + uint64(i)
	}

	if !r.increaseUncommittedSize(entries) {
		klog.Warningf(fmt.Sprintf("%x appending new entries to log would exceed uncommitted entry size limit; dropping proposal", r.id))
		return false
	}

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
	commitedIndex := r.progress.Committed()
	return r.raftLog.maybeCommit(commitedIndex, r.Term)
}

// bcastHeartbeat sends RPC, without entries to all the peers.
func (r *raft) bcastHeartbeat() {
	lastCtx := r.readOnly.lastPendingRequestCtx()
	if len(lastCtx) == 0 {
		r.bcastHeartbeatWithCtx(nil)
	} else {
		r.bcastHeartbeatWithCtx([]byte(lastCtx))
	}
}

func (r *raft) bcastHeartbeatWithCtx(ctx []byte) {
	r.progress.Visit(func(id uint64, _ *tracker.Progress) {
		if id == r.id {
			return
		}

		r.sendHeartbeat(id, ctx)
	})
}

// sendHeartbeat sends a heartbeat RPC to the given peer.
func (r *raft) sendHeartbeat(to uint64, ctx []byte) {
	commit := min(r.progress.Progress[to].Match, r.raftLog.committed)
	m := pb.Message{
		To:      to,
		Type:    pb.MsgHeartbeat,
		Commit:  commit,
		Context: ctx,
	}

	r.send(m)
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

func (r *raft) abortLeaderTransfer() {
	r.leadTransferee = None
}

// Raft 为了优化选票被瓜分导致选举失败的问题，引入了随机数，每个节点等待发起选举的时间点不一致，优雅的解决了潜在的竞选活锁，同时易于理解
func (r *raft) resetRandomizedElectionTimeout() {
	r.randomizedElectionTimeout = r.electionTimeout + globalRand.Intn(r.electionTimeout)
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

func (r *raft) handleHeartbeat(m pb.Message) {
	r.raftLog.commitTo(m.Commit)
	r.send(pb.Message{To: m.From, Type: pb.MsgHeartbeatResp, Context: m.Context}) // 返回的是 MsgHeartbeatResp
}

// Step INFO: node 里会监听用户的 propose channel，并调用 raft.Step 来推动 raft
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
		err := r.step(message) // INFO: leader 节点接收到了用户提交的 Put/Delete propose, 见 stepLeader()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *raft) increaseUncommittedSize(ents []pb.Entry) bool {
	var s uint64
	for _, e := range ents {
		s += uint64(len(e.Data))
	}

	if r.uncommittedSize > 0 && s > 0 && r.uncommittedSize+s > r.maxUncommittedSize {
		return false
	}

	r.uncommittedSize += s
	return true
}

func (r *raft) reduceUncommittedSize(entries []pb.Entry) {
	if r.uncommittedSize == 0 {
		// Fast-path for followers, who do not track or enforce the limit.
		return
	}

	var s uint64
	for _, e := range entries {
		s += uint64(len(e.Data))
	}

	if s > r.uncommittedSize {
		// uncommittedSize may underestimate the size of the uncommitted Raft
		// log tail but will never overestimate it. Saturate at 0 instead of
		// allowing overflow.
		r.uncommittedSize = 0
	} else {
		r.uncommittedSize -= s
	}
}

// INFO: campaign 成为 follower -> leader
func (r *raft) hup(t CampaignType) {
	if r.state == StateLeader {
		klog.Warningf(fmt.Sprintf("%x ignoring MsgHup because already leader", r.id))
		return
	}
	if !r.promotable() {
		klog.Warningf(fmt.Sprintf("%x is unpromotable and can not campaign", r.id))
		return
	}

	// INFO: 检查 ConfChange 是否有 pending
	ents, err := r.raftLog.slice(r.raftLog.applied+1, r.raftLog.committed+1, noLimit)
	if err != nil {
		klog.Fatalf(fmt.Sprintf("unexpected error getting unapplied entries (%v)", err))
	}
	if n := numOfPendingConf(ents); n != 0 && r.raftLog.committed > r.raftLog.applied {
		klog.Warningf(fmt.Sprintf("%x cannot campaign at term %d since there are still %d pending configuration changes to apply", r.id, r.Term, n))
		return
	}

	klog.Infof(fmt.Sprintf("%x is starting a new election at term %d", r.id, r.Term))
	r.campaign(t)
}

func (r *raft) campaign(t CampaignType) {
	if !r.promotable() {
		klog.Warningf(fmt.Sprintf("%x is unpromotable; campaign() should have been called", r.id))
	}

	//var term uint64
	var voteMsg pb.MessageType
	if t == campaignPreElection {
		r.becomePreCandidate()
		voteMsg = pb.MsgPreVote
		// PreVote RPCs are sent for the next term before we've incremented r.Term.
		//term = r.Term + 1
	} else {
		r.becomeCandidate()
		voteMsg = pb.MsgVote
		//term = r.Term
	}

	if _, _, res := r.poll(r.id, voteRespMsgType(voteMsg), true); res == quorum.VoteWon {
		// We won the election after voting for ourselves (which must mean that
		// this is a single-node cluster). Advance to the next state.
		if t == campaignPreElection {
			r.campaign(campaignElection)
		} else {
			r.becomeLeader()
		}
		return
	}

	// TODO: campaignPreElection
}

func (r *raft) poll(id uint64, t pb.MessageType, v bool) (granted int, rejected int, result quorum.VoteResult) {
	if v {
		klog.Infof(fmt.Sprintf("%x received %s from %x at term %d", r.id, t, id, r.Term))
	} else {
		klog.Infof(fmt.Sprintf("%x received %s rejection from %x at term %d", r.id, t, id, r.Term))
	}

	r.progress.RecordVote(id, v)

	return r.progress.TallyVotes()
}

func numOfPendingConf(ents []pb.Entry) int {
	n := 0
	for i := range ents {
		if ents[i].Type == pb.EntryConfChange || ents[i].Type == pb.EntryConfChangeV2 {
			n++
		}
	}
	return n
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

// INFO: 推动用户提交的 []pb.Message 在 ready 结构体中，然后要提交到 log 模块中，这里才是最终目标!!!
//  调用 Node.Advance() 通知Node，之前调用 Node.Ready() 所接受的数据已经被异步应用到状态机了
func (r *raft) advance(ready Ready) {
	r.reduceUncommittedSize(ready.CommittedEntries)

	if newApplied := ready.appliedCursor(); newApplied > 0 {
		oldApplied := r.raftLog.applied
		r.raftLog.appliedTo(newApplied)

		// INFO: 必须是 leader
		if r.state == StateLeader && r.progress.Config.AutoLeave && oldApplied <= r.pendingConfIndex && newApplied >= r.pendingConfIndex {
			ent := pb.Entry{
				Type: pb.EntryConfChangeV2,
				Data: nil,
			}
			// There's no way in which this proposal should be able to be rejected.
			if !r.appendEntry(ent) {
				panic("refused un-refusable auto-leaving ConfChangeV2")
			}

			r.pendingConfIndex = r.raftLog.lastIndex()
			klog.Infof(fmt.Sprintf("[raft advance]initiating automatic transition out of joint configuration %s", r.progress.Config))
		}
	}

	if len(ready.Entries) > 0 {
		e := ready.Entries[len(ready.Entries)-1]
		r.raftLog.stableTo(e.Index, e.Term)
	}

	if !IsEmptySnap(ready.Snapshot) {
		r.raftLog.stableSnapTo(ready.Snapshot.Metadata.Index)
	}
}

// INFO: raft 没有去实现 transport 模块，交给用户去实现。所以只是把 []pb.Message 存储到 r.msgs，然后在 node.Ready() 中被 raft 外部调用
//  参考 KVServer 中上层调用模块 raft 中, 获得 raftNode.Ready() 之后在使用 transport.Send() 发给各个 follower
func (r *raft) send(message pb.Message) {
	if message.From == None {
		message.From = r.id
	}

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

func (r *raft) checkRaftID(snapshot pb.Snapshot) bool {
	confState := snapshot.Metadata.ConfState
	// `LearnersNext` doesn't need to be checked. According to the rules, if a peer in `LearnersNext`, it has to be in `VotersOutgoing`.
	for _, ids := range [][]uint64{confState.Voters, confState.Learners, confState.LearnersNext} {
		for _, id := range ids {
			if id == r.id {
				return true
			}
		}
	}

	return false
}

// restore recovers the state machine from a snapshot.
func (r *raft) restore(snapshot pb.Snapshot) bool {
	if snapshot.Metadata.Index <= r.raftLog.committed {
		return false
	}

	// INFO: 从 snapshot recover 还必须是 Follower
	if r.state != StateFollower {
		klog.Warningf(fmt.Sprintf("%x attempted to restore snapshot as leader; should never happen", r.id))
		r.becomeFollower(r.Term+1, None)
		return false
	}

	if !r.checkRaftID(snapshot) {
		klog.Warningf(fmt.Sprintf("%x attempted to restore snapshot but it is not in the ConfState %v; should never happen", r.id, snapshot.Metadata.ConfState))
		return false
	}

	if r.raftLog.matchTerm(snapshot.Metadata.Index, snapshot.Metadata.Term) {
		klog.Infof(fmt.Sprintf("%x [commit: %d, lastindex: %d, lastterm: %d] fast-forwarded commit to snapshot [index: %d, term: %d]",
			r.id, r.raftLog.committed, r.raftLog.lastIndex(), r.raftLog.lastTerm(), snapshot.Metadata.Index, snapshot.Metadata.Term))
		r.raftLog.commitTo(snapshot.Metadata.Index)
		return false
	}

	r.raftLog.restore(snapshot)
	r.progress = tracker.MakeProgressTracker(r.progress.MaxInflight)
	cfg, progress, err := confchange.Restore(confchange.Changer{
		Tracker:   r.progress,
		LastIndex: r.raftLog.lastIndex(),
	}, snapshot.Metadata.ConfState)
	if err != nil {
		panic(err)
	}
	if err = snapshot.Metadata.ConfState.Equivalent(r.switchToConfig(cfg, progress)); err != nil {
		panic(err)
	}

	pr := r.progress.Progress[r.id]
	pr.MaybeUpdate(pr.Next - 1)

	klog.Infof(fmt.Sprintf("%x [commit: %d, lastindex: %d, lastterm: %d] restored snapshot [index: %d, term: %d]",
		r.id, r.raftLog.committed, r.raftLog.lastIndex(), r.raftLog.lastTerm(), snapshot.Metadata.Index, snapshot.Metadata.Term))
	return true
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

	r.pendingConfIndex = 0
	r.uncommittedSize = 0
	//r.readOnly = newReadOnly(r.readOnly.option)
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

// INFO: @see RawNode.readyWithoutAccept()
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

// appliedCursor extracts from the Ready the highest index the client has
// applied (once the Ready is confirmed via Advance). If no information is
// contained in the Ready, returns zero.
func (ready Ready) appliedCursor() uint64 {
	if n := len(ready.CommittedEntries); n > 0 {
		return ready.CommittedEntries[n-1].Index
	}
	if index := ready.Snapshot.Metadata.Index; index > 0 {
		return index
	}
	return 0
}
