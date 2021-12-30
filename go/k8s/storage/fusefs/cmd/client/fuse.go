package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"k8s-lx1036/k8s/storage/fusefs/pkg/sdk/meta"

	"golang.org/x/time/rate"
	"k8s-lx1036/k8s/storage/fuse"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s.io/klog/v2"
)

type Config struct {
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

// INFO: inode operations
//  https://www.kernel.org/doc/html/latest/filesystems/vfs.html#struct-inode-operations
//  Directory Entry Cache (dcache): https://www.kernel.org/doc/html/latest/filesystems/vfs.html#directory-entry-cache-dcache
//  Inode Object: https://www.kernel.org/doc/html/latest/filesystems/vfs.html#the-inode-object

type FuseFS struct {
	sync.RWMutex

	cluster    string
	endpoint   string
	localIP    string
	volname    string
	owner      string
	inodeCache *InodeCache

	metaClient *MetaClient
	s3Client   *S3Client

	//ic                 *InodeCache
	hc *HandleCache
	//readDirc           *ReadDirCache
	readDirLimiter *rate.Limiter
	//orphan             *OrphanInodeList
	enSyncWrite        bool
	keepCache          bool
	fullPathName       bool
	s3ObjectNameVerify bool
	maxMultiParts      int
	HTTPServer         *http.Server

	// INFO: FUSE file
	fileHandles  map[fuseops.HandleID]*FileHandle
	nextHandleID fuseops.HandleID

	//bufferPool *BufferPool
	dataBuffers map[uint64]*Buffer // [inodeID]*Buffer

	//replicators *Ticket
}

// INFO: 不同操作对应其OP: https://k8s-lx1036/k8s/storage/fuse/blob/master/fuseutil/file_system.go#L135-L222

func NewFuseFS(opt *Config) (*FuseFS, error) {
	var err error
	fs := new(FuseFS)
	fs.metaClient, err = NewMetaClient(opt.Volname, opt.Owner, opt.MasterAddr)
	if err != nil {
		return nil, fmt.Errorf("NewMetaWrapper failed with err %v", err)
	}

	fs.endpoint = fs.metaClient.S3Endpoint
	fs.cluster = fs.metaClient.Cluster()
	fs.localIP = fs.metaClient.LocalIP()

	fs.s3Client, err = NewS3Backend(opt.Volname, &S3Config{
		Endpoint:         fs.metaClient.S3Endpoint, // S3Endpoint 是调用 meta cluster api 获取的，实际上数据存在 master cluster 中
		AccessKey:        opt.AccessKey,
		SecretKey:        opt.SecretKey,
		DisableSSL:       true,
		S3ForcePathStyle: true, // 必须为 true，这样 url 才是 http://S3Endpoint/bucket
	})
	if err != nil {
		klog.Errorf(fmt.Sprintf("[NewFuseFS]new s3 client err %v", err))
		return nil, err
	}

	return fs, nil
}

func (fs *FuseFS) Destroy() {
	//fs.mw.UnMountClient()
}
