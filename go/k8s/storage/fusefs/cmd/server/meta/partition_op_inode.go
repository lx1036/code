package meta

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"github.com/google/btree"
)

// OpInode defines the interface for the inode operations.
type OpInode interface {
	CreateInode(req *proto.CreateInodeRequest, p *proto.Packet) (err error)
	UnlinkInode(req *proto.UnlinkInodeRequest, p *proto.Packet) (err error) // delete inode
	InodeGet(req *proto.InodeGetRequest, p *proto.Packet) (err error)
	InodeGetBatch(req *proto.BatchInodeGetRequest, p *proto.Packet) (err error)
	CreateInodeLink(req *proto.CreateInodeLinkRequest, p *proto.Packet) (err error)
	EvictInode(req *proto.EvictInodeRequest, p *proto.Packet) (err error)
	SetAttr(reqData []byte, p *proto.Packet) (err error)
	GetInodeTree() *btree.BTree
}

func (partition *PartitionFSM) getInodeTree() *BTree {
	return partition.inodeTree.GetTree()
}

func (partition *PartitionFSM) GetInodeTree() *btree.BTree {
	return partition.inodeTree.tree.Clone()
}

// CreateInode submit `create inode` cmd to raft
func (partition *PartitionFSM) CreateInode(req *proto.CreateInodeRequest, p *proto.Packet) error {
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
	value, _ := ino.Marshal()
	resp, err := partition.Put(opFSMCreateInode, value)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return err
	}

	reply, _ := json.Marshal(resp)
	p.PacketErrorWithBody(proto.OpOk, reply)
	return nil
}

type InodeResponse struct {
	Status uint8
	Msg    *Inode
}

func (partition *PartitionFSM) CreateInodeLink(req *proto.CreateInodeLinkRequest, p *proto.Packet) (err error) {
	ino := NewInode(req.Inode, 0)
	value, _ := ino.Marshal()
	resp, err := partition.Put(opFSMCreateLinkInode, value)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return err
	}
	retMsg := resp.(*InodeResponse)
	reply, _ := json.Marshal(retMsg)
	p.PacketErrorWithBody(proto.OpOk, reply)
	return nil
}

func (partition *PartitionFSM) UnlinkInode(req *proto.UnlinkInodeRequest, p *proto.Packet) (err error) {
	ino := NewInode(req.Inode, 0)
	value, _ := ino.Marshal()
	resp, err := partition.Put(opFSMUnlinkInode, value)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return err
	}
	retMsg := resp.(*InodeResponse)
	reply, _ := json.Marshal(retMsg)
	p.PacketErrorWithBody(proto.OpOk, reply)
	return nil
}

func (partition *PartitionFSM) BatchUnlinkInode(req *proto.BatchUnlinkInodeRequest, p *proto.Packet) (err error) {
	if len(req.Inodes) == 0 {
		return nil
	}

	var inodes InodeBatch
	for _, id := range req.Inodes {
		inodes = append(inodes, NewInode(id, 0))
	}
	value, _ := inodes.Marshal()
	resp, err := partition.Put(opFSMUnlinkInodeBatch, value)
	if err != nil {
		p.PacketErrorWithBody(proto.OpAgain, []byte(err.Error()))
		return err
	}

	retMsg := resp.([]*InodeResponse)
	reply, _ := json.Marshal(retMsg)
	p.PacketErrorWithBody(proto.OpOk, reply)
	return nil
}

func (partition *PartitionFSM) InodeGet(req *proto.InodeGetRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) InodeGetBatch(req *proto.BatchInodeGetRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) EvictInode(req *proto.EvictInodeRequest, p *proto.Packet) (err error) {
	panic("implement me")
}

func (partition *PartitionFSM) SetAttr(reqData []byte, p *proto.Packet) (err error) {
	panic("implement me")
}

// Return a new inode ID and update the offset.
func (partition *PartitionFSM) nextInodeID() (inodeId uint64, err error) {
	for {
		cur := atomic.LoadUint64(&partition.config.Cursor)
		end := partition.config.End
		if cur >= end {
			return 0, fmt.Errorf("inode ID out of range")
		}
		newId := cur + 1
		if atomic.CompareAndSwapUint64(&partition.config.Cursor, cur, newId) {
			return newId, nil
		}
	}
}
