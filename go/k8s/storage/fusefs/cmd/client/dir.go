package client

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
	"os"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"

	"k8s.io/klog/v2"
)

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

// Assign INFO: 在 dir 整个生命周期内(OpenDir -> ReleaseDirHandle)，为该 HandleID 分配一个 *DirHandle 对象
func (dirHandleCache *DirHandleCache) Assign(inodeID fuseops.InodeID, handleID fuseops.HandleID) {
	dirHandleCache.Lock()
	defer dirHandleCache.Unlock()

	if _, ok := dirHandleCache.handles[handleID]; !ok {
		dirHandleCache.handles[handleID] = &DirHandle{
			inodeID: inodeID,
			entries: make([]proto.Dentry, 0),
		}
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

/*
INFO:
 # `ll globalmount`
 # `ll globalmount/abc`
 # 当前 ./globalmount 目录下只有一个 hello 文件
 # `ll` 命令触发的 fuse 接口函数
	OpenDir (inode 1, PID 34787) -> ReadDir (inode 1, PID 34787) ->
	LookUpInode (parent 1, name "hello", PID 34787) -> ReadDir (inode 1, PID 34787) ->
	ReadDir (inode 1, PID 34787) -> ReadDir (inode 1, PID 34787) ->
	GetInodeAttributes (inode 1, PID 34787) -> ReleaseDirHandle(PID 34787) ->
	LookUpInode(parent 1, name "hello", PID 34787) -> LookUpInode (parent 1, name "hello", PID 34787)
*/

// OpenDir INFO: 响应用户态的 open()
func (fs *FuseFS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	fs.Lock()
	defer fs.Unlock()

	// op.Handle 在整个生命周期(OpenDir -> ReleaseDirHandle) 内，该 HandleID 是唯一有效的

	fs.dirHandleCache.Assign(op.Inode, op.Handle)
	return nil
}

// ReleaseDirHandle INFO: 响应用户态的 close(), all file descriptors are closed 后，调用该方法，且该 HandleID 不会哎后续 op 里出现
func (fs *FuseFS) ReleaseDirHandle(ctx context.Context, op *fuseops.ReleaseDirHandleOp) error {
	fs.Lock()
	defer fs.Unlock()

	_ = fs.dirHandleCache.Release(op.Handle)
	return nil
}

// ReadDir INFO: `ll ./dir` Read entries from a directory previously opened with OpenDir.
func (fs *FuseFS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	fs.Lock()
	defer fs.Unlock()

	klog.Infof(fmt.Sprintf("[ReadDir]inodeID:%d, handleID:%d", op.Inode, op.Handle))
	dirHandle := fs.dirHandleCache.Get(op.Handle)
	if op.Offset == 0 || len(dirHandle.entries) == 0 {
		children, err := fs.metaClient.ReadDir(op.Inode)
		if err != nil {
			klog.Errorf(fmt.Sprintf("[ReadDir]%+v", err))
			return err
		}
		dirHandle.entries = children
	}
	if int(op.Offset) > len(dirHandle.entries) {
		klog.Errorf(fmt.Sprintf("[ReadDir]op.Offset %d > dirHandle.entries %d", op.Offset, len(dirHandle.entries)))
		return fuse.EIO
	}

	klog.Infof(fmt.Sprintf("[ReadDir]children dirs: %+v, op.Inode:%d", dirHandle.entries, op.Inode))

	inode, err := fs.GetInode(op.Inode)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[ReadDir]inodeID:%d, err:%v", op.Inode, err))
		return err
	}

	// TODO: op.Offset 从 0 开始，防止当前目录下 inode 过大，那就是多次 ReadDir，这里暂时不考虑
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

// LookUpInode INFO: 根据 child name，从 parent inode 中查询出 inode
func (fs *FuseFS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	fs.Lock()
	defer fs.Unlock()

	// {Parent:1 Name:1.txt Entry:{Child:0 Generation:0 Attributes:{Size:0 Nlink:0 Mode:---------- Atime:0001-01-01 00:00:00 +0000 UTC Mtime:0001-01-01 00:00:00 +0000 UTC Ctime:0001-01-01 00:00:00 +0000 UTC Crtime:0001-01-01 00:00:00 +0000 UTC Uid:0 Gid:0} AttributesExpiration:0001-01-01 00:00:00 +0000 UTC EntryExpiration:0001-01-01 00:00:00 +0000 UTC} OpContext:{Pid:36698}}
	//klog.Infof(fmt.Sprintf("[LookUpInode]LookUpInodeOp:%+v", *op))

	parentInode, err := fs.GetInode(op.Parent)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
		return err
	}

	childInodeID, ok := parentInode.dentryCache.Get(op.Name) // op.Name: "2.txt"
	if !ok {
		// 没有缓存，从当前根目录 "./globalmount" 下，即 inode:1，查询 "2.txt" 的 inode
		childInodeID, _, err = fs.metaClient.Lookup(op.Parent, op.Name) // op.Parent: 1
		if err != nil {
			return fuse.ENOENT
		}
	}

	childInode, err := fs.GetInode(fuseops.InodeID(childInodeID))
	if err != nil {
		klog.Errorf(fmt.Sprintf("[LookUpInode]inodeID:%d, err:%v", op.Parent, err))
		return err
	}

	// {inodeID:16777218 parentInodeID:1 size:6 nlink:1 uid:0 gid:0 gen:1 createTime:1639994505 modifyTime:1639994505 accessTime:1639994505 mode:420 target:[] fullPathName: expiration:1641274151526565000 dentryCache:<nil>}
	//klog.Infof(fmt.Sprintf("[LookUpInode]childInode:%+v", *childInode))

	op.Entry = fuseops.ChildInodeEntry{
		Child: childInode.inodeID,
		Attributes: fuseops.InodeAttributes{
			Size:   childInode.size,
			Nlink:  childInode.nlink,
			Mode:   childInode.mode,
			Atime:  time.Unix(childInode.accessTime, 0),
			Mtime:  time.Unix(childInode.modifyTime, 0),
			Ctime:  time.Unix(childInode.createTime, 0),
			Crtime: time.Unix(childInode.createTime, 0),
			Uid:    childInode.uid,
			Gid:    childInode.gid,
		},
		AttributesExpiration: time.Now().Add(AttrValidDuration),
		EntryExpiration:      time.Now().Add(LookupValidDuration),
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

// GetInodeAttributes 中的 op.Inode 来自于 LookUpInode(name) 中获得的 inode，当 AttributesExpiration 过期后，内核 VFS FUSE 就会调用
// 该函数重新刷新 op.Inode 的 Attributes
func (fs *FuseFS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	fs.Lock()
	defer fs.Unlock()

	//klog.Infof(fmt.Sprintf("[GetInodeAttributes]inodeID:%d", op.Inode))
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

	// 1.txt: "inodeID:33598380, attr:7 1 -rw-r--r-- 501 20"
	// abc 目录："inodeID:16821113, attr:0 2 drwxr-xr-x 501 20"
	klog.Infof(fmt.Sprintf("[GetInodeAttributes]inodeID:%d, attr:%s", op.Inode, op.Attributes.DebugString()))
	return nil
}

// SetInodeAttributes TODO: 用户态 chmod 去修改 file/dir attr 属性，暂不支持
func (fs *FuseFS) SetInodeAttributes(ctx context.Context, op *fuseops.SetInodeAttributesOp) error {
	return nil
}

func (fs *FuseFS) ForgetInode(ctx context.Context, op *fuseops.ForgetInodeOp) error {
	return nil
}

// MkDir `mkdir globalmount/1` @see CreateFile()
func (fs *FuseFS) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	fs.Lock()
	defer fs.Unlock()

	//klog.Infof(fmt.Sprintf("[MkDir]:%+v", *op))

	parent, err := fs.GetInode(op.Parent)
	if err != nil {
		return err
	}
	if _, ok := parent.dentryCache.Get(op.Name); ok {
		klog.Errorf(fmt.Sprintf("[MkDir]name:%s is already exist in parent inodeID:%d", op.Name, op.Parent))
		return fuse.EEXIST // Ensure that the name doesn't already exist, so we don't wind up with a duplicate.
	}

	inodeInfo, err := fs.metaClient.CreateInodeAndDentry(op.Parent, op.Name,
		uint32(os.ModeDir|op.Mode.Perm()), fs.uid, fs.gid, nil)
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
			Crtime: time.Unix(child.createTime, 0),
			Uid:    child.uid,
			Gid:    child.gid,
		},
		AttributesExpiration: time.Now().Add(AttrValidDuration),
		EntryExpiration:      time.Now().Add(LookupValidDuration),
	}
	fs.inodeCache.Put(child)
	parent, err = fs.GetInode(op.Parent)
	if err == nil {
		parent.dentryCache.Put(op.Name, inodeInfo.Inode)
	}

	// {inodeID:43911, name:abd} 为 abd 目录新分配的 inodeID:43911
	klog.Infof(fmt.Sprintf("[MkDir]mkdir allocate inodeID:%d for dir name:%s from meta cluster", inodeInfo.Inode, op.Name))
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
