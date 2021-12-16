package meta

import (
	"errors"
	"fmt"
	"io/ioutil"
	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition"
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
	partitions          map[uint64]partition.MetaPartitionFSM // Key: metaRangeId, Val: partition.MetaPartitionFSM
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
			snapshotDir := path.Join(partitionConfig.RootDir, SnapshotDir) // data/metanode/partition/partition_1/snapshot/
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
func (m *metadataManager) attachPartition(id uint64, partition partition.MetaPartitionFSM) error {
	if err := partition.Start(); err != nil {
		klog.Errorf("finish load partition.MetaPartitionFSM %v error %v", id, err)
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

// LoadMetaPartition returns the meta partition with the specified volName.
func (m *metadataManager) getPartition(id uint64) (mp partition.MetaPartitionFSM, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mp, ok := m.partitions[id]
	if ok {
		return
	}
	err = errors.New(fmt.Sprintf("unknown meta partition: %d", id))
	return
}

func (m *metadataManager) GetPartition(id uint64) (partition.MetaPartitionFSM, error) {
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
		partitions: make(map[uint64]partition.MetaPartitionFSM),
		connPool:   util.NewConnectPool(),
	}
}
