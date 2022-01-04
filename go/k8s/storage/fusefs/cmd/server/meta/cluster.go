package meta

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"

	"k8s.io/klog/v2"
)

const partitionPrefix = "partition_"

// MetadataManagerConfig defines the configures in the metadata manager.
type MetadataManagerConfig struct {
	NodeID    uint64
	StoreDir  string
	RaftStore *raftstore.RaftStore
}

type Cluster struct {
	sync.RWMutex

	nodeId              uint64
	storeDir            string
	localPartitionCount int

	partitions map[uint64]*PartitionFSM // Key: metaRangeId, Val: partition.MetaPartitionFSM
	raftStore  *raftstore.RaftStore

	connPool *util.ConnectPool
	state    uint32
}

func NewCluster(conf MetadataManagerConfig) *Cluster {
	return &Cluster{
		nodeId:     conf.NodeID,
		storeDir:   conf.StoreDir,
		raftStore:  conf.RaftStore,
		partitions: make(map[uint64]*PartitionFSM),
		connPool:   util.NewConnectPool(),
	}
}

// Start INFO: 从 metadataDir 目录中加载本地已有的 partitions
func (cluster *Cluster) Start() error {
	// Check metadataDir directory
	fileInfo, err := os.Stat(cluster.storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(cluster.storeDir, 0755)
		}
		if err != nil {
			return err
		}
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("storeDir must be directory")
	}

	// scan the data directory
	var wg sync.WaitGroup
	dirEntryList, _ := ioutil.ReadDir(cluster.storeDir)
	for _, dirEntry := range dirEntryList {
		// 必须是 partition_xxx 目录
		fileName := dirEntry.Name()
		if !dirEntry.IsDir() || !strings.HasPrefix(fileName, partitionPrefix) {
			continue
		}

		wg.Add(1)
		cluster.localPartitionCount++
		go func(fileName string) {
			defer wg.Done()

			partitionId := fileName[len(partitionPrefix):]
			id, err := strconv.ParseUint(partitionId, 10, 64)
			if err != nil {
				klog.Warningf("ignore path: %s,not partition", partitionId)
				return
			}
			partitionConfig := &PartitionConfig{
				NodeId:    cluster.nodeId,
				RaftStore: cluster.raftStore,
				StoreDir:  path.Join(cluster.storeDir, fileName),
				ConnPool:  cluster.connPool,
				AfterStop: func() {
					cluster.detachPartition(id)
				},
			}
			// /data/metanode/partition/partition_1/snapshot/
			snapshotDir := path.Join(partitionConfig.StoreDir, SnapshotDir)
			// 如果没有/snapshot目录，就从 /.snapshot_backup 目录rename到 /snapshot，如果 /.snapshot_backup 存在的话
			if _, errload := os.Stat(snapshotDir); errload != nil {
				backupDir := path.Join(partitionConfig.StoreDir, snapshotBackup)
				if _, errload = os.Stat(backupDir); errload == nil {
					if errload = os.Rename(backupDir, snapshotDir); errload != nil {
						errload = fmt.Errorf("fail recover backup snapshot %s with err %v", snapshotDir, errload)
						return
					}
				}
			}

			if err = cluster.attachPartition(id, NewPartitionFSM(partitionConfig)); err != nil {
				klog.Errorf(fmt.Sprintf("load partition id=%d failed: %v.", id, err))
			}
		}(fileName)
	}
	wg.Wait()

	return nil
}

// INFO: 启动每一个 partition，并缓存已经启动的 partition
func (cluster *Cluster) attachPartition(id uint64, partition *PartitionFSM) error {
	if err := partition.Start(); err != nil {
		klog.Errorf("finish load partition.MetaPartitionFSM %v error %v", id, err)
		return err
	}

	cluster.Lock()
	defer cluster.Unlock()
	cluster.partitions[id] = partition

	return nil
}

func (cluster *Cluster) detachPartition(id uint64) error {
	cluster.Lock()
	defer cluster.Unlock()

	if _, has := cluster.partitions[id]; has {
		delete(cluster.partitions, id)
		return nil
	}

	return fmt.Errorf("unknown partition: %d", id)
}

