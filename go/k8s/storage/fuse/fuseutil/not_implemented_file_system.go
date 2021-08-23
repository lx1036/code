package fuseutil

import (
	"context"
	"k8s.io/klog/v2"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
)

// A FileSystem that responds to all ops with fuse.ENOSYS. Embed this in your
// struct to inherit default implementations for the methods you don't care
// about, ensuring your struct will continue to implement FileSystem even as
// new methods are added.
type NotImplementedFileSystem struct {
}

var _ FileSystem = &NotImplementedFileSystem{}

func (fs *NotImplementedFileSystem) StatFS(
	ctx context.Context,
	op *fuseops.StatFSOp) error {
	klog.Info("StatFS")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) LookUpInode(
	ctx context.Context,
	op *fuseops.LookUpInodeOp) error {
	klog.Info("LookUpInode")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) GetInodeAttributes(
	ctx context.Context,
	op *fuseops.GetInodeAttributesOp) error {
	klog.Info("GetInodeAttributes")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) SetInodeAttributes(
	ctx context.Context,
	op *fuseops.SetInodeAttributesOp) error {
	klog.Info("SetInodeAttributes")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) ForgetInode(
	ctx context.Context,
	op *fuseops.ForgetInodeOp) error {
	klog.Info("ForgetInode")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) MkDir(
	ctx context.Context,
	op *fuseops.MkDirOp) error {
	klog.Info("MkDir")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) MkNode(
	ctx context.Context,
	op *fuseops.MkNodeOp) error {
	klog.Info("MkNode")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) CreateFile(
	ctx context.Context,
	op *fuseops.CreateFileOp) error {
	klog.Info("CreateFile")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) CreateSymlink(
	ctx context.Context,
	op *fuseops.CreateSymlinkOp) error {
	klog.Info("CreateSymlink")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) CreateLink(
	ctx context.Context,
	op *fuseops.CreateLinkOp) error {
	klog.Info("CreateLink")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) Rename(
	ctx context.Context,
	op *fuseops.RenameOp) error {
	klog.Info("Rename")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) RmDir(
	ctx context.Context,
	op *fuseops.RmDirOp) error {
	klog.Info("RmDir")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) Unlink(
	ctx context.Context,
	op *fuseops.UnlinkOp) error {
	klog.Info("Unlink")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) OpenDir(
	ctx context.Context,
	op *fuseops.OpenDirOp) error {
	klog.Info("OpenDir")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) ReadDir(
	ctx context.Context,
	op *fuseops.ReadDirOp) error {
	klog.Info("ReadDir")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) ReleaseDirHandle(
	ctx context.Context,
	op *fuseops.ReleaseDirHandleOp) error {
	klog.Info("ReleaseDirHandle")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) OpenFile(
	ctx context.Context,
	op *fuseops.OpenFileOp) error {
	klog.Info("OpenFile")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) ReadFile(
	ctx context.Context,
	op *fuseops.ReadFileOp) error {
	klog.Info("ReadFile")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) WriteFile(
	ctx context.Context,
	op *fuseops.WriteFileOp) error {
	klog.Info("WriteFile")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) SyncFile(
	ctx context.Context,
	op *fuseops.SyncFileOp) error {
	klog.Info("SyncFile")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) FlushFile(
	ctx context.Context,
	op *fuseops.FlushFileOp) error {
	klog.Info("FlushFile")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) ReleaseFileHandle(
	ctx context.Context,
	op *fuseops.ReleaseFileHandleOp) error {
	klog.Info("ReleaseFileHandle")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) ReadSymlink(
	ctx context.Context,
	op *fuseops.ReadSymlinkOp) error {
	klog.Info("ReadSymlink")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) RemoveXattr(
	ctx context.Context,
	op *fuseops.RemoveXattrOp) error {
	klog.Info("RemoveXattr")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) GetXattr(
	ctx context.Context,
	op *fuseops.GetXattrOp) error {
	klog.Info("GetXattr")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) ListXattr(
	ctx context.Context,
	op *fuseops.ListXattrOp) error {
	klog.Info("ListXattr")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) SetXattr(
	ctx context.Context,
	op *fuseops.SetXattrOp) error {
	klog.Info("SetXattr")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) Fallocate(
	ctx context.Context,
	op *fuseops.FallocateOp) error {
	klog.Info("Fallocate")
	return fuse.ENOSYS
}

func (fs *NotImplementedFileSystem) Destroy() {
	klog.Info("Destroy")
}
