package client

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
	"strconv"
	"time"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

// MkDir Create a directory inode as a child of an existing directory inode.
// The kernel sends this in response to a mkdir(2) call.
//func (fs *FuseFS) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
//	parentInodeID := op.Parent
//
//	inodeInfo, err := fs.metaClient.CreateInodeAndDentry(parentInodeID, op.Name, uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)
//	if err != nil {
//		klog.Errorf(fmt.Sprintf("[MkDir]create inode/dentry for %d/%s err %v", uint64(parentInodeID), op.Name, err))
//		return err
//	}
//
//	child := NewInode(inodeInfo)
//	fs.inodeCache.Put(child)
//	parent, err := fs.GetInode(parentInodeID)
//	if err == nil {
//		parent.dentryCache.Put(op.Name, inodeInfo.Inode)
//	}
//
//	op.Entry = GetChildInodeEntry(child)
//
//	klog.Infof(fmt.Sprintf("[MkDir]mkdir op name %s", op.Name))
//
//	return nil
//}
//
//func (fs *FuseFS) MkNode(ctx context.Context, op *fuseops.MkNodeOp) error {
//	klog.Warningf("MkNode is not support!")
//	return fuse.ENOSYS
//}

//func (fs *FuseFS) RmDir(ctx context.Context, op *fuseops.RmDirOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) Unlink(ctx context.Context, op *fuseops.UnlinkOp) error {
//	panic("implement me")
//}

func (fs *FuseFS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	return nil
}

func (fs *FuseFS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	klog.Infof(fmt.Sprintf("[ReadDir]%+v", *op))
	inode, err := fs.GetInode(op.Inode)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[ReadDir]inodeID:%d, err:%v", op.Inode, err))
		return err
	}

	klog.Infof(fmt.Sprintf("[ReadDir]inode:%+v", *inode))

	children, err := fs.metaClient.ReadDir(op.Inode)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[ReadDir]inodeID:%d, err:%v", op.Inode, err))
		return err
	}

	for i := int(op.Offset); i < len(children); i++ {
		child := children[i]
		n := fuseutil.WriteDirent(op.Dst[op.BytesRead:], fuseutil.Dirent{
			Offset: fuseops.DirOffset(i) + 1,
			Inode:  fuseops.InodeID(child.Inode),
			Name:   child.Name,
			Type:   ParseType(child.Type),
		})
		if n == 0 {
			break
		}
		op.BytesRead += n
	}

	klog.Infof(fmt.Sprintf("[ReadDir]%+v", *op))
	return nil
}

//func (fs *FuseFS) ReleaseDirHandle(ctx context.Context, op *fuseops.ReleaseDirHandleOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) CreateLink(ctx context.Context, op *fuseops.CreateLinkOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) CreateSymlink(ctx context.Context, op *fuseops.CreateSymlinkOp) error {
//	panic("implement me")
//}

func (fs *FuseFS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	parentInode, err := fs.GetInode(op.Parent)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
		return err
	}

	inodeID, ok := parentInode.dentryCache.Get(op.Name)
	if !ok {
		inodeID, _, err = fs.metaClient.Lookup(op.Parent, op.Name)
		if err != nil {
			klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
			return err
		}
	}

	childInode, err := fs.GetInode(fuseops.InodeID(inodeID))
	if err != nil {
		klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
		return err
	}
	
	klog.Infof(fmt.Sprintf("[LookUpInode]childInode:%+v", *childInode))
	
	op.Entry.Child = childInode.inodeID
	op.Entry.AttributesExpiration = time.Now().Add(AttrValidDuration)
	op.Entry.EntryExpiration = time.Now().Add(LookupValidDuration)
	op.Entry.Attributes = fuseops.InodeAttributes{
		Size:   childInode.size,
		Nlink:  childInode.nlink,
		Mode:   childInode.mode,
		Atime:  time.Unix(childInode.accessTime, 0),
		Mtime:  time.Unix(childInode.modifyTime, 0),
		Ctime:  time.Unix(childInode.createTime, 0),
		Crtime: time.Unix(childInode.createTime, 0),
		Uid:    childInode.uid,
		Gid:    childInode.gid,
	}
	parentInode, err = fs.GetInode(op.Parent)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
		return err
	}

	parentInode.dentryCache.Put(op.Name, uint64(childInode.inodeID))
	klog.Infof(fmt.Sprintf("[LookUpInode]inodeID:%d, name:%s", childInode.inodeID, op.Name))
	return nil
}

func (fs *FuseFS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	klog.Infof(fmt.Sprintf("[GetInodeAttributes]inodeID:%d", op.Inode))
	inode, err := fs.GetInode(op.Inode)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[GetInodeAttributes]inodeID:%d, err:%v", op.Inode, err))
		return err
	}

	op.Attributes = fuseops.InodeAttributes{
		Size:   inode.size,
		Nlink:  inode.nlink,
		Mode:   inode.mode,
		Atime:  time.Unix(inode.accessTime, 0),
		Mtime:  time.Unix(inode.modifyTime, 0),
		Ctime:  time.Unix(inode.createTime, 0),
		Crtime: time.Unix(inode.createTime, 0),
		Uid:    inode.uid,
		Gid:    inode.gid,
	}
	op.AttributesExpiration = time.Now().Add(AttrValidDuration)

	klog.Infof(fmt.Sprintf("[GetInodeAttributes]inodeID:%d, attr:%s", op.Inode, op.Attributes.DebugString()))
	return nil
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

func ParseType(t uint32) fuseutil.DirentType {
	if proto.IsDir(t) {
		return fuseutil.DT_Directory
	} else if proto.IsSymlink(t) {
		return fuseutil.DT_Link
	}
	return fuseutil.DT_File
}
