package fs

import (
	"context"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
)

// MkDir Create a directory inode as a child of an existing directory inode.
// The kernel sends this in response to a mkdir(2) call.
func (super *Super) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {

	parentInodeID := op.Parent

	super.metaClient.Create_ll(uint64(parentInodeID), op.Name, uint32(op.Mode.Perm()), op.Uid, op.Gid, nil)

}

func (super *Super) MkNode(ctx context.Context, op *fuseops.MkNodeOp) error {
	panic("implement me")
}

func (super *Super) RmDir(ctx context.Context, op *fuseops.RmDirOp) error {
	panic("implement me")
}

func (super *Super) Unlink(ctx context.Context, op *fuseops.UnlinkOp) error {
	panic("implement me")
}

func (super *Super) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	panic("implement me")
}

func (super *Super) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	panic("implement me")
}

func (super *Super) ReleaseDirHandle(ctx context.Context, op *fuseops.ReleaseDirHandleOp) error {
	panic("implement me")
}

func (super *Super) CreateLink(ctx context.Context, op *fuseops.CreateLinkOp) error {
	panic("implement me")
}

func (super *Super) CreateSymlink(ctx context.Context, op *fuseops.CreateSymlinkOp) error {
	panic("implement me")
}

func (super *Super) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	return nil
}
