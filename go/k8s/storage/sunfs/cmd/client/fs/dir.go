package fs

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/storage/fuse/fuseops"

	"k8s.io/klog/v2"
)

// MkDir Create a directory inode as a child of an existing directory inode.
// The kernel sends this in response to a mkdir(2) call.
func (super *Super) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	parentInodeID := op.Parent

	inodeInfo, err := super.metaClient.Create_ll(uint64(parentInodeID), op.Name, uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)
	if err != nil {
		return err
	}

	child := NewInode(inodeInfo)
	super.inodeCache.Put(child)
	parent, err := super.InodeGet(uint64(parentInodeID))
	if err == nil {
		parent.dentryCache.Put(op.Name, inodeInfo.Inode)
	}

	fillChildEntry(&op.Entry, child)

	klog.Infof(fmt.Sprintf("[MkDir]mkdir op name %s", op.Name))

	return nil
}

// INFO: 从本地缓存 InodeCache 取值，如果没有调用 meta cluster api 获取并存入 InodeCache
func (super *Super) InodeGet(inodeID uint64) (*Inode, error) {
	inode := super.inodeCache.Get(inodeID)
	if inode != nil {
		return inode, nil
	}

	// 本地缓存里没有，从 meta cluster 中api取
	inodeInfo, err := super.metaClient.InodeGet_ll(inodeID)
	if err != nil {
		return nil, err
	}

	inode = NewInode(inodeInfo)
	super.inodeCache.Put(inode)

	return inode, nil
}

func (super *Super) MkNode(ctx context.Context, op *fuseops.MkNodeOp) error {
	panic("implement me")
}

func (super *Super) RmDir(ctx context.Context, op *fuseops.RmDirOp) error {
	panic("implement me")
}

func (super *Super) Unlink(ctx context.Context, op *fuseops.UnlinkOp) error {
	panic("implement me")
}

func (super *Super) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	panic("implement me")
}

func (super *Super) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	panic("implement me")
}

func (super *Super) ReleaseDirHandle(ctx context.Context, op *fuseops.ReleaseDirHandleOp) error {
	panic("implement me")
}

func (super *Super) CreateLink(ctx context.Context, op *fuseops.CreateLinkOp) error {
	panic("implement me")
}

func (super *Super) CreateSymlink(ctx context.Context, op *fuseops.CreateSymlinkOp) error {
	panic("implement me")
}

func (super *Super) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	return nil
}
