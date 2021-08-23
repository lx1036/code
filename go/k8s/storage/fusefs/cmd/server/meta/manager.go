package meta

import (
	"errors"
	"fmt"
	"io/ioutil"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s-lx1036/k8s/storage/fusefs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"k8s.io/klog/v2"
)

const partitionPrefix = "partition_"

// MetadataManagerConfig defines the configures in the metadata manager.
type MetadataManagerConfig struct {
	NodeID    uint64
	RootDir   string
	RaftStore raftstore.RaftStore
}

type metadataManager struct {
	nodeId              uint64
	rootDir             string
	raftStore           raftstore.RaftStore
	connPool            *util.ConnectPool
	state               uint32
	mu                  sync.RWMutex
	partitions          map[uint64]MetaPartition // Key: metaRangeId, Val: metaPartition
	localPartitionCount int
}

// INFO: 从 metadataDir 目录中加载本地已有的 partitions
func (m *metadataManager) Start() error {
	// Check metadataDir directory
	fileInfo, err := os.Stat(m.rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(m.rootDir, 0755)
		}
		if err != nil {
			return err
		}
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("metadataDir must be directory")
	}

	// scan the data directory
	dirEntryList, err := ioutil.ReadDir(m.rootDir)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, dirEntry := range dirEntryList {
		// 必须是 partition_xxx 目录
		if !dirEntry.IsDir() || !strings.HasPrefix(dirEntry.Name(), partitionPrefix) {
			continue
		}

		m.localPartitionCount++
		wg.Add(1)
		go func(fileName string) {
			var errload error
			defer func() {
				if r := recover(); r != nil {
					klog.Errorf("loadPartitions partition: %s, "+
						"error: %s, failed: %v", fileName, errload, r)
					panic(r)
				}
				if errload != nil {
					klog.Errorf("loadPartitions partition: %s, "+
						"error: %s", fileName, errload)
					panic(errload)
				}
			}()
			defer wg.Done()

			if len(fileName) < 10 {
				klog.Warningf("ignore unknown partition dir: %s", fileName)
				return
			}
			partitionId := fileName[len(partitionPrefix):]
			id, err := strconv.ParseUint(partitionId, 10, 64)
			if err != nil {
				klog.Warningf("ignore path: %s,not partition", partitionId)
				return
			}
			partitionConfig := &MetaPartitionConfig{
				NodeId:    m.nodeId,
				RaftStore: m.raftStore,
				RootDir:   path.Join(m.rootDir, fileName),
				ConnPool:  m.connPool,
				AfterStop: func() {
					m.detachPartition(id)
				},
			}
			// check snapshot dir or backup
			snapshotDir := path.Join(partitionConfig.RootDir, snapshotDir) // data/metanode/partition/partition_1/snapshot/
			// 如果没有/snapshot目录，就从 /.snapshot_backup 目录rename到 /snapshot，如果 /.snapshot_backup 存在的话
			if _, errload = os.Stat(snapshotDir); errload != nil {
				backupDir := path.Join(partitionConfig.RootDir, snapshotBackup)
				if _, errload = os.Stat(backupDir); errload == nil {
					if errload = os.Rename(backupDir, snapshotDir); errload != nil {
						errload = fmt.Errorf("fail recover backup snapshot %s with err %v", snapshotDir, errload)
						return
					}
				}
				errload = nil
			}

			errload = m.attachPartition(id, NewMetaPartition(partitionConfig, m))
			if errload != nil {
				klog.Errorf("load partition id=%d failed: %s.", id, errload.Error())
			}
		}(dirEntry.Name())
	}
	wg.Wait()

	return nil
}

