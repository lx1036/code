package meta

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"k8s-lx1036/k8s/storage/fusefs/pkg/raftstore"
)

// StartRaftServer initializes the address resolver and the raftStore server instance.
func (m *MetaNode) startRaftServer() (err error) {
	if _, err = os.Stat(m.raftDir); err != nil {
		if err = os.MkdirAll(m.raftDir, 0755); err != nil {
			err = fmt.Errorf("create raft server dir: %v", err)
			return
		}
	}

	heartbeatPort, _ := strconv.Atoi(m.raftHeartbeatPort)
	replicaPort, _ := strconv.Atoi(m.raftReplicatePort)
	raftConf := &raftstore.Config{
		NodeID:            m.nodeId,
		RaftPath:          m.raftDir,
		IPAddr:            m.localAddr,
		HeartbeatPort:     heartbeatPort,
		ReplicaPort:       replicaPort,
		NumOfLogsToRetain: 2000000,
	}
	m.raftStore, err = raftstore.NewRaftStore(raftConf)
	if err != nil {
		err = errors.New(fmt.Sprintf("new raftStore: %s", err.Error()))
	}

	return
}

func (m *MetaNode) stopRaftServer() {
	if m.raftStore != nil {
		m.raftStore.Stop()
	}
}
