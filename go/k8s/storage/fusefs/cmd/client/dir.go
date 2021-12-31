package client

import (
	"context"
	"fmt"
	"strconv"
	"syscall"
	"time"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

// MkDir Create a directory inode as a child of an existing directory inode.
// The kernel sends this in response to a mkdir(2) call.
func (fs *FuseFS) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	parentInodeID := op.Parent

	inodeInfo, err := fs.metaClient.CreateInodeAndDentry(parentInodeID, op.Name, uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[MkDir]create inode/dentry for %d/%s err %v", uint64(parentInodeID), op.Name, err))
		return err
	}

	child := NewInode(inodeInfo)
	fs.inodeCache.Put(child)
	parent, err := fs.GetInode(parentInodeID)
	if err == nil {
		parent.dentryCache.Put(op.Name, inodeInfo.Inode)
	}

	op.Entry = GetChildInodeEntry(child)

	klog.Infof(fmt.Sprintf("[MkDir]mkdir op name %s", op.Name))

	return nil
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
	inode, err := fs.GetInode(op.Inode)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[GetInodeAttributes]inodeID:%d, err:%v", op.Inode, err))
		return err
	}

	attr := &op.Attributes
	attr.Nlink = inode.nlink
	attr.Mode = inode.mode
	attr.Size = inode.size
	attr.Atime = time.Unix(inode.accessTime, 0)
	attr.Ctime = time.Unix(inode.createTime, 0)
	attr.Mtime = time.Unix(inode.modifyTime, 0)
	attr.Uid = inode.uid
	attr.Gid = inode.gid
	op.AttributesExpiration = time.Now().Add(AttrValidDuration)

	klog.Infof(fmt.Sprintf("[GetInodeAttributes]inodeID:%d, attr:%s", op.Inode, op.Attributes.DebugString()))
	return nil
}

func (fs *FuseFS) SetInodeAttributes(ctx context.Context, op *fuseops.SetInodeAttributesOp) error {
	if fs.metaClient.IsVolumeReadOnly() {
		return syscall.EROFS
	}

}

// fullPathName=true, s3 key 则是 path；否则是 inodeID
func (fs *FuseFS) getS3Key(inodeID fuseops.InodeID) (string, error) {
	if !fs.fullPathName {
		return strconv.FormatUint(uint64(inodeID), 10), nil
	}

	if uint64(inodeID) == proto.RootInode {
		return "", nil
	}

	inode, err := fs.GetInode(inodeID)
	if err != nil {
		return "", err
	}
	if len(inode.fullPathName) != 0 {
		return inode.fullPathName, nil
	}

	pInode := inode
	currentInodeID := pInode.inodeID
	for parentInodeID := pInode.parentInodeID; uint64(pInode.inodeID) != proto.RootInode; parentInodeID = pInode.parentInodeID {
		pInode, err = fs.GetInode(parentInodeID)
		if err != nil {
			klog.Errorf(fmt.Sprintf("[getS3Key]get inodeID:%d err:%v", parentInodeID, err))
			return "", err
		}

		name, ok := pInode.dentryCache.GetByInode(currentInodeID)
		if !ok {
			name, err = fs.metaClient.LookupName(parentInodeID, currentInodeID)
			if err != nil {
				klog.Errorf(fmt.Sprintf("[getS3Key]get inodeID:%d LookupName err:%v", parentInodeID, err))
				return "", err
			}
		}

		inode.fullPathName = name
	}

	return inode.fullPathName, nil
}
