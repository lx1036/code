package client

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

/*
`ls globalmount` 必须的方法: OpenDir()/ReadDir()/LookUpInode()/GetInodeAttributes()


*/

// MkDir `mkdir globalmount/1`
func (fs *FuseFS) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	klog.Infof(fmt.Sprintf("[MkDir]:%+v", *op))

	inodeInfo, err := fs.metaClient.CreateInodeAndDentry(op.Parent, op.Name, uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[MkDir]create inode/dentry for %d/%s err %v", op.Parent, op.Name, err))
		return err
	}

	child := NewInode(inodeInfo)
	op.Entry = fuseops.ChildInodeEntry{
		Child: child.inodeID,
		Attributes: fuseops.InodeAttributes{
			Size:   child.size,
			Nlink:  child.nlink,
			Mode:   child.mode,
			Atime:  time.Unix(child.accessTime, 0),
			Mtime:  time.Unix(child.modifyTime, 0),
			Ctime:  time.Unix(child.createTime, 0),
			Crtime: time.Time{},
			Uid:    child.uid,
			Gid:    child.gid,
		},
		AttributesExpiration: time.Now().Add(AttrValidDuration),
		EntryExpiration:      time.Now().Add(LookupValidDuration),
	}
	fs.inodeCache.Put(child)
	parent, err := fs.GetInode(op.Parent)
	if err == nil {
		parent.dentryCache.Put(op.Name, inodeInfo.Inode)
	}

	klog.Infof(fmt.Sprintf("[MkDir]mkdir op name %s", op.Name))

	return nil
}

func (fs *FuseFS) MkNode(ctx context.Context, op *fuseops.MkNodeOp) error {
	klog.Warningf("MkNode is not support!")
	return fuse.ENOSYS
}

//func (fs *FuseFS) RmDir(ctx context.Context, op *fuseops.RmDirOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) Unlink(ctx context.Context, op *fuseops.UnlinkOp) error {
//	panic("implement me")
//}

type DirHandleCache struct {
	sync.RWMutex

	handles map[fuseops.HandleID]*DirHandle
}
type DirHandle struct {
	sync.RWMutex

	inodeID fuseops.InodeID
	entries []proto.Dentry
}

func NewDirHandleCache() *DirHandleCache {
	return &DirHandleCache{
		handles: make(map[fuseops.HandleID]*DirHandle),
	}
}

func (dirHandleCache *DirHandleCache) Put(inodeID fuseops.InodeID, handleID fuseops.HandleID) {
	dirHandleCache.Lock()
	defer dirHandleCache.Unlock()
	dirHandleCache.handles[handleID] = &DirHandle{
		inodeID: inodeID,
	}
}

func (dirHandleCache *DirHandleCache) Release(handleID fuseops.HandleID) *DirHandle {
	dirHandleCache.Lock()
	defer dirHandleCache.Unlock()
	if dirHandle, ok := dirHandleCache.handles[handleID]; ok {
		delete(dirHandleCache.handles, handleID)
		return dirHandle
	}

	return nil
}

func (dirHandleCache *DirHandleCache) Get(handleID fuseops.HandleID) *DirHandle {
	return dirHandleCache.handles[handleID]
}

func (fs *FuseFS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	fs.dirHandleCache.Put(op.Inode, op.Handle)
	return nil
}

func (fs *FuseFS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	klog.Infof(fmt.Sprintf("[ReadDir]inodeID:%d, handleID:%d", op.Inode, op.Handle))
	dirHandle := fs.dirHandleCache.Get(op.Handle)
	if dirHandle == nil {
		fs.dirHandleCache.Put(op.Inode, op.Handle)
		dirHandle = fs.dirHandleCache.Get(op.Handle)
	}

	dirHandle.Lock()
	defer dirHandle.Unlock()
	if op.Offset == 0 || len(dirHandle.entries) == 0 {
		children, err := fs.metaClient.ReadDir(op.Inode)
		if err != nil {
			klog.Errorf(fmt.Sprintf("[ReadDir]%+v", err))
			return err
		}
		dirHandle.entries = children
	}
	if int(op.Offset) > len(dirHandle.entries) {
		klog.Errorf(fmt.Sprintf("[ReadDir]op.Offset %d > dirHandle.entries %d", op.Offset, dirHandle.entries))
		return fuse.EIO
	}

	klog.Infof(fmt.Sprintf("[ReadDir]children dirs: %+v, op.Inode:%d", dirHandle.entries, op.Inode))

	inode, err := fs.GetInode(op.Inode)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[ReadDir]inodeID:%d, err:%v", op.Inode, err))
		return err
	}

	entries := dirHandle.entries[op.Offset:]
	for i, entry := range entries {
		n := fuseutil.WriteDirent(op.Dst[op.BytesRead:], fuseutil.Dirent{
			Offset: fuseops.DirOffset(i + 1),
			Inode:  fuseops.InodeID(entry.Inode),
			Name:   entry.Name,
			Type:   ParseType(entry.Type),
		})
		if n == 0 {
			break
		}

		op.BytesRead += n

		inode.dentryCache.Put(entry.Name, entry.Inode)
	}

	return nil
}

