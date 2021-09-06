package meta

import (
	"os"
	"path"
	"strings"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"k8s-lx1036/k8s/storage/fusefs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/fusefs/pkg/util"

	"github.com/google/btree"
	raftproto "github.com/tiglabs/raft/proto"
)

const DefaultBTreeDegree = 32

const (
	opFSMCreateInode uint32 = iota
	opFSMUnlinkInode
	opFSMCreateDentry
	opFSMDeleteDentry
	opFSMDeletePartition
	opFSMUpdatePartition
	opFSMDecommissionPartition
	opFSMStoreTick
	startStoreTick
	stopStoreTick
	opFSMUpdateDentry
	opFSMCreateLinkInode
	opFSMEvictInode
	opFSMSetAttr
)

// MetaPartition defines the interface for the meta partition operations.
type MetaPartition interface {
	Start() error
	Stop()
	OpMeta
}

// OpMeta defines the interface for the metadata operations.
type OpMeta interface {
	OpInode

	OpDentry

	OpPartition
}

// OpPartition defines the interface for the partition operations.
type OpPartition interface {
	IsLeader() (leaderAddr string, isLeader bool)
	GetCursor() uint64
	GetSize() uint64
	GetBaseConfig() MetaPartitionConfig
	LoadSnapshotSign(p *proto.Packet) (err error)
	PersistMetadata() (err error)
	ChangeMember(changeType raftproto.ConfChangeType, peer raftproto.Peer, context []byte) (resp interface{}, err error)
	DeletePartition() (err error)
	UpdatePartition(req *proto.UpdateMetaPartitionRequest, resp *proto.UpdateMetaPartitionResponse) (err error)
	DeleteRaft() error
	IsExistPeer(peer raftproto.Peer) bool
	TryToLeader(groupID uint64) error
	CanRemoveRaftMember(peer raftproto.Peer) error
}

// MetaPartitionConfig is used to create a meta partition.
type MetaPartitionConfig struct {
	// Identity for raftStore group. RaftStore nodes in the same raftStore group must have the same groupID.
	PartitionId uint64              `json:"partition_id"`
	VolName     string              `json:"vol_name"`
	Start       uint64              `json:"start"` // Minimal Inode ID of this range. (Required during initialization)
	End         uint64              `json:"end"`   // Maximal Inode ID of this range. (Required during initialization)
	Peers       []proto.Peer        `json:"peers"` // Peers information of the raftStore
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

// metaPartition manages the range of the inode IDs.
// When a new inode is requested, it allocates a new inode id for this inode if possible.
// States:
//  +-----+             +-------+
//  | New | → Restore → | Ready |
//  +-----+             +-------+

// INFO: metadata 是由成百上千的partition组成，每个partition是由两个BTree inode 和 dentry 组成。
//  每个partition使用multi raft组成保证高可用和一致性。
//  @see https://chubaofs.readthedocs.io/zh_CN/latest/design/metanode.html
type metaPartition struct {
	config  *MetaPartitionConfig
	size    uint64 // For partition all file size
	applyID uint64 // Inode/Dentry max applyID, this index will be update after restoring from the dumped data.

	// TODO: 这里BTree没有并发功能，可以考虑使用读写锁 sync.RWMutex 包装下
	// 每个Inode代表文件系统中的一个文件或目录
	// 每个dentry代表一个目录项，dentry由parentId和name组成
	dentryTree *btree.BTree // dir entry fs.DirEntry
	inodeTree  *btree.BTree // index node

	// raftPartition 需要多个参数一起构造
	raftPartition raftstore.Partition
	stopC         chan bool
	storeChan     chan *storeMsg
	state         uint32
	delInodeFp    *os.File
	manager       *metadataManager
}

func (partition *metaPartition) Stop() {
	panic("implement me")
}

func (partition *metaPartition) ChangeMember(changeType raftproto.ConfChangeType, peer raftproto.Peer, context []byte) (resp interface{}, err error) {
	panic("implement me")
}

func (partition *metaPartition) UpdatePartition(req *proto.UpdateMetaPartitionRequest, resp *proto.UpdateMetaPartitionResponse) (err error) {
	panic("implement me")
}

func (partition *metaPartition) IsLeader() (leaderAddr string, isLeader bool) {
	panic("implement me")
}

func (partition *metaPartition) GetCursor() uint64 {
	panic("implement me")
}

func (partition *metaPartition) GetSize() uint64 {
	panic("implement me")
}

func (partition *metaPartition) GetBaseConfig() MetaPartitionConfig {
	panic("implement me")
}

func (partition *metaPartition) LoadSnapshotSign(p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) PersistMetadata() (err error) {
	panic("implement me")
}

func (partition *metaPartition) DeletePartition() (err error) {
	panic("implement me")
}

func (partition *metaPartition) DeleteRaft() error {
	panic("implement me")
}

func (partition *metaPartition) IsExistPeer(peer raftproto.Peer) bool {
	panic("implement me")
}

func (partition *metaPartition) TryToLeader(groupID uint64) error {
	panic("implement me")
}

func (partition *metaPartition) CanRemoveRaftMember(peer raftproto.Peer) error {
	panic("implement me")
}

// INFO: 从本地snapshot加载 metadata/inode/dentry/applyID
func (partition *metaPartition) loadFromSnapshot() error {
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

// INFO: 创建 raft partition，实际上每一个 partition 都有其自己的 wal path
func (partition *metaPartition) startRaft() error {
	var (
		err           error
		heartbeatPort int
		replicaPort   int
		peers         []raftstore.PeerAddress
	)
	if heartbeatPort, replicaPort, err = partition.getRaftPort(); err != nil {
		return err
	}
	for _, peer := range partition.config.Peers {
		addr := strings.Split(peer.Addr, ":")[0]
		rp := raftstore.PeerAddress{
			Peer: raftproto.Peer{
				ID: peer.ID,
			},
			Address:       addr,
			HeartbeatPort: heartbeatPort,
			ReplicaPort:   replicaPort,
		}
		peers = append(peers, rp)
	}
	partition.raftPartition, err = partition.config.RaftStore.CreatePartition(&raftstore.PartitionConfig{
		ID:      partition.config.PartitionId,
		Applied: partition.applyID,
		Peers:   peers,
		SM:      partition,
	})

	return err
}

// INFO: 启动各个 partition
func (partition *metaPartition) Start() error {
	if err := partition.loadFromSnapshot(); err != nil {
		return err
	}

	partition.startSchedule(partition.applyID)

	if err := partition.startRaft(); err != nil {
		return err
	}

	return nil
}

// NewMetaPartition creates a new meta partition with the specified configuration.
func NewMetaPartition(conf *MetaPartitionConfig, manager *metadataManager) MetaPartition {
	return &metaPartition{
		config:     conf,
		dentryTree: btree.New(DefaultBTreeDegree),
		inodeTree:  btree.New(DefaultBTreeDegree),
		stopC:      make(chan bool),
		storeChan:  make(chan *storeMsg, 5), // INFO: storeMsg channel buffer size = 5
		manager:    manager,
	}
}
