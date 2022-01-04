package client

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/storage/fuse"
	"sync"
	"syscall"
	"time"

	"k8s-lx1036/k8s/storage/fuse/fuseops"

	"k8s.io/klog/v2"
)

const (
	NameMaxLen = 256
)

type FileHandleCache struct {
	sync.RWMutex

	buffer map[fuseops.InodeID]*Buffer

	inodeID uint64
	flag    uint32
}

func NewFileHandleCache() *FileHandleCache {
	return &FileHandleCache{
		buffer: make(map[fuseops.InodeID]*Buffer),
	}
}

func (fileHandleCache *FileHandleCache) Put(inodeID fuseops.InodeID, handleID fuseops.HandleID, fs *FuseFS) {
	fileHandleCache.Lock()
	defer fileHandleCache.Unlock()

	if _, ok := fileHandleCache.buffer[inodeID]; !ok {
		key, err := fs.getS3Key(inodeID)
		if err != nil {
			klog.Error(err)
			return
		}
		fileHandleCache.buffer[inodeID] = NewBuffer(key, fs.s3Client)
	}
}

func (fileHandleCache *FileHandleCache) Get(inodeID fuseops.InodeID) *Buffer {
	return fileHandleCache.buffer[inodeID]
}

func (fileHandleCache *FileHandleCache) Release(handleID fuseops.HandleID) {

}

func (fs *FuseFS) OpenFile(ctx context.Context, op *fuseops.OpenFileOp) error {
	fs.fileHandleCache.Put(op.Inode, op.Handle, fs)
	return nil
}

// ReadFile `cat globalmount/1.txt`
func (fs *FuseFS) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
	klog.Infof(fmt.Sprintf("[ReadFile]inodeID:%d, handleID:%d, Offset:%d", op.Inode, op.Handle, op.Offset))
	fileBuffer := fs.fileHandleCache.Get(op.Inode)
	if fileBuffer == nil {
		fs.fileHandleCache.Put(op.Inode, op.Handle, fs)
		fileBuffer = fs.fileHandleCache.Get(op.Inode)
	}

	inode, err := fs.GetInode(op.Inode)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[ReadFile]inodeID:%d, handleID:%d, err:%v", op.Inode, op.Handle, err))
		return err
	}
	filesize := inode.size
	if int64(filesize) <= op.Offset {
		op.BytesRead = 0
		klog.Infof(fmt.Sprintf("[ReadFile]read offset:%d is large than filesize:%d", op.Offset, filesize))
		return nil
	}
	bytesRead := int64(len(op.Dst))
	if op.Offset+bytesRead > int64(filesize) {
		bytesRead = int64(filesize) - op.Offset
	}
	op.BytesRead, err = fileBuffer.ReadFile(op.Offset, op.Dst[0:bytesRead], filesize, true)
	if err != nil {
		return fuse.EIO
	}

	return nil
}

// CreateFile INFO: 创建文件，其实是在 meta partition 中新建 inode/dentry 对象, `touch globalmount/2.txt`
func (fs *FuseFS) CreateFile(ctx context.Context, op *fuseops.CreateFileOp) error {
	if fs.metaClient.IsVolumeReadOnly() {
		return syscall.EROFS
	}
	if len(op.Name) >= NameMaxLen {
		return syscall.ENAMETOOLONG
	}

	// 在 meta partition 中写 inode 和 dentry 数据
	inodeInfo, err := fs.metaClient.CreateInodeAndDentry(op.Parent, op.Name,
		uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[CreateFile]create inode/dentry for %d/%s err %v", op.Parent, op.Name, err))
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
	fs.fileHandleCache.Put(child.inodeID, op.Handle, fs)
	parent, err := fs.GetInode(op.Parent)
	if err == nil {
		parent.dentryCache.Put(op.Name, inodeInfo.Inode)
	}

	klog.Infof(fmt.Sprintf("[CreateFile]create filename:%s, parentInodeID:%d, handleID:%d successfully", op.Name, op.Parent, op.Handle))
	return nil
}

//func (fs *FuseFS) WriteFile(ctx context.Context, op *fuseops.WriteFileOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) SyncFile(ctx context.Context, op *fuseops.SyncFileOp) error {
//	panic("implement me")
//}
//

func (fs *FuseFS) FlushFile(ctx context.Context, op *fuseops.FlushFileOp) error {
	fileBuffer := fs.fileHandleCache.Get(op.Inode)
	if fileBuffer == nil {
		fs.fileHandleCache.Put(op.Inode, op.Handle, fs)
		fileBuffer = fs.fileHandleCache.Get(op.Inode)
	}

	fileBuffer.FlushFile()

	return nil
}

func (fs *FuseFS) ReleaseFileHandle(ctx context.Context, op *fuseops.ReleaseFileHandleOp) error {
	fs.fileHandleCache.Release(op.Handle)

	return nil
}

//
//func (fs *FuseFS) Rename(ctx context.Context, op *fuseops.RenameOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) ReadSymlink(ctx context.Context, op *fuseops.ReadSymlinkOp) error {
//	panic("implement me")
//}

//func (fs *FuseFS) RemoveXattr(ctx context.Context, op *fuseops.RemoveXattrOp) error {
//	return fuse.ENOSYS
//}
//
//// Get an extended attribute.
//func (fs *FuseFS) GetXattr(ctx context.Context, op *fuseops.GetXattrOp) error {
//	return fuse.ENOSYS
//}
//
//func (fs *FuseFS) ListXattr(ctx context.Context, op *fuseops.ListXattrOp) error {
//	return fuse.ENOSYS
//}
//
//func (fs *FuseFS) SetXattr(ctx context.Context, op *fuseops.SetXattrOp) error {
//	return fuse.ENOSYS
//}
//
//func (fs *FuseFS) Fallocate(ctx context.Context, op *fuseops.FallocateOp) error {
//	panic("implement me")
//}

// StatFS INFO: `stat ${MountPoint}`
func (fs *FuseFS) StatFS(ctx context.Context, op *fuseops.StatFSOp) error {
	const defaultMaxMetaPartitionInodeID uint64 = 1<<63 - 1 // 64位操作系统
	total, used, inodeCount := fs.metaClient.Statfs()
	op.BlockSize = uint32(DefaultBlksize)
	op.Blocks = total / uint64(DefaultBlksize)
	op.BlocksFree = (total - used) / uint64(DefaultBlksize)
	op.BlocksAvailable = op.BlocksFree
	op.IoSize = 1 << 20
	op.Inodes = defaultMaxMetaPartitionInodeID
	op.InodesFree = defaultMaxMetaPartitionInodeID - inodeCount

	klog.Infof(fmt.Sprintf("[StatFS]op: %+v", *op))
	return nil
}
