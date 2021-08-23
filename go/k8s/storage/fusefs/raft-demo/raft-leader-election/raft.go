package main

import (
	"fmt"
	"k8s.io/klog"
	"math/rand"
	"sync"
	"time"
)

// INFO: https://github.com/eliben/raft/blob/master/part1/raft.go
//  https://eli.thegreenplace.net/2020/implementing-raft-part-1-elections/

type CMState int

const (
	Follower CMState = iota
	Candidate
	Leader
	Dead
)

func (s CMState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	case Dead:
		return "Dead"
	default:
		panic("unreachable")
	}
}

type LogEntry struct {
	Command interface{}
	Term    int
}

// ConsensusModule (CM) implements a single node of Raft consensus.
type ConsensusModule struct {
	// mu protects concurrent access to a CM.
	mu sync.Mutex

	// id is the server ID of this CM.
	// 当前node的编号
	id int
	// peerIds lists the IDs of our peers in the cluster.
	peerIds []int

	// Persistent Raft state on all servers
	// 当前选举周期
	currentTerm int
	// 投票给哪个node的编号
	votedFor int

	// Volatile Raft state on all servers
	// 表示当前node的角色，有 leader/follower
	state CMState
	// 上次参与选举时的时间
	electionResetEvent time.Time

	// server is the server containing this CM. It's used to issue RPC calls
	// to peers.
	server *Server
}

// runElectionTimer implements an election timer. It should be launched whenever
// we want to start a timer towards becoming a candidate in a new election.
//
// This function is blocking and should be launched in a separate goroutine;
// it's designed to work for a single (one-shot) election timer, as it exits
// whenever the CM state changes from follower/candidate or the term changes.
func (cm *ConsensusModule) runElectionTimer() {
	timeoutDuration := cm.electionTimeout()
	cm.mu.Lock()
	termStarted := cm.currentTerm
	cm.mu.Unlock()
	klog.Infof(fmt.Sprintf("election timer started (%v), term=%d", timeoutDuration, termStarted))

	// This loops until either:
	// - we discover the election timer is no longer needed, or
	// - the election timer expires and this CM becomes a candidate
	// In a follower, this typically keeps running in the background for the
	// duration of the CM's lifetime.
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		<-ticker.C

		cm.mu.Lock()
		// 当前node成为leader节点，就不需要选举了
		if cm.state != Candidate && cm.state != Follower {
			klog.Infof(fmt.Sprintf("in election timer state=%s, bailing out", cm.state))
			cm.mu.Unlock()
			return
		}

		if termStarted != cm.currentTerm {
			klog.Infof(fmt.Sprintf("in election timer term changed from %d to %d, bailing out", termStarted, cm.currentTerm))
			cm.mu.Unlock()
			return
		}

		// Start an election if we haven't heard from a leader or haven't voted for
		// someone for the duration of the timeout.
		// 超过了 timeoutDuration 之后，可以发起选举了
		if elapsed := time.Since(cm.electionResetEvent); elapsed >= timeoutDuration {
			cm.startElection()
			cm.mu.Unlock()
			return
		}

		cm.mu.Unlock()
	}
}

// See figure 2 in the paper.
type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// INFO: startElection starts a new election with this CM as a candidate.
//   Expects cm.mu to be locked.
//   发起选举，依次 rpc RequestVote 给 peers，
func (cm *ConsensusModule) startElection() {
	// INFO: 投票过程：
	// 1. follower递增自己的term
	// 2. follower将自己的状态变为candidate
	// 3. 投票给自己
	// 4. 向集群其它机器发起投票请求（RequestVote请求）
	cm.state = Candidate
	cm.currentTerm += 1 // 每一次选举，term 都会变化
	savedCurrentTerm := cm.currentTerm
	cm.electionResetEvent = time.Now()
	cm.votedFor = cm.id // 给自己投票

	klog.Infof(fmt.Sprintf("[startElection]becomes Candidate (currentTerm=%d)", savedCurrentTerm))

	// 自己已经先给自己一票
	votesReceived := 1

	// Send RequestVote RPCs to all other servers concurrently.
	for _, peerId := range cm.peerIds {
		go func(peerId int) {
			args := RequestVoteArgs{
				Term:        savedCurrentTerm,
				CandidateId: cm.id,
			}
			var reply RequestVoteReply
			// INFO: RequestVote 发起投票请求
			if err := cm.server.Call(peerId, "ConsensusModule.RequestVote", args, &reply); err == nil {
				// INFO: 这里使用lock，这样rpc调用之后，不能并发修改下面的逻辑
				//  这里的 lock 很重要，下面逻辑不容许并发

				// INFO: 这里由于有 lock 存在，所以 votesReceived 是依次累加的，比如
				//  reply=7,votesReceived=8;reply=1;votesReceived=8+1=9;reply=8,votesReceived=9+8=17

				cm.mu.Lock()
				defer cm.mu.Unlock()
				if cm.state != Candidate { // 选举结束了，当前节点要么是 leader 或者 follower 了
					klog.Infof(fmt.Sprintf("node state switch from %s to %s", Candidate, cm.state))
					return
				}

				// INFO: (1) 选举失败了
				if reply.Term > savedCurrentTerm {
					// INFO: 说明别人已经是leader了，自己的term过期了，得切换到最新的reply.Term，自己切换到follower
					cm.becomeFollower(reply.Term) // cm.state 应该是 follower
					return
				} else if reply.Term == savedCurrentTerm {
					if reply.VoteGranted {
						// INFO: 如果得到了一票
						votesReceived += 1
						if votesReceived*2 > len(cm.peerIds)+1 { // INFO: (2) 选举成功了，成为 leader
							klog.Infof(fmt.Sprintf("wins election with %d votes", votesReceived))
							cm.startLeader() // cm.state 应该是 leader
							return
						}
					}
				}
			}
		}(peerId)
	}
}

