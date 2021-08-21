package fs

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"k8s-lx1036/k8s/storage/sunfs/pkg/backend"
	"k8s-lx1036/k8s/storage/sunfs/pkg/sdk/meta"

	"golang.org/x/time/rate"
	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s.io/klog/v2"
)

type MountOption struct {
	MountPoint string `json:"mountPoint"`

	//volname is also s3 bucket
	Volname   string `json:"volName"`
	Owner     string `json:"owner"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`

	MasterAddr string `json:"masterAddr"`

	ReadRate    int64
	WriteRate   int64
	EnSyncWrite int64
	BufSize     int64 `json:"bufSize"`

	ReadOnly bool `json:"readOnly"`

	FullPathName bool `json:"fullPathName"`
}

type Super struct {
	cluster  string
	endpoint string
	localIP  string
	volname  string
	owner    string

	inodeCache *InodeCache

	//ic                 *InodeCache
	hc *HandleCache
	//readDirc           *ReadDirCache
	readDirLimiter *rate.Limiter
	metaClient     *meta.MetaWrapper
	//orphan             *OrphanInodeList
	enSyncWrite        bool
	keepCache          bool
	fullPathName       bool
	s3ObjectNameVerify bool
	maxMultiParts      int
	HTTPServer         *http.Server

	s3 *backend.S3Backend

	mu           sync.RWMutex
	nextHandleID fuseops.HandleID
	//fileHandles  map[fuseops.HandleID]*FileHandle

	//bufferPool *BufferPool

	//replicators *Ticket
}

// INFO: `stat ${mountOption.MountPoint}` 命令执行结果
func (super *Super) StatFS(ctx context.Context, op *fuseops.StatFSOp) error {
	total, used := super.metaClient.Statfs()
	op.BlockSize = uint32(DefaultBlksize)
	op.Blocks = total / uint64(DefaultBlksize)
	op.BlocksFree = (total - used) / uint64(DefaultBlksize)
	op.BlocksAvailable = op.BlocksFree
	op.IoSize = 1 << 20
	op.Inodes = 1 << 50
	op.InodesFree = op.Inodes

	klog.Infof(fmt.Sprintf("[StatFS]op: %+v", *op))
	return nil
}

func (super *Super) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	panic("implement me")
}

func (super *Super) SetInodeAttributes(ctx context.Context, op *fuseops.SetInodeAttributesOp) error {
	panic("implement me")
}

func (super *Super) ForgetInode(ctx context.Context, op *fuseops.ForgetInodeOp) error {
	panic("implement me")
}

func (super *Super) CreateFile(ctx context.Context, op *fuseops.CreateFileOp) error {
	panic("implement me")
}

func (super *Super) Rename(ctx context.Context, op *fuseops.RenameOp) error {
	panic("implement me")
}

func (super *Super) OpenFile(ctx context.Context, op *fuseops.OpenFileOp) error {
	panic("implement me")
}

func (super *Super) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
	panic("implement me")
}

func (super *Super) WriteFile(ctx context.Context, op *fuseops.WriteFileOp) error {
	panic("implement me")
}

func (super *Super) SyncFile(ctx context.Context, op *fuseops.SyncFileOp) error {
	panic("implement me")
}

func (super *Super) FlushFile(ctx context.Context, op *fuseops.FlushFileOp) error {
	panic("implement me")
}

func (super *Super) ReleaseFileHandle(ctx context.Context, op *fuseops.ReleaseFileHandleOp) error {
	panic("implement me")
}

func (super *Super) ReadSymlink(ctx context.Context, op *fuseops.ReadSymlinkOp) error {
	panic("implement me")
}

func (super *Super) RemoveXattr(ctx context.Context, op *fuseops.RemoveXattrOp) error {
	panic("implement me")
}

// Get an extended attribute.
func (super *Super) GetXattr(ctx context.Context, op *fuseops.GetXattrOp) error {
	return fuse.ENOSYS
}

func (super *Super) ListXattr(ctx context.Context, op *fuseops.ListXattrOp) error {
	panic("implement me")
}

func (super *Super) SetXattr(ctx context.Context, op *fuseops.SetXattrOp) error {
	panic("implement me")
}

func (super *Super) Fallocate(ctx context.Context, op *fuseops.FallocateOp) error {
	panic("implement me")
}

func (super *Super) Destroy() {
	//super.mw.UnMountClient()
}

// INFO: 不同操作对应其OP: https://k8s-lx1036/k8s/storage/fuse/blob/master/fuseutil/file_system.go#L135-L222

func NewSuper(opt *MountOption) (*Super, error) {
	var err error
	super := new(Super)
	super.metaClient, err = meta.NewMetaWrapper(opt.Volname, opt.Owner, opt.MasterAddr)
	if err != nil {
		return nil, fmt.Errorf("NewMetaWrapper failed with err %v", err)
	}

	super.endpoint = super.metaClient.S3Endpoint
	super.cluster = super.metaClient.Cluster()
	super.localIP = super.metaClient.LocalIP()

	return super, nil
}
