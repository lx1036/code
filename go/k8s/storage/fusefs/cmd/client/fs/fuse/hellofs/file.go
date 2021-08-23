package main

import (
	"context"
	"io"
	"strings"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
)

func (fs *helloFS) OpenFile(ctx context.Context, op *fuseops.OpenFileOp) error {
	// Allow opening any file.
	return nil
}

func (fs *helloFS) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
	// Let io.ReaderAt deal with the semantics.
	reader := strings.NewReader("Hello, world!")

	var err error
	op.BytesRead, err = reader.ReadAt(op.Dst, op.Offset)

	// Special case: FUSE doesn't expect us to return io.EOF.
	if err == io.EOF {
		return nil
	}

	return err
}
