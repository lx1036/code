package partition

import (
	"encoding/json"

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

func (partition *MetaPartitionFSM) getInodeTree() *BTree {
	return partition.inodeTree.GetTree()
}

func (partition *MetaPartitionFSM) GetInodeTree() *btree.BTree {
	return partition.inodeTree.tree.Clone()
}

func (partition *MetaPartitionFSM) CreateInode(req *proto.CreateInodeRequest, p *proto.Packet) error {
	inoID, err := partition.nextInodeID()
	if err != nil {
		p.PacketErrorWithBody(proto.OpInodeFullErr, []byte(err.Error()))
		return err
	}

	ino := NewInode(inoID, req.Mode)
	ino.Uid = req.Uid
	ino.Gid = req.Gid
	ino.LinkTarget = req.Target
	ino.PInode = req.PInode
	value, err := ino.Marshal()
	if err != nil {
		p.PacketErrorWithBody(proto.OpErr, []byte(err.Error()))
		return err
	}

	resp, err := partition.Put(opFSMCreateInode, value)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return err
	}

	reply, _ := json.Marshal(resp)
	p.PacketErrorWithBody(proto.OpOk, reply)
	return nil
}

func (partition *MetaPartitionFSM) UnlinkInode(req *proto.UnlinkInodeRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) InodeGet(req *proto.InodeGetRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) InodeGetBatch(req *proto.BatchInodeGetRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) CreateInodeLink(req *proto.LinkInodeRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) EvictInode(req *proto.EvictInodeRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *MetaPartitionFSM) SetAttr(reqData []byte, p *proto.Packet) (err error) {
	panic("implement me")
}
