package meta

import (
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"github.com/google/btree"
)

// OpInode defines the interface for the inode operations.
type OpInode interface {
	CreateInode(req *proto.CreateInodeRequest, p *proto.Packet) (err error)
	UnlinkInode(req *proto.UnlinkInodeRequest, p *proto.Packet) (err error) // delete inode
	InodeGet(req *proto.InodeGetRequest, p *proto.Packet) (err error)
	InodeGetBatch(req *proto.BatchInodeGetRequest, p *proto.Packet) (err error)
	CreateInodeLink(req *proto.LinkInodeRequest, p *proto.Packet) (err error)
	EvictInode(req *proto.EvictInodeRequest, p *proto.Packet) (err error)
	SetAttr(reqData []byte, p *proto.Packet) (err error)
	GetInodeTree() *btree.BTree
}

func (partition *metaPartition) CreateInode(req *proto.CreateInodeRequest, p *proto.Packet) error {
	inoID, err := partition.nextInodeID()
	if err != nil {
		return err
	}

	ino := NewInode(inoID, req.Mode)
	ino.Uid = req.Uid
	ino.Gid = req.Gid
	ino.LinkTarget = req.Target
	ino.PInode = req.PInode
	value, err := ino.Marshal()
	if err != nil {
		return err
	}

	resp, err := partition.Put(opFSMCreateInode, value)
	if err != nil {

	}

	return nil
}

func (partition *metaPartition) UnlinkInode(req *proto.UnlinkInodeRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) InodeGet(req *proto.InodeGetRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) InodeGetBatch(req *proto.BatchInodeGetRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) CreateInodeLink(req *proto.LinkInodeRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) EvictInode(req *proto.EvictInodeRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) SetAttr(reqData []byte, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *metaPartition) GetInodeTree() *btree.BTree {
	return partition.inodeTree.Clone()
}
