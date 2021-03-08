package client

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/dfs/pkg/backend"
	"k8s-lx1036/k8s/storage/dfs/pkg/client/meta"
	"k8s-lx1036/k8s/storage/dfs/pkg/util/config"

	"github.com/jacobsa/fuse/fuseops"

	"k8s.io/klog/v2"
)

type MountOption struct {
	Config     *config.Config
	MountPoint string
	//volname also s3 bucket
	Volname            string
	Owner              string
	Master             string
	Logpath            string
	Loglvl             string
	Profport           string
	IcacheTimeout      int64
	LookupValid        int64
	AttrValid          int64
	ReadRate           int64
	WriteRate          int64
	EnSyncWrite        int64
	BufSize            int64
	MaxMultiParts      int
	MaxCacheInode      int
	ReadDirBurst       int
	ReadDirLimit       int
	Rdonly             bool
	WriteCache         bool
	KeepCache          bool
	FullPathName       bool
	S3ObjectNameVerify bool

	S3Cfg *config.S3Config
}

type Super struct {
	cluster            string
	endpoint           string
	localIP            string
	volname            string
	owner              string
	ic                 *InodeCache
	hc                 *HandleCache
	readDirc           *ReadDirCache
	readDirLimiter     *rate.Limiter
	mw                 *meta.MetaWrapper
	orphan             *OrphanInodeList
	enSyncWrite        bool
	keepCache          bool
	fullPathName       bool
	s3ObjectNameVerify bool
	maxMultiParts      int
	HTTPServer         *http.Server

	s3 *backend.S3Backend

	mu           sync.RWMutex
	nextHandleID fuseops.HandleID
	fileHandles  map[fuseops.HandleID]*FileHandle

	bufferPool *BufferPool

	replicators *Ticket
}

func (s Super) StatFS(ctx context.Context, op *fuseops.StatFSOp) error {
	panic("implement me")
}

func (s Super) LookUpInode(ctx context.Context, op *fuseops.LookUpInodeOp) error {
	panic("implement me")
}

func (s Super) GetInodeAttributes(ctx context.Context, op *fuseops.GetInodeAttributesOp) error {
	panic("implement me")
}

func (s Super) SetInodeAttributes(ctx context.Context, op *fuseops.SetInodeAttributesOp) error {
	panic("implement me")
}

func (s Super) ForgetInode(ctx context.Context, op *fuseops.ForgetInodeOp) error {
	panic("implement me")
}

func (s Super) MkDir(ctx context.Context, op *fuseops.MkDirOp) error {
	panic("implement me")
}

func (s Super) MkNode(ctx context.Context, op *fuseops.MkNodeOp) error {
	panic("implement me")
}

func (s Super) CreateFile(ctx context.Context, op *fuseops.CreateFileOp) error {
	panic("implement me")
}

func (s Super) CreateLink(ctx context.Context, op *fuseops.CreateLinkOp) error {
	panic("implement me")
}

func (s Super) CreateSymlink(ctx context.Context, op *fuseops.CreateSymlinkOp) error {
	panic("implement me")
}

func (s Super) Rename(ctx context.Context, op *fuseops.RenameOp) error {
	panic("implement me")
}

func (s Super) RmDir(ctx context.Context, op *fuseops.RmDirOp) error {
	panic("implement me")
}

func (s Super) Unlink(ctx context.Context, op *fuseops.UnlinkOp) error {
	panic("implement me")
}

func (s Super) OpenDir(ctx context.Context, op *fuseops.OpenDirOp) error {
	panic("implement me")
}

func (s Super) ReadDir(ctx context.Context, op *fuseops.ReadDirOp) error {
	panic("implement me")
}

func (s Super) ReleaseDirHandle(ctx context.Context, op *fuseops.ReleaseDirHandleOp) error {
	panic("implement me")
}

func (s Super) OpenFile(ctx context.Context, op *fuseops.OpenFileOp) error {
	panic("implement me")
}

func (s Super) ReadFile(ctx context.Context, op *fuseops.ReadFileOp) error {
	panic("implement me")
}

func (s Super) WriteFile(ctx context.Context, op *fuseops.WriteFileOp) error {
	panic("implement me")
}

func (s Super) SyncFile(ctx context.Context, op *fuseops.SyncFileOp) error {
	panic("implement me")
}

