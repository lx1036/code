package master

import (
	"encoding/json"
	"strconv"
)

// RaftCmd defines the Raft commands.
type RaftCmd struct {
	Op uint32 `json:"op"`
	K  string `json:"k"`
	V  []byte `json:"v"`
}

type nodeSetValue struct {
	ID          uint64
	Capacity    int
	MetaNodeLen int
}

func newNodeSetValue(nset *nodeSet) *nodeSetValue {
	return &nodeSetValue{
		ID:          nset.ID,
		Capacity:    nset.Capacity,
		MetaNodeLen: nset.metaNodeLen,
	}
}

type metaNodeValue struct {
	ID        uint64
	NodeSetID uint64
	Addr      string
}

func newMetaNodeValue(metaNode *MetaNode) *metaNodeValue {
	return &metaNodeValue{
		ID:        metaNode.ID,
		NodeSetID: metaNode.NodeSetID,
		Addr:      metaNode.Addr,
	}
}

// key=#s#id
func (cluster *Cluster) syncAddNodeSet(nodeset *nodeSet) error {
	return cluster.putNodeSetInfo(opSyncAddNodeSet, nodeset)
}

func (cluster *Cluster) putNodeSetInfo(opType uint32, nset *nodeSet) error {
	raftCmd := &RaftCmd{
		Op: opType,
		K:  nodeSetPrefix + strconv.FormatUint(nset.ID, 10),
	}

	var err error
	nsv := newNodeSetValue(nset)
	raftCmd.V, err = json.Marshal(nsv)
	if err != nil {
		return err
	}

	cmd, err := json.Marshal(raftCmd)
	if err != nil {
		return err
	}

	if _, err = cluster.partition.Submit(cmd); err != nil {
		return err
	}

	return nil
}

func (cluster *Cluster) putMetaNodeInfo(opType uint32, metaNode *MetaNode) error {
	raftCmd := &RaftCmd{
		Op: opType,
		K:  metaNodePrefix + strconv.FormatUint(metaNode.ID, 10) + keySeparator + metaNode.Addr,
	}

	var err error
	metaNodeValue := newMetaNodeValue(metaNode)
	raftCmd.V, err = json.Marshal(metaNodeValue)
	if err != nil {
		return err
	}

	cmd, err := json.Marshal(raftCmd)
	if err != nil {
		return err
	}

	if _, err = cluster.partition.Submit(cmd); err != nil {
		return err
	}

	return nil
}
