package myraft

import (
	"fmt"
	"k8s.io/klog/v2"
)

type Processor interface {
	// HandleEvent external event and application event
	HandleEvent(evt *raftEvent) (interface{}, error)
	Init() error
	Loop()
	Stop() bool
}

type HandlerFunc func(*AppendLogRequest) (*AppendLogResponse, error)
type FollowerProcessor struct {
	handlers map[eventName]HandlerFunc

	state *RaftState

	raftLog RaftLog

	notifyC chan *raftEvent
}

func (processor *FollowerProcessor) HandleEvent(evt *raftEvent) (interface{}, error) {
	panic("implement me")
}

func (processor *FollowerProcessor) Init() error {

	processor.handlers = map[eventName]HandlerFunc{
		EventExtAppendLogRequest: processor.handleAppend,
	}

	return nil
}

func (processor *FollowerProcessor) handleAppend(request *AppendLogRequest) (*AppendLogResponse, error) {

	appendLogResponse := &AppendLogResponse{}
	// request term 不能比当前 raft state term 小
	if request.Term < processor.state.CurrentTerm {
		appendLogResponse.Term = processor.state.CurrentTerm
		appendLogResponse.Success = false
		appendLogResponse.MatchIndex = -1
		return appendLogResponse, nil
	}

	if request.Term > processor.state.CurrentTerm {
		processor.state.CurrentTerm = request.Term
		appendLogResponse.Term = processor.state.CurrentTerm
		// request.Term 大于自己的 term，说明自己term过时了，需要无条件切换到 Follower
		if processor.state.Role == Candidate || processor.state.Role == Leader {
			processor.notifySwitch(Follower)
			// 重置raft state
			processor.state.ResetVote()
		}

		processor.state.SaveState()
	}

	// 根据prevIndex查询本地prev log，论文里这一步是必要的，是 Safety log match 这一步逻辑
	logIterm := processor.raftLog.Index(request.PrevLogIndex)
	if logIterm.Term != request.PrevLogTerm || logIterm.Index != request.PrevLogIndex {
		appendLogResponse.Success = false
		appendLogResponse.MatchIndex = 0
		return appendLogResponse, nil
	}

	if request.LogItems != nil {
		lastIndex, err := processor.raftLog.AppendLog(request.LogItems, request.PrevLogIndex+1)
		if err != nil {
			appendLogResponse.Success = false
			appendLogResponse.MatchIndex = -1
			return appendLogResponse, nil
		}
		klog.Infof(fmt.Sprintf("[AppendLog]last index %d", lastIndex))
	}

	appendLogResponse.Success = true
	appendLogResponse.MatchIndex = logIterm.Index + int64(len(request.LogItems))
	// leader commitIndex 不应该小于当前 raft state commitIndex
	if request.LeaderCommit < processor.state.CommitIndex {
		klog.Fatalf(fmt.Sprintf("[AppendLog]leader commit %d < follower commit %d",
			request.LeaderCommit, processor.state.CommitIndex))
	}
	if request.LeaderCommit != processor.state.CommitIndex {
		processor.state.CommitIndex = request.LeaderCommit
		processor.notify(newEvent(EventNotifyApply, processor.state.CommitIndex, nil))
	}

	return appendLogResponse, nil
}

func (processor *FollowerProcessor) notify(event *raftEvent) interface{} {
	processor.notifyC <- event
	if event.resc != nil {
		return <-event.resc
	}

	return nil
}

func (processor *FollowerProcessor) Loop() {
	panic("implement me")
}

func (processor *FollowerProcessor) Stop() bool {
	panic("implement me")
}

type CandidateProcessor struct {
	FollowerProcessor

	ConnectionManager *PeerConnectionManager
}

type LeaderProcessor struct {
	CandidateProcessor

	syncC chan interface{}
}

func NewProcessor(role Role, processor Processor) Processor {

}
