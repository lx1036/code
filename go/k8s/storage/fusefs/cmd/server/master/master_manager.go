package master

import (
	"fmt"
	"strings"

	"github.com/tiglabs/raft/proto"

	"k8s.io/klog/v2"
)

// LeaderInfo represents the leader's information
type LeaderInfo struct {
	addr string //host:port
}

func (server *Server) handleLeaderChange(leader uint64) {
	if leader == 0 {
		klog.Error("action[handleLeaderChange] but no leader")
		return
	}

	oldLeaderAddr := server.leaderInfo.addr
	server.leaderInfo.addr = AddrDatabase[leader]
	klog.Infof("action[handleLeaderChange] change leader to [%v] ", server.leaderInfo.addr)
	server.reverseProxy = server.newReverseProxy()

	if server.id == leader {
		klog.Infof(server.clusterName, fmt.Sprintf("clusterID[%v] leader is changed to %v",
			server.clusterName, server.leaderInfo.addr))
		if oldLeaderAddr != server.leaderInfo.addr {
			//server.loadMetadata()
			server.metaReady = true
		}
		server.cluster.checkMetaNodeHeartbeat()
	} else {
		klog.Infof(server.clusterName, fmt.Sprintf("clusterID[%v] leader is changed to %v",
			server.clusterName, server.leaderInfo.addr))
		//server.clearMetadata()
		server.metaReady = false
	}
}

func (server *Server) handlePeerChange(confChange *proto.ConfChange) (err error) {
	var msg string
	addr := string(confChange.Context)
	switch confChange.Type {
	case proto.ConfAddNode:
		var arr []string
		if arr = strings.Split(addr, colonSplit); len(arr) < 2 {
			msg = fmt.Sprintf("action[handlePeerChange] clusterID[%v] nodeAddr[%v] is invalid", server.clusterName, addr)
			break
		}
		server.raftStore.AddNodeWithPort(confChange.Peer.ID, arr[0], int(server.config.heartbeatPort), int(server.config.replicaPort))
		AddrDatabase[confChange.Peer.ID] = string(confChange.Context)
		msg = fmt.Sprintf("clusterID[%v] peerID:%v,nodeAddr[%v] has been add", server.clusterName, confChange.Peer.ID, addr)
	case proto.ConfRemoveNode:
		server.raftStore.DeleteNode(confChange.Peer.ID)
		msg = fmt.Sprintf("clusterID[%v] peerID:%v,nodeAddr[%v] has been removed", server.clusterName, confChange.Peer.ID, addr)
	}
	klog.Infof(msg)
	return
}

func (server *Server) handleApplySnapshot() {
	server.fsm.restore()
	server.restoreIDAlloc()
	return
}

func (server *Server) restoreIDAlloc() {
	server.cluster.idAlloc.restore()
}
