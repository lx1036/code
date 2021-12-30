package client

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"

	"k8s.io/klog/v2"
)

type FileHandle struct {
	inodeID uint64
	flag    uint32
}

func (fs *FuseFS) newFileHandle(inodeID uint64, flag uint32) (fuseops.HandleID, error) {
	fs.Lock()
	defer fs.Unlock()

	key, err := fs.getS3Key(inodeID)
	if err != nil {
		return 0, err
	}

	buffer, ok := fs.dataBuffers[inodeID]
	if !ok {
		buffer = NewBuffer(inodeID, fs.s3Backend, fs)
		buffer.SetFilename(key)
		fs.dataBuffers[inodeID] = buffer
	}
	buffer.IncRef()

	handleID := fs.nextHandleID
	fs.nextHandleID++
	fs.fileHandles[handleID] = &FileHandle{
		inodeID: inodeID,
		flag:    flag,
	}

	return handleID, nil
}

// INFO: 创建文件，其实是在 meta partition 中新建 inode/dentry 对象
func (fs *FuseFS) CreateFile(ctx context.Context, op *fuseops.CreateFileOp) error {
	// 在 meta partition 中写 inode 和 dentry 数据
	parentInodeID := op.Parent
	inodeInfo, err := fs.metaClient.Create_ll(uint64(parentInodeID), op.Name, uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[CreateFile]create inode/dentry for %d/%s err %v", uint64(parentInodeID), op.Name, err))
		return err
	}

	child := NewInode(inodeInfo)
	fs.inodeCache.Put(child)
	parent, err := fs.InodeGet(uint64(parentInodeID))
	if err == nil {
		parent.dentryCache.Put(op.Name, inodeInfo.Inode)
	}

	// INFO: 需要填写 child entry，这里注意用的指针 &op.Entry，这样可以直接修改 op.Entry 属性值
	//fillChildEntry(&op.Entry, child)
	op.Entry = GetChildInodeEntry(child)

	handle, err := fs.newFileHandle(child.inodeID, 0)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[CreateFile]newFileHandle err %v", err))
		return err
	}
	op.Handle = handle

	klog.Infof(fmt.Sprintf("[CreateFile]create filename %s, parent inodeID %d successfully", op.Name, op.Parent))
	return nil
}

func (fs *FuseFS) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
	fs.Lock()
	defer fs.Unlock()

	var buf *Buffer
	if fileHandle, ok := fs.fileHandles[op.Handle]; ok {
		if buf, ok = fs.dataBuffers[fileHandle.inodeID]; !ok || buf.lastError != nil {
			return fuse.EIO
		}
	} else {
		return fuse.EIO
	}

	// read data from buffer
	inode, err := fs.InodeGet(buf.inodeID)
	if err != nil {
		return err
	}

	op.BytesRead, err = buf.ReadFile(op.Offset, op.Dst[0:rNeed], fileSize, false)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FuseFS) OpenFile(ctx context.Context, op *fuseops.OpenFileOp) error {
	panic("implement me")
}

func (fs *FuseFS) WriteFile(ctx context.Context, op *fuseops.WriteFileOp) error {
	panic("implement me")
}

func (fs *FuseFS) SyncFile(ctx context.Context, op *fuseops.SyncFileOp) error {
	panic("implement me")
}

func (fs *FuseFS) FlushFile(ctx context.Context, op *fuseops.FlushFileOp) error {
	panic("implement me")
}

func (fs *FuseFS) ReleaseFileHandle(ctx context.Context, op *fuseops.ReleaseFileHandleOp) error {
	panic("implement me")
}
