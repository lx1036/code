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

func (partition *metaPartition) CreateDentry(req *proto.CreateDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) DeleteDentry(req *proto.DeleteDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) UpdateDentry(req *proto.UpdateDentryRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) ReadDir(req *proto.ReadDirRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) Lookup(req *proto.LookupRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) LookupName(req *proto.LookupNameRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) GetDentryTree() *btree.BTree {
	return partition.dentryTree.Clone()
}
