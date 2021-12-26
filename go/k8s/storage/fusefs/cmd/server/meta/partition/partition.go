package partition

import (
	"errors"
	"k8s-lx1036/k8s/storage/fusefs/cmd/server/meta/partition/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"
	"os"
)

var (
	ErrIllegalHeartbeatAddress = errors.New("illegal heartbeat address")
	ErrIllegalReplicateAddress = errors.New("illegal replicate address")
	ErrInodeIDOutOfRange       = errors.New("inode ID out of range")
)

// MetaPartitionConfig is used to create a meta partition.
type MetaPartitionConfig struct {
	Peers []proto.Peer `json:"peers"` // raft peers

	// Identity for raftStore group. RaftStore nodes in the same raftStore group must have the same groupID.
	PartitionId uint64              `json:"partition_id"`
	VolName     string              `json:"vol_name"`
	Start       uint64              `json:"start"` // Minimal Inode ID of this range. (Required during initialization)
	End         uint64              `json:"end"`   // Maximal Inode ID of this range. (Required during initialization)
	Cursor      uint64              `json:"-"`     // Cursor ID of the inode that have been assigned
	NodeId      uint64              `json:"-"`
	RootDir     string              `json:"-"`
	BeforeStart func()              `json:"-"`
	AfterStart  func()              `json:"-"`
	BeforeStop  func()              `json:"-"`
	AfterStop   func()              `json:"-"`
	RaftStore   raftstore.RaftStore `json:"-"`
	ConnPool    *util.ConnectPool   `json:"-"`
}

// MetaPartitionFSM manages the range of the inode IDs.
// When a new inode is requested, it allocates a new inode id for this inode if possible.
// States:
//  +-----+             +-------+
//  | New | → Restore → | Ready |
//  +-----+             +-------+

// MetaPartitionFSM
// INFO: metadata 是由成百上千的partition组成，每个partition是由两个 BTree inode 和 BTree dentry 组成。
//  每个partition使用multi raft组成保证高可用和一致性。
//  @see https://chubaofs.readthedocs.io/zh_CN/latest/design/metanode.html
type MetaPartitionFSM struct {
	config  *MetaPartitionConfig
	size    uint64 // For partition all file size
	applyID uint64 // Inode/Dentry max applyID, this index will be update after restoring from the dumped data.

	// 每个Inode代表文件系统中的一个文件或目录
	// 每个dentry代表一个目录项，dentry由parentId和name组成
	dentryTree *BTree // dir entry fs.DirEntry
	inodeTree  *BTree // index node

	// raftPartition 需要多个参数一起构造
	raftPartition raftstore.Partition
	stopC         chan bool
	storeChan     chan *storeMsg
	state         uint32
	delInodeFp    *os.File
}

func NewMetaPartitionFSM(conf *MetaPartitionConfig) *MetaPartitionFSM {
	return &MetaPartitionFSM{
		config:     conf,
		dentryTree: NewBtree(),
		inodeTree:  NewBtree(),
		stopC:      make(chan bool),
		storeChan:  make(chan *storeMsg, 5), // INFO: storeMsg channel buffer size = 5
	}
}

// Start INFO: 启动各个 partition
func (partition *MetaPartitionFSM) Start() error {
	if err := partition.loadFromSnapshot(); err != nil {
		return err
	}

	partition.startSchedule(partition.applyID)

	if err := partition.startRaft(); err != nil {
		return err
	}

	return nil
}

// INFO: 从本地snapshot加载 metadata/inode/dentry/applyID
func (partition *MetaPartitionFSM) loadFromSnapshot() error {
	if err := partition.loadMetadata(); err != nil {
		return err
	}
	if err := partition.loadInode(); err != nil {
		return err
	}
	if err := partition.loadDentry(); err != nil {
		return err
	}
	return partition.loadApplyID()
}

func (partition *MetaPartitionFSM) Stop() {
	panic("implement me")
}

func (partition *MetaPartitionFSM) UpdatePartition(req *proto.UpdateMetaPartitionRequest, resp *proto.UpdateMetaPartitionResponse) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) GetCursor() uint64 {
	panic("implement me")
}

func (partition *MetaPartitionFSM) GetSize() uint64 {
	panic("implement me")
}

func (partition *MetaPartitionFSM) GetBaseConfig() MetaPartitionConfig {
	panic("implement me")
}

func (partition *MetaPartitionFSM) PersistMetadata() (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) DeletePartition() (err error) {
	panic("implement me")
}
