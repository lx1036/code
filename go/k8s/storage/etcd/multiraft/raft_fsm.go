package multiraft

import (
	"fmt"
	"k8s.io/klog/v2"
	"math/rand"

	"k8s-lx1036/k8s/storage/raft/proto"
)

type (
	fsmState     byte
	replicaState byte
)

const (
	stateFollower    fsmState = 0
	stateCandidate            = 1
	stateLeader               = 2
	stateElectionACK          = 3

	replicaStateProbe     replicaState = 0
	replicaStateReplicate              = 1
	replicaStateSnapshot               = 2
)

func (st fsmState) String() string {
	switch st {
	case 0:
		return "StateFollower"
	case 1:
		return "StateCandidate"
	case 2:
		return "StateLeader"
	case 3:
		return "StateElectionACK"
	}
	return ""
}

func (st replicaState) String() string {
	switch st {
	case 1:
		return "ReplicaStateReplicate"
	case 2:
		return "ReplicaStateSnapshot"
	default:
		return "ReplicaStateProbe"
	}
}

type stepFunc func(m *proto.Message)

// raft state machine
type raftFsm struct {
	nodeConfig *NodeConfig

	id               uint64
	term             uint64
	vote             uint64
	leader           uint64
	electionElapsed  int
	heartbeatElapsed int
	// randElectionTick is a random number between[electiontimetick, 2 * electiontimetick - 1].
	// It gets reset when raft changes its state to follower or candidate.
	randElectionTick int
	// New configuration is ignored if there exists unapplied configuration.
	pendingConf bool
	state       fsmState
	sm          StateMachine
	raftLog     *RaftLog
	rand        *rand.Rand
	votes       map[uint64]bool
	acks        map[uint64]bool

	replicas map[uint64]*Replica

	// INFO: 主要函数，根据 Leader/Follower 不同，函数不同
	step stepFunc

	//readOnly    *readOnly
	msgs []*proto.Message
	tick func()
}

func NewRaftFsm(nodeConfig *NodeConfig, raftConfig *RaftConfig) (*raftFsm, error) {
	raftLog, err := NewRaftLog(raftConfig.Storage)
	if err != nil {
		return nil, err
	}
	/*hs, err := raftConfig.Storage.InitialState()
	if err != nil {
		return nil, err
	}*/

	r := &raftFsm{
		id:         raftConfig.ID,
		sm:         raftConfig.StateMachine,
		nodeConfig: nodeConfig,
		leader:     NoLeader,
		raftLog:    raftLog,
		replicas:   make(map[uint64]*Replica),
		//readOnly: newReadOnly(raftConfig.ID, config.ReadOnlyOption),
	}
	r.rand = rand.New(rand.NewSource(int64(nodeConfig.NodeID + r.id)))
	for _, p := range raftConfig.Peers {
		r.replicas[p.ID] = NewReplica(p, 0)
	}
	/*if !hs.IsEmpty() {
		if raftConfig.Applied > r.raftLog.LastIndex() {
			raftConfig.Applied = r.raftLog.LastIndex()
		}
		if hs.Commit > r.raftLog.LastIndex() {
			hs.Commit = r.raftLog.LastIndex()
		}
		if err := r.loadState(hs); err != nil {
			return nil, err
		}
	}*/

	klog.Info(fmt.Sprintf("newRaft[%v] [commit: %d, applied: %d, lastindex: %d]", r.id, raftLog.Committed, raftConfig.Applied, raftLog.LastIndex()))

	/*if raftConfig.Applied > 0 {
		lasti := raftLog.LastIndex()
		if lasti == 0 {
			// If there is application data but no raft log, then restore to initial state.
			raftLog.Committed = 0
			raftConfig.Applied = 0
		} else if lasti < raftConfig.Applied {
			// If lastIndex<appliedIndex, then the log as the standard.
			raftLog.Committed = lasti
			raftConfig.Applied = lasti
		} else if raftLog.Committed < raftConfig.Applied {
			raftLog.Committed = raftConfig.Applied
		}
		raftLog.AppliedTo(raftConfig.Applied)
	}

	// recover committed
	if err := r.recoverCommit(); err != nil {
		return nil, err
	}
	if raftConfig.Leader == config.NodeID {
		if raftConfig.Term != 0 && r.term <= raftConfig.Term {
			r.term = raftConfig.Term
			r.state = stateLeader
			r.becomeLeader()
			r.bcastAppend()
		} else {
			r.becomeFollower(r.term, NoLeader)
		}
	} else {
		if raftConfig.Leader == NoLeader {
			r.becomeFollower(r.term, NoLeader)
		} else {
			r.becomeFollower(raftConfig.Term, raftConfig.Leader)
		}
	}*/

	//peerStrs := make([]string, 0)
	/*for _, p := range r.peers() {
		peerStrs = append(peerStrs, fmt.Sprintf("%v", p.String()))
	}*/
	//klog.Infof("newRaft[%v] [peers: [%s], term: %d, commit: %d, applied: %d, lastindex: %d, lastterm: %d]", r.id, strings.Join(peerStrs, ","), r.term, r.raftLog.Committed, r.raftLog.Applied, r.raftLog.LastIndex(), r.raftLog.LastTerm())

	if raftConfig.Leader == nodeConfig.NodeID {
		if raftConfig.Term != 0 && r.term <= raftConfig.Term {
			r.term = raftConfig.Term
			r.state = stateLeader
			r.becomeLeader()
			r.bcastAppend()
		} else {
			//r.becomeFollower(r.term, NoLeader)
		}
	} else {
		if raftConfig.Leader == NoLeader {
			//r.becomeFollower(r.term, NoLeader)
		} else {
			//r.becomeFollower(raftConfig.Term, raftConfig.Leader)
		}
	}

	//go r.doRandomSeed()

	return r, nil
}

// Step INFO: 根据不同类型 message 推动状态机运转
func (r *raftFsm) Step(message *proto.Message) {

	switch {
	case message.Term == 0:
		// local message
	case message.Term > r.term:

	case message.Term < r.term:
		klog.Infof(fmt.Sprintf("[raftFsm Step] %x [term: %d] ignored a %s message with lower term from %x [term: %d]",
			r.id, r.term, message.Type, message.From, message.Term))
		return

	}

	r.step(message)
}