// INFO: 启动每一个 partition，并缓存已经启动的 partition
func (m *metadataManager) attachPartition(id uint64, partition MetaPartition) error {
	if err := partition.Start(); err != nil {
		klog.Errorf("finish load metaPartition %v error %v", id, err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.partitions[id] = partition

	return nil
}

func (m *metadataManager) detachPartition(id uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, has := m.partitions[id]; has {
		delete(m.partitions, id)
		return nil
	}

	return fmt.Errorf("unknown partition: %d", id)
}

// INFO: stop 每一个 partition
func (m *metadataManager) Stop() {
	if m.partitions != nil {
		for _, partition := range m.partitions {
			partition.Stop()
		}
	}
}

func (m *metadataManager) HandleMetadataOperation(conn net.Conn, p *proto.Packet, remoteAddr string) error {
	var err error
	switch p.Opcode {
	case proto.OpMetaCreateInode:
		err = m.opCreateInode(conn, p, remoteAddr)
	//case proto.OpMetaLinkInode:
	//	err = m.opMetaLinkInode(conn, p, remoteAddr)
	//case proto.OpMetaUnlinkInode:
	//	err = m.opMetaUnlinkInode(conn, p, remoteAddr)
	//case proto.OpMetaInodeGet:
	//	err = m.opMetaInodeGet(conn, p, remoteAddr)
	//case proto.OpMetaEvictInode:
	//	err = m.opMetaEvictInode(conn, p, remoteAddr)
	//case proto.OpMetaSetattr:
	//	err = m.opSetAttr(conn, p, remoteAddr)
	//case proto.OpMetaCreateDentry:
	//	err = m.opCreateDentry(conn, p, remoteAddr)
	//case proto.OpMetaDeleteDentry:
	//	err = m.opDeleteDentry(conn, p, remoteAddr)
	//case proto.OpMetaUpdateDentry:
	//	err = m.opUpdateDentry(conn, p, remoteAddr)
	//case proto.OpMetaReadDir:
	//	err = m.opReadDir(conn, p, remoteAddr)
	//case proto.OpCreateMetaPartition:
	//	err = m.opCreateMetaPartition(conn, p, remoteAddr)
	//case proto.OpMetaNodeHeartbeat:
	//	err = m.opMasterHeartbeat(conn, p, remoteAddr)
	//case proto.OpMetaLookup:
	//	err = m.opMetaLookup(conn, p, remoteAddr)
	//case proto.OpMetaLookupName:
	//	err = m.opMetaLookupName(conn, p, remoteAddr)
	//case proto.OpDeleteMetaPartition:
	//	err = m.opDeleteMetaPartition(conn, p, remoteAddr)
	//case proto.OpUpdateMetaPartition:
	//	err = m.opUpdateMetaPartition(conn, p, remoteAddr)
	//case proto.OpLoadMetaPartition:
	//	err = m.opLoadMetaPartition(conn, p, remoteAddr)
	//case proto.OpDecommissionMetaPartition:
	//	err = m.opDecommissionMetaPartition(conn, p, remoteAddr)
	//case proto.OpAddMetaPartitionRaftMember:
	//	err = m.opAddMetaPartitionRaftMember(conn, p, remoteAddr)
	//case proto.OpRemoveMetaPartitionRaftMember:
	//	err = m.opRemoveMetaPartitionRaftMember(conn, p, remoteAddr)
	//case proto.OpMetaPartitionTryToLeader:
	//	err = m.opMetaPartitionTryToLeader(conn, p, remoteAddr)
	//case proto.OpMetaBatchInodeGet:
	//	err = m.opMetaBatchInodeGet(conn, p, remoteAddr)
	default:
		err = fmt.Errorf("%s unknown Opcode: %d, reqId: %d", remoteAddr,
			p.Opcode, p.GetReqID())
	}
	if err != nil {
		err = fmt.Errorf("%s [%s] req: %d - %v", remoteAddr, p.GetOpMsg(), p.GetReqID(), err)
	}

	return err

}

// LoadMetaPartition returns the meta partition with the specified volName.
func (m *metadataManager) getPartition(id uint64) (mp MetaPartition, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mp, ok := m.partitions[id]
	if ok {
		return
	}
	err = errors.New(fmt.Sprintf("unknown meta partition: %d", id))
	return
}

func (m *metadataManager) GetPartition(id uint64) (MetaPartition, error) {
	return m.getPartition(id)
}

func (m *metadataManager) LoadStat() string {
	return fmt.Sprintf("state total/loaded : %d/%d", m.localPartitionCount, len(m.partitions))
}

// NewMetadataManager returns a new metadata manager.
func NewMetadataManager(conf MetadataManagerConfig) *metadataManager {
	return &metadataManager{
		nodeId:     conf.NodeID,
		rootDir:    conf.RootDir,
		raftStore:  conf.RaftStore,
		partitions: make(map[uint64]MetaPartition),
		connPool:   util.NewConnectPool(),
	}
}