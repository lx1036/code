package client

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"

	"k8s.io/klog/v2"
)

// MkDir Create a directory inode as a child of an existing directory inode.
// The kernel sends this in response to a mkdir(2) call.
func (fs *FuseFS) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	parentInodeID := op.Parent

	inodeInfo, err := fs.metaClient.Create_ll(uint64(parentInodeID), op.Name, uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[MkDir]create inode/dentry for %d/%s err %v", uint64(parentInodeID), op.Name, err))
		return err
	}

	child := NewInode(inodeInfo)
	fs.inodeCache.Put(child)
	parent, err := fs.InodeGet(uint64(parentInodeID))
	if err == nil {
		parent.dentryCache.Put(op.Name, inodeInfo.Inode)
	}

	op.Entry = GetChildInodeEntry(child)

	klog.Infof(fmt.Sprintf("[MkDir]mkdir op name %s", op.Name))

	return nil
}

// INFO: 从本地缓存 InodeCache 取值，如果没有调用 meta cluster api 获取并存入 InodeCache
func (fs *FuseFS) InodeGet(inodeID uint64) (*Inode, error) {
	inode := fs.inodeCache.Get(inodeID)
	if inode != nil {
		return inode, nil
	}

	// 本地缓存里没有，从 meta cluster 中api取
	inodeInfo, err := fs.metaClient.InodeGet_ll(inodeID)
	if err != nil {
		return nil, err
	}

	inode = NewInode(inodeInfo)
	fs.inodeCache.Put(inode)

	return inode, nil
}

func (fs *FuseFS) MkNode(ctx context.Context, op *fuseops.MkNodeOp) error {
	klog.Warningf("MkNode is not support!")
	return fuse.ENOSYS
}

func (fs *FuseFS) RmDir(ctx context.Context, op *fuseops.RmDirOp) error {
	panic("implement me")
}

func (fs *FuseFS) Unlink(ctx context.Context, op *fuseops.UnlinkOp) error {
	panic("implement me")
}

func (fs *FuseFS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	panic("implement me")
}

func (fs *FuseFS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	panic("implement me")
}

func (fs *FuseFS) ReleaseDirHandle(ctx context.Context, op *fuseops.ReleaseDirHandleOp) error {
	panic("implement me")
}

func (fs *FuseFS) CreateLink(ctx context.Context, op *fuseops.CreateLinkOp) error {
	panic("implement me")
}

func (fs *FuseFS) CreateSymlink(ctx context.Context, op *fuseops.CreateSymlinkOp) error {
	panic("implement me")
}

func (fs *FuseFS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	return nil
}

func (fs *FuseFS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	panic("implement me")
}

func (fs *FuseFS) SetInodeAttributes(ctx context.Context, op *fuseops.SetInodeAttributesOp) error {
	panic("implement me")
}
