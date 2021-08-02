package meta

import (
	"github.com/google/btree"
	"k8s-lx1036/k8s/storage/sunfs/pkg/raftstore"
	"k8s-lx1036/k8s/storage/sunfs/pkg/util"
	"os"

	"github.com/tiglabs/raft/proto"
)

const DefaultBTreeDegree = 32

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

// OpInode defines the interface for the inode operations.
type OpInode interface {
	CreateInode(req *CreateInoReq, p *utilproto.Packet) (err error)
	UnlinkInode(req *UnlinkInoReq, p *utilproto.Packet) (err error)
	InodeGet(req *InodeGetReq, p *utilproto.Packet) (err error)
	InodeGetBatch(req *InodeGetReqBatch, p *utilproto.Packet) (err error)
	CreateInodeLink(req *LinkInodeReq, p *utilproto.Packet) (err error)
	EvictInode(req *EvictInodeReq, p *utilproto.Packet) (err error)
	SetAttr(reqData []byte, p *utilproto.Packet) (err error)
	GetInodeTree() *BTree
}

// OpDentry defines the interface for the dentry operations.
type OpDentry interface {
	CreateDentry(req *CreateDentryReq, p *utilproto.Packet) (err error)
	DeleteDentry(req *DeleteDentryReq, p *utilproto.Packet) (err error)
	UpdateDentry(req *UpdateDentryReq, p *utilproto.Packet) (err error)
	ReadDir(req *ReadDirReq, p *utilproto.Packet) (err error)
	Lookup(req *LookupReq, p *utilproto.Packet) (err error)
	LookupName(req *LookupNameReq, p *utilproto.Packet) (err error)
	GetDentryTree() *BTree
}

// OpPartition defines the interface for the partition operations.
type OpPartition interface {
	IsLeader() (leaderAddr string, isLeader bool)
	GetCursor() uint64
	GetSize() uint64
	GetBaseConfig() MetaPartitionConfig
	LoadSnapshotSign(p *utilproto.Packet) (err error)
	PersistMetadata() (err error)
	ChangeMember(changeType raftproto.ConfChangeType, peer raftproto.Peer, context []byte) (resp interface{}, err error)
	DeletePartition() (err error)
	UpdatePartition(req *UpdatePartitionReq, resp *UpdatePartitionResp) (err error)
	DeleteRaft() error
	IsExsitPeer(peer proto.Peer) bool
	TryToLeader(groupID uint64) error
	CanRemoveRaftMember(peer proto.Peer) error
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
type metaPartition struct {
	config        *MetaPartitionConfig
	size          uint64 // For partition all file size
	applyID       uint64 // Inode/Dentry max applyID, this index will be update after restoring from the dumped data.
	dentryTree    *btree.BTree
	inodeTree     *btree.BTree // btree for inodes
	raftPartition raftstore.Partition
	stopC         chan bool
	storeChan     chan *storeMsg
	state         uint32
	delInodeFp    *os.File
	manager       *metadataManager
}

func (m metaPartition) Start() error {
	panic("implement me")
}

func (m metaPartition) Stop() {
	panic("implement me")
}

func (m metaPartition) CreateInode(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) UnlinkInode(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) InodeGet(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) InodeGetBatch(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) CreateInodeLink(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) EvictInode(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) SetAttr(reqData []byte, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) GetInodeTree() *interface{} {
	panic("implement me")
}

func (m metaPartition) CreateDentry(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) DeleteDentry(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) UpdateDentry(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) ReadDir(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) Lookup(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) LookupName(req *interface{}, p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) GetDentryTree() *interface{} {
	panic("implement me")
}

func (m metaPartition) IsLeader() (leaderAddr string, isLeader bool) {
	panic("implement me")
}

func (m metaPartition) GetCursor() uint64 {
	panic("implement me")
}

func (m metaPartition) GetSize() uint64 {
	panic("implement me")
}

func (m metaPartition) GetBaseConfig() MetaPartitionConfig {
	panic("implement me")
}

func (m metaPartition) LoadSnapshotSign(p *utilproto.Packet) (err error) {
	panic("implement me")
}

func (m metaPartition) PersistMetadata() (err error) {
	panic("implement me")
}

func (m metaPartition) ChangeMember(changeType interface{}, peer interface{}, context []byte) (resp interface{}, err error) {
	panic("implement me")
}

func (m metaPartition) DeletePartition() (err error) {
	panic("implement me")
}

func (m metaPartition) UpdatePartition(req *interface{}, resp *interface{}) (err error) {
	panic("implement me")
}

func (m metaPartition) DeleteRaft() error {
	panic("implement me")
}

func (m metaPartition) IsExsitPeer(peer proto.Peer) bool {
	panic("implement me")
}

func (m metaPartition) TryToLeader(groupID uint64) error {
	panic("implement me")
}

func (m metaPartition) CanRemoveRaftMember(peer proto.Peer) error {
	panic("implement me")
}

// NewMetaPartition creates a new meta partition with the specified configuration.
func NewMetaPartition(conf *MetaPartitionConfig, manager *metadataManager) MetaPartition {
	mp := &metaPartition{
		config:     conf,
		dentryTree: btree.New(DefaultBTreeDegree),
		inodeTree:  btree.New(DefaultBTreeDegree),
		stopC:      make(chan bool),
		storeChan:  make(chan *storeMsg, 5),
		manager:    manager,
	}

	return mp
}
