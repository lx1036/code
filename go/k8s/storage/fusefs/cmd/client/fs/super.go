package fs

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"k8s-lx1036/k8s/storage/fusefs/pkg/backend"
	"k8s-lx1036/k8s/storage/fusefs/pkg/sdk/meta"

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

type FuseFS struct {
	sync.RWMutex

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

	s3Backend *backend.S3Backend

	// INFO: FUSE file
	fileHandles  map[fuseops.HandleID]*FileHandle
	nextHandleID fuseops.HandleID

	//bufferPool *BufferPool
	dataBuffers map[uint64]*Buffer // [inodeID]*Buffer

	//replicators *Ticket
}

// INFO: `stat ${mountOption.MountPoint}` 命令执行结果
func (fs *FuseFS) StatFS(ctx context.Context, op *fuseops.StatFSOp) error {
	total, used := fs.metaClient.Statfs()
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

func (fs *FuseFS) ForgetInode(ctx context.Context, op *fuseops.ForgetInodeOp) error {
	panic("implement me")
}

func (fs *FuseFS) Rename(ctx context.Context, op *fuseops.RenameOp) error {
	panic("implement me")
}

func (fs *FuseFS) ReadSymlink(ctx context.Context, op *fuseops.ReadSymlinkOp) error {
	panic("implement me")
}

func (fs *FuseFS) RemoveXattr(ctx context.Context, op *fuseops.RemoveXattrOp) error {
	panic("implement me")
}

// Get an extended attribute.
func (fs *FuseFS) GetXattr(ctx context.Context, op *fuseops.GetXattrOp) error {
	return fuse.ENOSYS
}

func (fs *FuseFS) ListXattr(ctx context.Context, op *fuseops.ListXattrOp) error {
	panic("implement me")
}

func (fs *FuseFS) SetXattr(ctx context.Context, op *fuseops.SetXattrOp) error {
	panic("implement me")
}

func (fs *FuseFS) Fallocate(ctx context.Context, op *fuseops.FallocateOp) error {
	panic("implement me")
}

func (fs *FuseFS) Destroy() {
	//fs.mw.UnMountClient()
}

func (fs *FuseFS) getS3Key(inodeID uint64) (string, error) {
	if !fs.fullPathName {
		return strconv.Itoa(int(inodeID)), nil
	}

	return "123", nil
}

// INFO: 不同操作对应其OP: https://k8s-lx1036/k8s/storage/fuse/blob/master/fuseutil/file_system.go#L135-L222

func NewFuseFS(opt *MountOption) (*FuseFS, error) {
	var err error
	fs := new(FuseFS)
	fs.metaClient, err = meta.NewMetaWrapper(opt.Volname, opt.Owner, opt.MasterAddr)
	if err != nil {
		return nil, fmt.Errorf("NewMetaWrapper failed with err %v", err)
	}

	fs.endpoint = fs.metaClient.S3Endpoint
	fs.cluster = fs.metaClient.Cluster()
	fs.localIP = fs.metaClient.LocalIP()

	s3Config := &backend.S3Config{
		Endpoint:         fs.metaClient.S3Endpoint, // S3Endpoint 是调用 meta cluster api 获取的，实际上数据存在 master cluster 中
		AccessKey:        opt.AccessKey,
		SecretKey:        opt.SecretKey,
		DisableSSL:       true,
		S3ForcePathStyle: true, // 必须为 true，这样 url 才是 http://S3Endpoint/bucket
	}
	fs.s3Backend, err = backend.NewS3Backend(opt.Volname, s3Config)
	if err != nil {
		klog.Errorf(fmt.Sprintf("[NewFuseFS]new s3 client err %v", err))
		return nil, err
	}

	return fs, nil
}