func (cluster *Cluster) getPartition(id uint64) (*PartitionFSM, error) {
	cluster.RLock()
	defer cluster.RUnlock()

	mp, ok := cluster.partitions[id]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("unknown meta partition: %d", id))
	}
	return mp, nil
}

func (cluster *Cluster) GetPartition(id uint64) (*PartitionFSM, error) {
	return cluster.getPartition(id)
}

func (cluster *Cluster) LoadStat() string {
	return fmt.Sprintf("state total/loaded : %d/%d", cluster.localPartitionCount, len(cluster.partitions))
}

////////////////////////////TCP///////////////////////////////////

func (cluster *Cluster) HandleMetadataOperation(conn net.Conn, p *proto.Packet, remoteAddr string) error {
	var err error
	switch p.Opcode {
	// inode
	case proto.OpMetaCreateInode:
		err = cluster.opCreateInode(conn, p, remoteAddr)
	case proto.OpMetaLinkInode:
		err = cluster.opCreateInodeLink(conn, p, remoteAddr)
	case proto.OpMetaUnlinkInode:
		err = cluster.opUnlinkInode(conn, p, remoteAddr)
	case proto.OpMetaBatchUnlinkInode:
		err = cluster.opBatchUnlinkInode(conn, p, remoteAddr)

	case proto.OpCreateMetaPartition:
		err = cluster.opCreateMetaPartition(conn, p, remoteAddr)

	//case proto.OpMetaInodeGet:
	//	err = cluster.opMetaInodeGet(conn, p, remoteAddr)
	//case proto.OpMetaEvictInode:
	//	err = cluster.opMetaEvictInode(conn, p, remoteAddr)
	//case proto.OpMetaSetattr:
	//	err = cluster.opSetAttr(conn, p, remoteAddr)
	//case proto.OpMetaCreateDentry:
	//	err = cluster.opCreateDentry(conn, p, remoteAddr)
	//case proto.OpMetaDeleteDentry:
	//	err = cluster.opDeleteDentry(conn, p, remoteAddr)
	//case proto.OpMetaUpdateDentry:
	//	err = cluster.opUpdateDentry(conn, p, remoteAddr)
	//case proto.OpMetaReadDir:
	//	err = cluster.opReadDir(conn, p, remoteAddr)
	//case proto.OpCreateMetaPartition:
	//	err = cluster.opCreateMetaPartition(conn, p, remoteAddr)
	//case proto.OpMetaNodeHeartbeat:
	//	err = cluster.opMasterHeartbeat(conn, p, remoteAddr)
	//case proto.OpMetaLookup:
	//	err = cluster.opMetaLookup(conn, p, remoteAddr)
	//case proto.OpMetaLookupName:
	//	err = cluster.opMetaLookupName(conn, p, remoteAddr)
	//case proto.OpDeleteMetaPartition:
	//	err = cluster.opDeleteMetaPartition(conn, p, remoteAddr)
	//case proto.OpUpdateMetaPartition:
	//	err = cluster.opUpdateMetaPartition(conn, p, remoteAddr)
	//case proto.OpLoadMetaPartition:
	//	err = cluster.opLoadMetaPartition(conn, p, remoteAddr)
	//case proto.OpDecommissionMetaPartition:
	//	err = cluster.opDecommissionMetaPartition(conn, p, remoteAddr)
	//case proto.OpAddMetaPartitionRaftMember:
	//	err = cluster.opAddMetaPartitionRaftMember(conn, p, remoteAddr)
	//case proto.OpRemoveMetaPartitionRaftMember:
	//	err = cluster.opRemoveMetaPartitionRaftMember(conn, p, remoteAddr)
	//case proto.OpMetaPartitionTryToLeader:
	//	err = cluster.opMetaPartitionTryToLeader(conn, p, remoteAddr)
	//case proto.OpMetaBatchInodeGet:
	//	err = cluster.opMetaBatchInodeGet(conn, p, remoteAddr)
	default:
		err = fmt.Errorf("%s unknown Opcode: %d, reqId: %d", remoteAddr,
			p.Opcode, p.GetReqID())
	}
	if err != nil {
		err = fmt.Errorf("%s [%s] req: %d - %v", remoteAddr, p.GetOpMsg(), p.GetReqID(), err)
	}

	return err
}