func (s Super) FlushFile(ctx context.Context, op *fuseops.FlushFileOp) error {
	panic("implement me")
}

func (s Super) ReleaseFileHandle(ctx context.Context, op *fuseops.ReleaseFileHandleOp) error {
	panic("implement me")
}

func (s Super) ReadSymlink(ctx context.Context, op *fuseops.ReadSymlinkOp) error {
	panic("implement me")
}

func (s Super) RemoveXattr(ctx context.Context, op *fuseops.RemoveXattrOp) error {
	panic("implement me")
}

func (s Super) GetXattr(ctx context.Context, op *fuseops.GetXattrOp) error {
	panic("implement me")
}

func (s Super) ListXattr(ctx context.Context, op *fuseops.ListXattrOp) error {
	panic("implement me")
}

func (s Super) SetXattr(ctx context.Context, op *fuseops.SetXattrOp) error {
	panic("implement me")
}

func (s Super) Fallocate(ctx context.Context, op *fuseops.FallocateOp) error {
	panic("implement me")
}

func (s Super) Destroy() {
	s.mw.UnMountClient()
}

func NewSuper(opt *MountOption) (s *Super, err error) {
	s = new(Super)
	s.mw, err = meta.NewMetaWrapper(opt.Volname, opt.Owner, opt.Master)
	if err != nil {
		return nil, fmt.Errorf("NewMetaWrapper failed with err %v", err)
	}

	opt.S3Cfg.Endpoint = s.mw.S3Endpoint
	s.s3, err = backend.NewS3(opt.Volname, opt.S3Cfg)
	if err != nil {
		klog.Errorf("Unable to setup backend: %v", err)
		return nil, err
	}

	err = s.mw.RegisterClientInfo(opt.Volname, opt.S3Cfg.Version, opt.MountPoint)
	if err != nil {
		klog.Errorf("Register client information failed, error: %v", err)
		return nil, err
	}

	s.volname = opt.Volname
	s.endpoint = s.mw.S3Endpoint
	s.owner = opt.Owner
	s.cluster = s.mw.Cluster()
	s.localIP = s.mw.LocalIP()
	inodeExpiration := DefaultInodeExpiration
	if opt.IcacheTimeout >= 0 {
		inodeExpiration = time.Duration(opt.IcacheTimeout) * time.Second
	}
	if opt.LookupValid >= 0 {
		LookupValidDuration = time.Duration(opt.LookupValid) * time.Second
	}
	if opt.AttrValid >= 0 {
		AttrValidDuration = time.Duration(opt.AttrValid) * time.Second
	}
	if opt.EnSyncWrite > 0 {
		s.enSyncWrite = true
	}
	s.fullPathName = opt.FullPathName
	s.s3ObjectNameVerify = opt.S3ObjectNameVerify
	s.keepCache = opt.KeepCache
	s.maxMultiParts = opt.MaxMultiParts
	s.hc = NewHandleCache()
	maxCacheInode := DefaultMaxCacheInode
	if opt.MaxCacheInode > 0 {
		if opt.MaxCacheInode < DefaultMinCacheInode {
			maxCacheInode = DefaultMinCacheInode
		} else {
			maxCacheInode = opt.MaxCacheInode
		}
	}
	s.ic = NewInodeCache(inodeExpiration, maxCacheInode)
	s.orphan = NewOrphanInodeList()

	s.bufferPool = BufferPool{
		maxBuffers: uint64(opt.BufSize / BUF_SIZE),
	}.Init()

	s.fileHandles = make(map[fuseops.HandleID]*FileHandle)

	s.replicators = Ticket{Total: 128}.Init()

	s.readDirc = NewReadDirCache(inodeExpiration)
	readDirBurst := opt.ReadDirBurst
	if readDirBurst == 0 {
		readDirBurst = DefaultReadDirBurst
	}
	readDirLimit := opt.ReadDirLimit
	if readDirLimit == 0 {
		readDirLimit = DefaultReadDirLimit
	}
	s.readDirLimiter = rate.NewLimiter(rate.Limit(readDirLimit), readDirBurst)

	klog.Infof(`NewSuper: cluster(%v) s3endpoint(%v) volname(%v) icacheExpiration(%v) LookupValidDuration(%v) AttrValidDuration(%v)`,
		s.cluster, s.endpoint, s.volname, inodeExpiration, LookupValidDuration, AttrValidDuration)
	return s, nil
}