func (fs *FuseFS) ReleaseDirHandle(ctx context.Context, op *fuseops.ReleaseDirHandleOp) error {
	dirHandle := fs.dirHandleCache.Release(op.Handle)
	if dirHandle != nil {
		klog.Infof(fmt.Sprintf("[ReleaseDirHandle]inodeID:%d, handleID:%d", dirHandle.inodeID, op.Handle))
	} else {
		klog.Warningf(fmt.Sprintf("[ReleaseDirHandle]handleID:%d", op.Handle))
	}

	return nil
}

func (fs *FuseFS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	// {Parent:1 Name:1.txt Entry:{Child:0 Generation:0 Attributes:{Size:0 Nlink:0 Mode:---------- Atime:0001-01-01 00:00:00 +0000 UTC Mtime:0001-01-01 00:00:00 +0000 UTC Ctime:0001-01-01 00:00:00 +0000 UTC Crtime:0001-01-01 00:00:00 +0000 UTC Uid:0 Gid:0} AttributesExpiration:0001-01-01 00:00:00 +0000 UTC EntryExpiration:0001-01-01 00:00:00 +0000 UTC} OpContext:{Pid:36698}}
	klog.Infof(fmt.Sprintf("[LookUpInode]LookUpInodeOp:%+v", *op))

	parentInode, err := fs.GetInode(op.Parent)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
		return err
	}

	inodeID, ok := parentInode.dentryCache.Get(op.Name)
	if !ok {
		inodeID, _, err = fs.metaClient.Lookup(op.Parent, op.Name)
		if err != nil {
			klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, name:%s, err:%v", op.Parent, op.Name, err))
			return err
		}
	}

	child, err := fs.GetInode(fuseops.InodeID(inodeID))
	if err != nil {
		klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
		return err
	}

	// {inodeID:16777218 parentInodeID:1 size:6 nlink:1 uid:0 gid:0 gen:1 createTime:1639994505 modifyTime:1639994505 accessTime:1639994505 mode:420 target:[] fullPathName: expiration:1641274151526565000 dentryCache:<nil>}
	klog.Infof(fmt.Sprintf("[LookUpInode]childInode:%+v", *child))

	op.Entry = fuseops.ChildInodeEntry{
		Child: child.inodeID,
		Attributes: fuseops.InodeAttributes{
			Size:   child.size,
			Nlink:  child.nlink,
			Mode:   child.mode,
			Atime:  time.Unix(child.accessTime, 0),
			Mtime:  time.Unix(child.modifyTime, 0),
			Ctime:  time.Unix(child.createTime, 0),
			Crtime: time.Time{},
			Uid:    child.uid,
			Gid:    child.gid,
		},
		AttributesExpiration: time.Now().Add(AttrValidDuration),
		EntryExpiration:      time.Now().Add(LookupValidDuration),
	}
	parentInode, err = fs.GetInode(op.Parent)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
		return err
	}

	parentInode.dentryCache.Put(op.Name, uint64(child.inodeID))
	klog.Infof(fmt.Sprintf("[LookUpInode]inodeID:%d, name:%s", child.inodeID, op.Name))
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

func (fs *FuseFS) ForgetInode(ctx context.Context, op *fuseops.ForgetInodeOp) error {
	return nil
}

//func (fs *FuseFS) CreateLink(ctx context.Context, op *fuseops.CreateLinkOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) CreateSymlink(ctx context.Context, op *fuseops.CreateSymlinkOp) error {
//	panic("implement me")
//}

func ParseType(t uint32) fuseutil.DirentType {
	if proto.IsDir(t) {
		return fuseutil.DT_Directory
	} else if proto.IsSymlink(t) {
		return fuseutil.DT_Link
	}
	return fuseutil.DT_File
}
