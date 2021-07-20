package meta

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/sunfs/pkg/proto"
	"k8s-lx1036/k8s/storage/sunfs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/sunfs/pkg/util"

	"k8s.io/klog/v2"
)

const partitionPrefix = "partition_"

var localPartionCount int = 0

// MetadataManager manages all the meta partitions.
type MetadataManager interface {
	Start() error
	Stop()
	//CreatePartition(id string, start, end uint64, peers []proto.Peer) error
	HandleMetadataOperation(conn net.Conn, p *proto.Packet, remoteAddr string) error
	GetPartition(id uint64) (MetaPartition, error)
	LoadStat() string
}

// MetadataManagerConfig defines the configures in the metadata manager.
type MetadataManagerConfig struct {
	NodeID    uint64
	RootDir   string
	RaftStore raftstore.RaftStore
}

type metadataManager struct {
	nodeId     uint64
	rootDir    string
	raftStore  raftstore.RaftStore
	connPool   *util.ConnectPool
	state      uint32
	mu         sync.RWMutex
	partitions map[uint64]MetaPartition // Key: metaRangeId, Val: metaPartition
}

func (m *metadataManager) Start() (err error) {
	if atomic.CompareAndSwapUint32(&m.state, StateStandby, StateStart) {
		defer func() {
			var newState uint32
			if err != nil {
				newState = StateStandby
			} else {
				newState = StateRunning
			}
			atomic.StoreUint32(&m.state, newState)
		}()
		err = m.onStart()
	}

	return
}

// onStart creates the connection pool and loads the partitions.
func (m *metadataManager) onStart() (err error) {
	m.connPool = util.NewConnectPool()
	err = m.loadPartitions()
	return
}

func (m *metadataManager) loadPartitions() (err error) {
	// Check metadataDir directory
	fileInfo, err := os.Stat(m.rootDir)
	if err != nil {
		os.MkdirAll(m.rootDir, 0755)
		err = nil
		return
	}
	if !fileInfo.IsDir() {
		err = fmt.Errorf("metadataDir must be directory")
		return
	}
	// scan the data directory
	fileInfoList, err := ioutil.ReadDir(m.rootDir)
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	for _, fileInfo := range fileInfoList {
		if fileInfo.IsDir() && strings.HasPrefix(fileInfo.Name(), partitionPrefix) {
			localPartionCount++
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
				var id uint64
				partitionId := fileName[len(partitionPrefix):]
				id, errload = strconv.ParseUint(partitionId, 10, 64)
				if errload != nil {
					klog.Warningf("ignore path: %s,not partition", partitionId)
					return
				}
				partitionConfig := &MetaPartitionConfig{
					NodeId:    m.nodeId,
					RaftStore: m.raftStore,
					RootDir:   path.Join(m.rootDir, fileName),
					ConnPool:  m.connPool,
				}
				partitionConfig.AfterStop = func() {
					m.detachPartition(id)
				}
				// check snapshot dir or backup
				snapshotDir := path.Join(partitionConfig.RootDir, snapshotDir)
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
					klog.Errorf("load partition id=%d failed: %s.",
						id, errload.Error())
				}
			}(fileInfo.Name())
		}
	}
	wg.Wait()
	return
}

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

func (m *metadataManager) Stop() {
	panic("implement me")
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
	panic("implement me")
}

// NewMetadataManager returns a new metadata manager.
func NewMetadataManager(conf MetadataManagerConfig) MetadataManager {
	return &metadataManager{
		nodeId:     conf.NodeID,
		rootDir:    conf.RootDir,
		raftStore:  conf.RaftStore,
		partitions: make(map[uint64]MetaPartition),
	}
}
