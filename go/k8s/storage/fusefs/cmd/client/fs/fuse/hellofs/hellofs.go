package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"

	"k8s.io/klog/v2"
)

// NewHelloFS INFO: Create a file system with a fixed structure that looks like this:
//
//     hello
//     dir/
//         world
//
// Each file contains the string "Hello, world!".
func NewHelloFS() (fuse.Server, error) {
	fs := &helloFS{}

	return fuseutil.NewFileSystemServer(fs), nil
}

type helloFS struct {
	fuseutil.NotImplementedFileSystem
}

const (
	rootInode fuseops.InodeID = fuseops.RootInodeID + iota
	helloInode
	dirInode
	worldInode
)

const (
	DefaultBlksize = 1 << 20 // 1M
)

type inodeInfo struct {
	attributes fuseops.InodeAttributes

	// File or directory?
	dir bool

	// For directories, children.
	children []fuseutil.Dirent
}

// INFO:
//  hello("Hello, world!")
//  dir/
//    world("Hello, world!")

// We have a fixed directory structure.
var gInodeInfo = map[fuseops.InodeID]inodeInfo{
	// root
	rootInode: inodeInfo{
		attributes: fuseops.InodeAttributes{
			Nlink: 1,
			Mode:  0555 | os.ModeDir,
		},
		dir: true,
		children: []fuseutil.Dirent{
			fuseutil.Dirent{
				Offset: 1,
				Inode:  helloInode,
				Name:   "hello",
				Type:   fuseutil.DT_File,
			},
			fuseutil.Dirent{
				Offset: 2,
				Inode:  dirInode,
				Name:   "dir",
				Type:   fuseutil.DT_Directory,
			},
		},
	},

	// hello
	helloInode: inodeInfo{
		attributes: fuseops.InodeAttributes{
			Nlink: 1,
			Mode:  0444,
			Size:  uint64(len("Hello, world!")),
		},
	},

	// dir
	dirInode: inodeInfo{
		attributes: fuseops.InodeAttributes{
			Nlink: 1,
			Mode:  0555 | os.ModeDir,
		},
		dir: true,
		children: []fuseutil.Dirent{
			fuseutil.Dirent{
				Offset: 1,
				Inode:  worldInode,
				Name:   "world",
				Type:   fuseutil.DT_File,
			},
		},
	},

	// world
	worldInode: inodeInfo{
		attributes: fuseops.InodeAttributes{
			Nlink: 1,
			Mode:  0444,
			Size:  uint64(len("Hello, world!")),
		},
	},
}

func findChildInode(name string, children []fuseutil.Dirent) (fuseops.InodeID, error) {
	for _, e := range children {
		if e.Name == name {
			return e.Inode, nil
		}
	}

	return 0, fuse.ENOENT
}

// INFO: `stat /mnt/hellofs2`
func (fs *helloFS) StatFS(ctx context.Context, op *fuseops.StatFSOp) error {
	total, used := uint64(1073741824), uint64(6)
	op.BlockSize = uint32(DefaultBlksize)
	op.Blocks = total / uint64(DefaultBlksize)
	op.BlocksFree = (total - used) / uint64(DefaultBlksize)
	op.BlocksAvailable = op.BlocksFree
	op.IoSize = 1 << 20
	op.Inodes = 1 << 50
	op.InodesFree = op.Inodes

	klog.Infof(fmt.Sprintf("[StatFS]op %+v", *op))

	return nil
}

var (
	fMountPoint = flag.String("mountpoint", "", "Path to mount point.")
	fReadOnly   = flag.Bool("read_only", false, "Mount in read-only mode.")
	fDebug      = flag.Bool("debug", false, "Enable debug logging.")
)

/*
INFO:
	[root@stark12 liuxiang3]# ll /mnt/hellofs2
	total 1
	dr-xr-xr-x 1 root root  0 Jul 10 20:26 dir
	-r--r--r-- 1 root root 13 Jul 10 20:26 hello
	[root@stark12 liuxiang3]# ll /mnt/hellofs2/dir
	total 1
	-r--r--r-- 1 root root 13 Jul 10 20:26 world
	[root@stark12 liuxiang3]# cat /mnt/hellofs2/hello
	Hello, world!
	[root@stark12 liuxiang3]# cat /mnt/hellofs2/dir/world
	Hello, world!
*/

// TODO: 目前还不支持 mkdir

// mkdir -p /tmp/fuse/hellofs
// go run . --mountpoint=/tmp/fuse/hellofs
func main() {
	flag.Parse()

	// filesystem server
	server, err := NewHelloFS()
	if err != nil {
		klog.Fatalf("makeFS: %v", err)
	}

	// Mount the file system.
	if *fMountPoint == "" {
		klog.Fatalf("You must set --mountpoint.")
	}

	cfg := &fuse.MountConfig{
		ReadOnly: *fReadOnly,
		FSName:   "hellofs",
		Subtype:  "hellofs_subtype",
	}
	if *fDebug {
		cfg.DebugLogger = log.New(os.Stderr, "fuse: ", 0)
	}

	mountedFileSystem, err := fuse.Mount(*fMountPoint, server, cfg)
	if err != nil {
		klog.Fatalf("Mount: %v", err)
	}

	klog.Infof(fmt.Sprintf("fuse mount point %s successfully", *fMountPoint))
	// Wait for it to be unmounted.
	if err = mountedFileSystem.Join(context.Background()); err != nil {
		klog.Fatalf("Join: %v", err)
	}
}
