package fs

import (
	"context"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
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
		buffer = NewBuffer()
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

func (fs *FuseFS) CreateFile(ctx context.Context, op *fuseops.CreateFileOp) error {
	panic("implement me")
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
