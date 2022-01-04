package meta

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

func (partition *PartitionFSM) getDentryTree() *BTree {
	return partition.dentryTree.GetTree()
}

func (partition *PartitionFSM) GetDentryTree() *btree.BTree {
	return partition.dentryTree.tree.Clone()
}

func (partition *PartitionFSM) CreateDentry(req *proto.CreateDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) DeleteDentry(req *proto.DeleteDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) UpdateDentry(req *proto.UpdateDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) ReadDir(req *proto.ReadDirRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) Lookup(req *proto.LookupRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) LookupName(req *proto.LookupNameRequest, p *proto.Packet) (err error) {
	panic("implement me")
}
