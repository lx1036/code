package raft

import (
	"k8s.io/klog/v2"
	"testing"
)

func TestCreateRaft(test *testing.T) {

	serverConfig := &Config{
		TransportConfig: TransportConfig{},
		NodeID:          0,
		TickInterval:    0,
		HeartbeatTick:   0,
		ElectionTick:    0,
		MaxSizePerMsg:   0,
		MaxInflightMsgs: 0,
		ReqBufferSize:   0,
		AppBufferSize:   0,
		RetainLogs:      0,
		LeaseCheck:      false,
		ReadOnlyOption:  0,
		transport:       nil,
	}
	server, err := NewRaftServer(serverConfig)
	if err != nil {
		klog.Fatal(err)
	}

	raftConfg := &RaftConfig{
		ID:           0,
		Term:         0,
		Leader:       0,
		Applied:      0,
		Peers:        nil,
		Storage:      nil,
		StateMachine: nil,
	}
	err = server.CreateRaft(raftConfg)
	if err != nil {
		klog.Fatal(err)
	}

	go server.run()

	<-server.stopc
}
