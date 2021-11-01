package master

import (
	"fmt"
	"strconv"
	"strings"

	"k8s-lx1036/k8s/storage/raft/proto"

	"k8s.io/klog/v2"
)

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
	peers               []PeerAddress
	peerAddrs           []string
	heartbeatPort       int64
	replicaPort         int64
	s3Endpoint          string
	region              string
}

func parsePeerAddr(peerAddr string) (id uint64, ip string, port uint64, err error) {
	peerStr := strings.Split(peerAddr, colonSplit)
	id, err = strconv.ParseUint(peerStr[0], 10, 64)
	if err != nil {
		return
	}
	port, err = strconv.ParseUint(peerStr[2], 10, 64)
	if err != nil {
		return
	}
	ip = peerStr[1]
	return
}

func (cfg *clusterConfig) parsePeers(peerStr string) error {
	peerArr := strings.Split(peerStr, commaSplit)
	cfg.peerAddrs = peerArr
	for _, peerAddr := range peerArr {
		id, ip, port, err := parsePeerAddr(peerAddr)
		if err != nil {
			return err
		}
		cfg.peers = append(cfg.peers, PeerAddress{
			Peer:          proto.Peer{ID: id},
			Address:       ip,
			HeartbeatPort: int(cfg.heartbeatPort),
			ReplicaPort:   int(cfg.replicaPort),
		})
		address := fmt.Sprintf("%v:%v", ip, port)
		klog.Infof("address: %s", address)
		AddrDatabase[id] = address
	}
	return nil
}
