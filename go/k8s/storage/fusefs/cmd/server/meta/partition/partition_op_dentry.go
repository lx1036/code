package partition

import (
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"github.com/google/btree"
)

// OpDentry defines the interface for the dentry operations.
type OpDentry interface {
	CreateDentry(req *proto.CreateDentryRequest, p *proto.Packet) (err error)
	DeleteDentry(req *proto.DeleteDentryRequest, p *proto.Packet) (err error)
	UpdateDentry(req *proto.UpdateDentryRequest, p *proto.Packet) (err error)
	ReadDir(req *proto.ReadDirRequest, p *proto.Packet) (err error)
	Lookup(req *proto.LookupRequest, p *proto.Packet) (err error)
	LookupName(req *proto.LookupNameRequest, p *proto.Packet) (err error)
	GetDentryTree() *btree.BTree
}

func (partition *MetaPartitionFSM) getDentryTree() *BTree {
	return partition.dentryTree.GetTree()
}

func (partition *MetaPartitionFSM) GetDentryTree() *btree.BTree {
	return partition.dentryTree.tree.Clone()
}

func (partition *MetaPartitionFSM) CreateDentry(req *proto.CreateDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) DeleteDentry(req *proto.DeleteDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) UpdateDentry(req *proto.UpdateDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) ReadDir(req *proto.ReadDirRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) Lookup(req *proto.LookupRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) LookupName(req *proto.LookupNameRequest, p *proto.Packet) (err error) {
	panic("implement me")
}