func (cluster *Cluster) opCreateInode(conn net.Conn, p *proto.Packet, remoteAddr string) error {
	req := &proto.CreateInodeRequest{}
	if err := json.Unmarshal(p.Data, req); err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return err
	}

	partition, err := cluster.getPartition(req.PartitionID)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return err
	}

	// INFO: 如果不是leader，可以 proxy request to leader，然后直接返回
	if cluster.serveProxy(conn, partition, p) {
		return nil
	}

	err = partition.CreateInode(req, p)
	p.WriteToConn(conn)
	klog.Infof(fmt.Sprintf("[opCreateInode]%s req: %v, resp: %v, body: %s", remoteAddr, *req, p.GetResultMsg(), p.Data))

	return err
}

func (cluster *Cluster) opCreateInodeLink(conn net.Conn, p *proto.Packet, remoteAddr string) error {
	req := &proto.CreateInodeLinkRequest{}
	if err := json.Unmarshal(p.Data, req); err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return err
	}

	partition, err := cluster.getPartition(req.PartitionID)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return err
	}

	// INFO: 如果不是leader，可以 proxy request to leader，然后直接返回
	if cluster.serveProxy(conn, partition, p) {
		return nil
	}

	err = partition.CreateInodeLink(req, p)
	p.WriteToConn(conn)
	klog.Infof(fmt.Sprintf("[opCreateInodeLink]%s req: %v, resp: %v, body: %s", remoteAddr, *req, p.GetResultMsg(), p.Data))

	return err
}

func (cluster *Cluster) opUnlinkInode(conn net.Conn, p *proto.Packet, remoteAddr string) error {
	req := &proto.UnlinkInodeRequest{}
	if err := json.Unmarshal(p.Data, req); err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return err
	}

	partition, err := cluster.getPartition(req.PartitionID)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return err
	}

	// INFO: 如果不是leader，可以 proxy request to leader，然后直接返回
	if cluster.serveProxy(conn, partition, p) {
		return nil
	}

	err = partition.UnlinkInode(req, p)
	p.WriteToConn(conn)
	klog.Infof(fmt.Sprintf("[opUnlinkInode]%s req: %v, resp: %v, body: %s", remoteAddr, *req, p.GetResultMsg(), p.Data))

	return err
}

func (cluster *Cluster) opBatchUnlinkInode(conn net.Conn, p *proto.Packet, remoteAddr string) error {
	req := &proto.BatchUnlinkInodeRequest{}
	if err := json.Unmarshal(p.Data, req); err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return err
	}

	partition, err := cluster.getPartition(req.PartitionID)
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return err
	}

	// INFO: 如果不是leader，可以 proxy request to leader，然后直接返回
	if cluster.serveProxy(conn, partition, p) {
		return nil
	}

	err = partition.BatchUnlinkInode(req, p)
	p.WriteToConn(conn)
	klog.Infof(fmt.Sprintf("[opUnlinkInode]%s req: %v, resp: %v, body: %s", remoteAddr, *req, p.GetResultMsg(), p.Data))

	return err
}

func (cluster *Cluster) opCreateMetaPartition(conn net.Conn, p *proto.Packet, remoteAddr string) (err error) {
	return nil
}

func (cluster *Cluster) serveProxy(conn net.Conn, partition *PartitionFSM, p *proto.Packet) bool {
	leaderAddr, isLeader := partition.IsLeader()
	if isLeader {
		return false
	}
	if len(leaderAddr) == 0 {
		p.PacketErrorWithBody(proto.OpErr, []byte("no leader"))
		p.WriteToConn(conn)
		return true
	}

	leaderConn, _ := net.Dial("tcp", leaderAddr)
	if err := p.WriteToConn(leaderConn); err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return true
	}
	if err := p.ReadFromConn(leaderConn, proto.ReadDeadlineTime); err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		p.WriteToConn(conn)
		return true
	}

	// proxy leaderConn to client conn
	p.WriteToConn(conn)
	return true
}

// Stop INFO: stop 每一个 partition
func (cluster *Cluster) Stop() {
	for _, partition := range cluster.partitions {
		partition.Stop()
	}
}
