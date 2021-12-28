package meta

import (
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
	RootDir   string
	RaftStore raftstore.RaftStore
}

type Cluster struct {
	sync.RWMutex

	nodeId              uint64
	rootDir             string
	raftStore           raftstore.RaftStore
	connPool            *util.ConnectPool
	state               uint32
	partitions          map[uint64]*MetaPartitionFSM // Key: metaRangeId, Val: partition.MetaPartitionFSM
	localPartitionCount int
}

func NewCluster(conf MetadataManagerConfig) *Cluster {
	return &Cluster{
		nodeId:     conf.NodeID,
		rootDir:    conf.RootDir,
		raftStore:  conf.RaftStore,
		partitions: make(map[uint64]*MetaPartitionFSM),
		connPool:   util.NewConnectPool(),
	}
}

// Start INFO: 从 metadataDir 目录中加载本地已有的 partitions
func (m *Cluster) Start() error {
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
	var wg sync.WaitGroup
	dirEntryList, _ := ioutil.ReadDir(m.rootDir)
	for _, dirEntry := range dirEntryList {
		// 必须是 partition_xxx 目录
		fileName := dirEntry.Name()
		if !dirEntry.IsDir() || !strings.HasPrefix(fileName, partitionPrefix) {
			continue
		}

		wg.Add(1)
		m.localPartitionCount++
		go func(fileName string) {
			defer wg.Done()

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
			// /data/metanode/partition/partition_1/snapshot/
			snapshotDir := path.Join(partitionConfig.RootDir, SnapshotDir)
			// 如果没有/snapshot目录，就从 /.snapshot_backup 目录rename到 /snapshot，如果 /.snapshot_backup 存在的话
			if _, errload := os.Stat(snapshotDir); errload != nil {
				backupDir := path.Join(partitionConfig.RootDir, snapshotBackup)
				if _, errload = os.Stat(backupDir); errload == nil {
					if errload = os.Rename(backupDir, snapshotDir); errload != nil {
						errload = fmt.Errorf("fail recover backup snapshot %s with err %v", snapshotDir, errload)
						return
					}
				}
			}

			if err = m.attachPartition(id, NewMetaPartitionFSM(partitionConfig)); err != nil {
				klog.Errorf(fmt.Sprintf("load partition id=%d failed: %v.", id, err))
			}
		}(fileName)
	}
	wg.Wait()

	return nil
}

// INFO: 启动每一个 partition，并缓存已经启动的 partition
func (m *Cluster) attachPartition(id uint64, partition *MetaPartitionFSM) error {
	if err := partition.Start(); err != nil {
		klog.Errorf("finish load partition.MetaPartitionFSM %v error %v", id, err)
		return err
	}

	m.Lock()
	defer m.Unlock()
	m.partitions[id] = partition

	return nil
}

func (m *Cluster) detachPartition(id uint64) error {
	m.Lock()
	defer m.Unlock()

	if _, has := m.partitions[id]; has {
		delete(m.partitions, id)
		return nil
	}

	return fmt.Errorf("unknown partition: %d", id)
}

func (m *Cluster) getPartition(id uint64) (*MetaPartitionFSM, error) {
	m.RLock()
	defer m.RUnlock()

	mp, ok := m.partitions[id]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("unknown meta partition: %d", id))
	}
	return mp, nil
}

func (m *Cluster) GetPartition(id uint64) (*MetaPartitionFSM, error) {
	return m.getPartition(id)
}

func (m *Cluster) LoadStat() string {
	return fmt.Sprintf("state total/loaded : %d/%d", m.localPartitionCount, len(m.partitions))
}

// Reply data through tcp connection to the client.
func (m *Cluster) respondToClient(conn net.Conn, p *proto.Packet) (err error) {
	// Handle panic
	defer func() {
		if r := recover(); r != nil {
			switch data := r.(type) {
			case error:
				err = data
			default:
				err = fmt.Errorf(data.(string))
			}
		}
	}()

	// process data and send reply though specified tcp connection.
	err = p.WriteToConn(conn)
	if err != nil {
		klog.Errorf("response to client[%v], request[%s], response packet[%s]", err, p.GetOpMsg(), p.GetResultMsg())
	}
	return
}

// Stop INFO: stop 每一个 partition
func (m *Cluster) Stop() {
	for _, partition := range m.partitions {
		partition.Stop()
	}
}
