package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	
	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
	
	"k8s.io/klog/v2"
)

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

type helloFS struct {
	fuseutil.NotImplementedFileSystem
}

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

//func (fs *helloFS) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
//	info, ok := gInodeInfo[op.Inode]
//	if !ok {
//		return fuse.ENOENT
//	}
//
//	if !info.attributes.Mode.IsDir() {
//		return fmt.Errorf(fmt.Sprintf("[OpenDir]inodeID:%d is not dir", op.Inode))
//	}
//
//	return nil
//}

// ReadDir INFO: `ll /tmp/fuse/hellofs` read dir
//func (fs *helloFS) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
//	// Find the info for this inode.
//	info, ok := gInodeInfo[op.Inode]
//	if !ok {
//		return fuse.ENOENT
//	}
//
//	klog.Infof(fmt.Sprintf("[ReadDirOp]Inode: %+v, InodeInfo: %+v", op.Inode, info))
//
//	if !info.dir {
//		return fuse.EIO
//	}
//
//	entries := info.children
//	// Grab the range of interest.
//	if op.Offset > fuseops.DirOffset(len(entries)) {
//		return fuse.EIO
//	}
//
//	entries = entries[op.Offset:]
//	// Resume at the specified offset into the array.
//	for _, e := range entries {
//		n := fuseutil.WriteDirent(op.Dst[op.BytesRead:], e)
//		if n == 0 {
//			break
//		}
//
//		op.BytesRead += n
//	}
//
//	return nil
//}
//
//func (fs *helloFS) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
//	// Let io.ReaderAt deal with the semantics.
//	reader := strings.NewReader("Hello, world!")
//
//	var err error
//	op.BytesRead, err = reader.ReadAt(op.Dst, op.Offset)
//
//	// Special case: FUSE doesn't expect us to return io.EOF.
//	if err == io.EOF {
//		return nil
//	}
//
//	return err
//}
//
//func (fs *helloFS) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
//	// Find the info for the parent.
//	parentInfo, ok := gInodeInfo[op.Parent]
//	if !ok {
//		return fuse.ENOENT
//	}
//
//	// Find the child within the parent.
//	var childInode fuseops.InodeID
//	found := false
//	for _, child := range parentInfo.children {
//		if child.Name == op.Name {
//			childInode = child.Inode
//			found = true
//			break
//		}
//	}
//	if !found {
//		return fuse.ENOENT
//	}
//
//	op.Entry.Child = childInode
//	op.Entry.Attributes = gInodeInfo[childInode].attributes
//	fs.patchAttributes(&op.Entry.Attributes)
//
//	klog.Infof(fmt.Sprintf("[LookUpInode]inodeID:%d", op.Parent))
//	return nil
//}

func (fs *helloFS) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	if op.OpContext.Pid == 0 {
		// CreateFileOp should have a valid pid in context.
		return fuse.EINVAL
	}

	klog.Infof(fmt.Sprintf("[GetInodeAttributes]inodeID:%d", op.Inode))

	// Find the info for this inode.
	info, ok := gInodeInfo[op.Inode]
	if !ok {
		return fuse.ENOENT
	}

	// Copy over its attributes.
	op.Attributes = info.attributes
	op.AttributesExpiration = time.Now().Add(365 * 24 * time.Hour)

	// Patch attributes.
	fs.patchAttributes(&op.Attributes)

	return nil
}

func (fs *helloFS) patchAttributes(attr *fuseops.InodeAttributes) {
	now := time.Now()
	attr.Atime = now
	attr.Mtime = now
	attr.Crtime = now
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
	mountPoint = flag.String("mountpoint", "", "Path to mount point.")
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

	if len(*mountPoint) == 0 {
		klog.Fatalf("You must set --mountpoint.")
	}

	//server, err := NewHelloFS()
	server, err := NewHelloFS()
	if err != nil {
		klog.Fatalf("makeFS: %v", err)
	}
	mountPath, _ := filepath.Abs(*mountPoint)
	mountedFileSystem, err := fuse.Mount(mountPath, server, &fuse.MountConfig{
		FSName:                  "hellofs",
		Subtype:                 "hellofs_subtype",
		ReadOnly:                false,
		DisableWritebackCaching: true,
		DebugLogger:             log.New(os.Stderr, "fuse: ", log.LstdFlags),
	})
	if err != nil {
		klog.Fatalf("Mount: %v", err)
	}

	klog.Infof(fmt.Sprintf("fuse mount point %s successfully", mountPath))
	// Wait for it to be unmounted.
	if err = mountedFileSystem.Join(context.Background()); err != nil {
		klog.Fatalf("Join: %v", err)
	}
}
