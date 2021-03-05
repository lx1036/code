package master

import "k8s-lx1036/k8s/storage/dfs/pkg/raftstore"

//config key
const (
	colonSplit             = ":"
	commaSplit             = ","
	cfgPeers               = "peers"
	nodeSetCapacity        = "nodeSetCap"
	cfgMetaNodeReservedMem = "metaNodeReservedMem"
	heartbeatPortKey       = "heartbeatPort"
	replicaPortKey         = "replicaPort"
)

type clusterConfig struct {
	NodeTimeOutSec      int64
	metaNodeReservedMem uint64
	nodeSetCapacity     int
	MetaNodeThreshold   float32
	peers               []raftstore.PeerAddress
	peerAddrs           []string
	heartbeatPort       int64
	replicaPort         int64
	s3Endpoint          string
	region              string
}
