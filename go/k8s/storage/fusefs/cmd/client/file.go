package client

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/storage/fuse"
	"sync"

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

func (fs *FuseFS) OpenFile(ctx context.Context, op *fuseops.OpenFileOp) error {
	fs.fileHandleCache.Put(op.Inode, op.Handle, fs)
	return nil
}

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

// CreateFile INFO: 创建文件，其实是在 meta partition 中新建 inode/dentry 对象
//func (fs *FuseFS) CreateFile(ctx context.Context, op *fuseops.CreateFileOp) error {
//	if fs.metaClient.IsVolumeReadOnly() {
//		return syscall.EROFS
//	}
//	if len(op.Name) >= NameMaxLen {
//		return syscall.ENAMETOOLONG
//	}
//
//	// 在 meta partition 中写 inode 和 dentry 数据
//	parentInodeID := op.Parent
//	inodeInfo, err := fs.metaClient.CreateInodeAndDentry(op.Parent, op.Name,
//		uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)
//	if err != nil {
//		klog.Errorf(fmt.Sprintf("[CreateFile]create inode/dentry for %d/%s err %v", uint64(parentInodeID), op.Name, err))
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
//	// INFO: 需要填写 child entry，这里注意用的指针 &op.Entry，这样可以直接修改 op.Entry 属性值
//	//fillChildEntry(&op.Entry, child)
//	op.Entry = GetChildInodeEntry(child)
//
//	/*handle, err := fs.newFileHandle(child.inodeID, 0)
//	if err != nil {
//		klog.Errorf(fmt.Sprintf("[CreateFile]newFileHandle err %v", err))
//		return err
//	}
//	op.Handle = handle*/
//
//	klog.Infof(fmt.Sprintf("[CreateFile]create filename %s, parent inodeID %d successfully", op.Name, op.Parent))
//	return nil
//}

//func (fs *FuseFS) WriteFile(ctx context.Context, op *fuseops.WriteFileOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) SyncFile(ctx context.Context, op *fuseops.SyncFileOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) FlushFile(ctx context.Context, op *fuseops.FlushFileOp) error {
//	panic("implement me")
//}
//
//func (fs *FuseFS) ReleaseFileHandle(ctx context.Context, op *fuseops.ReleaseFileHandleOp) error {
//	panic("implement me")
//}
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
