package client

import (
	"fmt"
	"k8s-lx1036/k8s/storage/fuse/fuseutil"
	"k8s-lx1036/k8s/storage/fusefs/cmd/client/meta"
	"k8s-lx1036/k8s/storage/fusefs/cmd/client/s3"
	"net/http"
	"os/user"
	"strconv"
	"sync"

	"golang.org/x/time/rate"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s.io/klog/v2"
)

// INFO: VFS(虚拟文件系统) = 超级块super_block + inode + dentry + vfsmount
//  inode 索引节点: 记录文件系统对象中的一般元数据，比如 kind/size/created/modified/sharing&permissions/指向存储该内容的磁盘区块的指针/文件分类
//  dentry 目录项: 记录文件系统对象在全局文件系统树中的位置，例如：open一个文件/home/xxx/yyy.txt，那么/、home、xxx、yyy.txt都是一个目录项，
//  VFS在查找的时候，根据一层一层的目录项找到对应的每个目录项的inode，那么沿着目录项进行操作就可以找到最终的文件。
//  **[Linux 虚拟文件系统四大对象：超级块、inode、dentry、file之间关系](https://www.eet-china.com/mp/a38145.html)**

// INFO: VFS 就是一层接口，定义了一些接口函数
//  VFS是一种软件机制，只存在于内存中，每次系统初始化期间Linux都会先在内存中构造一棵VFS的目录树（也就是源码中的namespace）。
//  VFS主要的作用是对上层应用屏蔽底层不同的调用方法，提供一套统一的调用接口，二是便于对不同的文件系统进行组织管理。
//  VFS提供了一个抽象层，将POSIX API接口与不同存储设备的具体接口实现进行了分离，使得底层的文件系统类型、设备类型对上层应用程序透明。
//  例如read，write，那么映射到VFS中就是sys_read，sys_write，那么VFS可以根据你操作的是哪个“实际文件系统”(哪个分区)来进行不同的实际的操作！这个技术也是很熟悉的“钩子结构”技术来处理的。
//  其实就是VFS中提供一个抽象的struct结构体，然后对于每一个具体的文件系统要把自己的字段和函数填充进去，这样就解决了异构问题（内核很多子系统都大量使用了这种机制）

const (
	DefaultPageSize        = 1 << 17 // 128KB
	DefaultBlksize         = 1 << 20 // 1M
	DefaultBufSize         = 5 << 30 //5GB
	DefaultBufDirtyMax     = 6 << 20 // 6MB
	DefaultPartBlocks      = 5
	DefaultFlushInterval   = 5
	DefaultFlushWait       = 30
	DefaultMaxNameLen      = uint32(256)
	DefaultBlkExpiration   = 10
	DefaultReadAheadSize   = 100 << 20 // 100MB
	DefaultReadThreshold   = 20 << 20  // 20MB
	DefautlBufFreeTime     = 300
	DefaultMinBufferBlocks = 15
	DefaultMaxReleaseCount = 400
)

type Config struct {
	MountPoint string `json:"mountPoint"`

	//volname is also s3 bucket
	Region    string `json:"region" default:"beijing"`
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

	Debug bool `json:"debug" default:"false"`
}

// INFO: inode operations
//  https://www.kernel.org/doc/html/latest/filesystems/vfs.html#struct-inode-operations
//  Directory Entry Cache (dcache): https://www.kernel.org/doc/html/latest/filesystems/vfs.html#directory-entry-cache-dcache
//  Inode Object: https://www.kernel.org/doc/html/latest/filesystems/vfs.html#the-inode-object

type FuseFS struct {
	sync.RWMutex

	fuseutil.NotImplementedFileSystem

	cluster  string
	endpoint string
	localIP  string
	volname  string
	owner    string
	uid      uint32
	gid      uint32

	inodeCache *InodeCache

	metaClient *meta.MetaClient
	s3Client   *s3.S3Client

	dirHandleCache  *DirHandleCache
	fileHandleCache *FileHandleCache

	//ic                 *InodeCache
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
	nextHandleID fuseops.HandleID

	//bufferPool *BufferPool
	//dataBuffers map[uint64]*Buffer // [inodeID]*Buffer

	//replicators *Ticket
}

// INFO: 不同操作对应其OP: https://k8s-lx1036/k8s/storage/fuse/blob/master/fuseutil/file_system.go#L135-L222

func NewFuseFS(opt *Config) (*FuseFS, error) {
	var err error
	value, _ := user.Current()
	uid, _ := strconv.ParseUint(value.Uid, 10, 32)
	gid, _ := strconv.ParseUint(value.Gid, 10, 32)
	fs := &FuseFS{
		fullPathName: opt.FullPathName,
		uid:          uint32(uid),
		gid:          uint32(gid),

		inodeCache:      NewInodeCache(),
		dirHandleCache:  NewDirHandleCache(),
		fileHandleCache: NewFileHandleCache(),
	}
	fs.metaClient, err = meta.NewMetaClient(opt.Volname, opt.Owner, opt.MasterAddr)
	if err != nil {
		return nil, fmt.Errorf("NewMetaWrapper failed with err %v", err)
	}

	//fs.endpoint = fs.metaClient.S3Endpoint
	//fs.cluster = fs.metaClient.Cluster()
	//fs.localIP = fs.metaClient.LocalIP()

	fs.s3Client, err = s3.NewS3Backend(opt.Volname, &s3.S3Config{
		Endpoint:         fs.metaClient.S3Endpoint, // S3Endpoint 是调用 meta cluster api 获取的，实际上数据存在 master cluster 中
		Region:           "beijing",
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