// INFO: rpc 触发的，node 返回投票结果
func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.state == Dead {
		return nil
	}

	if args.Term > cm.currentTerm {
		klog.Infof(fmt.Sprintf("[RequestVote]term is outdated"))
		cm.becomeFollower(args.Term)
	}

	if cm.currentTerm == args.Term && (cm.votedFor == -1 || cm.votedFor == args.CandidateId) {
		reply.VoteGranted = true
		cm.votedFor = args.CandidateId
		cm.electionResetEvent = time.Now()
	} else {
		reply.VoteGranted = false
	}

	reply.Term = cm.currentTerm
	klog.Infof(fmt.Sprintf("[RequestVote]reply: %+v", reply))
	return nil
}

// becomeFollower makes cm a follower and resets its state.
// Expects cm.mu to be locked.
func (cm *ConsensusModule) becomeFollower(term int) {
	klog.Infof(fmt.Sprintf("becomes %s with term=%d", Follower, term))
	cm.state = Follower
	cm.currentTerm = term
	cm.votedFor = -1
	cm.electionResetEvent = time.Now()

	// INFO: 为何又要再次发起选举?
	go cm.runElectionTimer()
}

// startLeader switches cm into a leader state and begins process of heartbeats.
// Expects cm.mu to be locked.
func (cm *ConsensusModule) startLeader() {
	cm.state = Leader
	klog.Infof(fmt.Sprintf("becomes %s with term=%d", Leader, cm.currentTerm))

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		// INFO: 周期发心跳
		for {
			cm.leaderSendHeartbeats()
			<-ticker.C

			// INFO: 如果不是 leader 则不用发心跳了，为何加上这个逻辑?

			// INFO: Update: leaderSendHeartbeats() 会更新state成为 Follower，所以需要检查
			cm.mu.Lock()
			if cm.state != Leader {
				cm.mu.Unlock()
				return
			}
			cm.mu.Unlock()
		}
	}()
}

// See figure 2 in the paper.
type AppendEntriesArgs struct {
	Term     int
	LeaderId int

	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}
type AppendEntriesReply struct {
	Term    int
	Success bool
}

// leaderSendHeartbeats sends a round of heartbeats to all peers, collects their
// replies and adjusts cm's state.
func (cm *ConsensusModule) leaderSendHeartbeats() {
	cm.mu.Lock()
	savedCurrentTerm := cm.currentTerm
	cm.mu.Unlock()

	for _, peerId := range cm.peerIds {
		args := AppendEntriesArgs{
			Term:     savedCurrentTerm,
			LeaderId: cm.id,
		}

		go func(peerId int) {
			var reply AppendEntriesReply
			if err := cm.server.Call(peerId, "ConsensusModule.AppendEntries", args, &reply); err == nil {
				cm.mu.Lock()
				defer cm.mu.Unlock()
				if reply.Term > savedCurrentTerm {
					klog.Infof(fmt.Sprintf("current term %d is outdated, newest term is %d, switch leader to follower", savedCurrentTerm, reply.Term))
					cm.becomeFollower(reply.Term)
					return
				}
			}
		}(peerId)
	}
}

func (cm *ConsensusModule) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.state == Dead {
		return nil
	}

	if args.Term > cm.currentTerm {
		klog.Infof(fmt.Sprintf("[AppendEntries]term is outdated"))
		cm.becomeFollower(args.Term)
	}

	reply.Success = false
	if args.Term == cm.currentTerm {
		if cm.state != Follower {
			cm.becomeFollower(args.Term)
		}
		cm.electionResetEvent = time.Now()
		reply.Success = true
	}

	reply.Term = cm.currentTerm
	klog.Infof(fmt.Sprintf("[AppendEntries]reply: %+v", reply))
	return nil
}

// electionTimeout generates a pseudo-random election timeout duration.
func (cm *ConsensusModule) electionTimeout() time.Duration {
	// raft 论文里建议是 150-300 ms 之间，并且每次都是随机的，为了减少 split votes 出现，导致没有leader选出来
	/*
		INFO:
			Raft uses randomized election timeouts to ensure that split votes are rare and that they are resolved quickly.
			To prevent split votes in the first place, election timeouts are chosen randomly from a fixed interval (e.g., 150–300 ms).
			This spreads out the servers so that in most cases only a single server will time out;
			it wins the election and sends heartbeats before any other servers time out.
			The same mechanism is used to handle split votes.
			Each candidate restarts its randomized election timeout at the start of an election,
			and it waits for that timeout to elapse before starting the next election;
			this reduces the likelihood of another split vote in the new election.
	*/
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}
