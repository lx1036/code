package wal

import (
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"k8s.io/klog/v2"
)

// filePipeline pipelines allocating disk space
type filePipeline struct {
	// dir to put files
	dir string
	// size of files to make, in bytes
	size int64
	// count number of files generated
	count int

	filec chan *fileutil.LockedFile
	errc  chan error
	donec chan struct{}
}

func newFilePipeline(dir string, fileSize int64) *filePipeline {
	fp := &filePipeline{
		dir:   dir,
		size:  fileSize,
		filec: make(chan *fileutil.LockedFile),
		errc:  make(chan error, 1),
		donec: make(chan struct{}),
	}

	go fp.run()

	return fp
}

func (fp *filePipeline) run() {
	defer close(fp.errc)
	for {
		f, err := fp.alloc()
		if err != nil {
			fp.errc <- err
			return
		}
		select {
		case fp.filec <- f:
		case <-fp.donec:
			os.Remove(f.Name())
			f.Close()
			return
		}
	}
}

func (fp *filePipeline) alloc() (f *fileutil.LockedFile, err error) {
	// count % 2 so this file isn't the same as the one last published
	fpath := filepath.Join(fp.dir, fmt.Sprintf("%d.tmp", fp.count%2))
	if f, err = fileutil.LockFile(fpath, os.O_CREATE|os.O_WRONLY, fileutil.PrivateFileMode); err != nil {
		return nil, err
	}
	if err = fileutil.Preallocate(f.File, fp.size, true); err != nil {
		klog.Errorf(fmt.Sprintf("failed to preallocate space when creating a new WAL size:%d err:%v", fp.size, err))
		f.Close()
		return nil, err
	}
	fp.count++
	return f, nil
}

func (fp *filePipeline) Close() error {
	close(fp.donec)
	return <-fp.errc
}
