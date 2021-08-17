package fs

import (
	"context"

	"github.com/jacobsa/fuse/fuseops"
)

func (super *Super) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	return nil
}
