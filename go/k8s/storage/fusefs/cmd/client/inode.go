package client

import (
	"container/list"
	"context"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
)

var (
	// The following two are used in the FUSE cache
	// every time the lookup will be performed on the fly, and the result will not be cached
	LookupValidDuration = 300 * time.Second

	// the expiration duration of the attributes in the FUSE cache
	AttrValidDuration = 300 * time.Second // 5min
)

type Inode struct {
	inodeID       fuseops.InodeID
	parentInodeID fuseops.InodeID

	size  uint64
	nlink uint32
	uid   uint32
	gid   uint32
	gen   uint64

	createTime int64 // time of last inode change
	modifyTime int64 // time of last modification
	accessTime int64 // time of last access

	mode   os.FileMode
	target []byte

	fullPathName string

	expiration int64 // nano second, 过期时间

	// For directory inode only
	dentryCache *DentryCache
}

func (inode *Inode) setExpiration(t time.Duration) {
	inode.expiration = time.Now().Add(t).UnixNano()
}

func (inode *Inode) expired() bool {
	// root inode never expire
	if inode.inodeID != fuseops.RootInodeID && time.Now().UnixNano() > inode.expiration {
		return true
	}

	return false
}

func NewInode(inodeInfo *proto.InodeInfo) *Inode {
	inode := &Inode{
		inodeID:       fuseops.InodeID(inodeInfo.Inode),
		parentInodeID: fuseops.InodeID(inodeInfo.PInode),
		size:          inodeInfo.Size,
		nlink:         inodeInfo.Nlink,
		uid:           inodeInfo.Uid,
		gid:           inodeInfo.Gid,
		gen:           inodeInfo.Generation,
		createTime:    inodeInfo.CreateTime,
		modifyTime:    inodeInfo.ModifyTime,
		accessTime:    inodeInfo.AccessTime,
		mode:          os.FileMode(inodeInfo.Mode),
		target:        inodeInfo.Target,
		fullPathName:  "",
	}

	if proto.IsDir(inodeInfo.Mode) {
		inode.dentryCache = NewDentryCache()
	}

	return inode
}

const (
	MinInodeCacheEvictNum = 100
)

// InodeCache
// INFO: 这里使用了 LRU 数据结构，并且每一个 item 都有过期时间
//  最新使用的置前，60min 后会 evict 过期的。然后再 tcp 从 meta partition cluster 重新获取新的 inode info，
//  然后每 AttrValidDuration 5min 内核会检查该 inode Attributes 会过期，重新调用 GetInodeAttributes 来刷新，见 GetInodeAttributes() 函数
type InodeCache struct {
	sync.RWMutex

	// INFO: LRU = Map + DoublyLinkedList
	cache   map[fuseops.InodeID]*list.Element // map[inodeID]Element
	lruList *list.List                        // a doubly linked list

	maxElements int           // 双向链表元素最大数量, 默认设置 1000
	expiration  time.Duration // 过期时间，默认设置 60min
}

func NewInodeCache() *InodeCache {
	inodeCache := &InodeCache{
		maxElements: 1000,
		expiration:  time.Minute * 60,

		cache:   make(map[fuseops.InodeID]*list.Element),
		lruList: new(list.List),
	}

	go inodeCache.start()

	return inodeCache
}

func (inodeCache *InodeCache) start() {
	wait.UntilWithContext(context.TODO(), func(ctx context.Context) {
		inodeCache.evict()
	}, time.Minute*10)
}

// 检查 LRU 后 100 元素，删除其中过期的
func (inodeCache *InodeCache) evict() {
	inodeCache.Lock()
	defer inodeCache.Unlock()

	for i := 0; i < MinInodeCacheEvictNum; i++ {
		element := inodeCache.lruList.Back()
		if element == nil {
			return
		}
		inode := element.Value.(*Inode)
		if !inode.expired() {
			continue
		}

		delete(inodeCache.cache, inode.inodeID)
		inodeCache.lruList.Remove(element)
	}
}

// Put INFO: 把 inode 存入 LRU 数据结构中
func (inodeCache *InodeCache) Put(inode *Inode) {
	inodeCache.Lock()
	defer inodeCache.Unlock()

	old, ok := inodeCache.cache[inode.inodeID]
	if ok {
		inodeCache.lruList.Remove(old)
		delete(inodeCache.cache, inode.inodeID)
	}
	inode.setExpiration(inodeCache.expiration)
	element := inodeCache.lruList.PushFront(inode)
	inodeCache.cache[inode.inodeID] = element

	if inodeCache.lruList.Len() >= inodeCache.maxElements {
		inodeCache.evict()
	}
}

// Get INFO: 从 map 中取，这是 LRU 性能好的重要一个原因
func (inodeCache *InodeCache) Get(inodeID fuseops.InodeID) *Inode {
	inodeCache.Lock()
	defer inodeCache.Unlock()

	element, ok := inodeCache.cache[inodeID]
	if !ok {
		return nil
	}

	inode := element.Value.(*Inode)
	if inode.expired() {
		return nil
	}

	return inode
}

// GetInode INFO: 从本地缓存 InodeCache 取值，如果没有调用 meta cluster api 获取并存入 InodeCache
func (fs *FuseFS) GetInode(inodeID fuseops.InodeID) (*Inode, error) {
	inode := fs.inodeCache.Get(inodeID)
	if inode != nil {
		return inode, nil
	}

	// 本地缓存里没有，从 meta cluster 中取
	inodeInfo, err := fs.metaClient.GetInode(inodeID)
	if err != nil {
		return nil, err
	}

	inode = NewInode(inodeInfo)
	fs.inodeCache.Put(inode)

	return inode, nil
}

func GetChildInodeEntry(child *Inode) fuseops.ChildInodeEntry {
	return fuseops.ChildInodeEntry{
		Child: fuseops.InodeID(child.inodeID),
		Attributes: fuseops.InodeAttributes{
			Size:   child.size,
			Nlink:  child.nlink,
			Mode:   child.mode,
			Atime:  time.Unix(child.accessTime, 0),
			Mtime:  time.Unix(child.modifyTime, 0),
			Ctime:  time.Unix(child.createTime, 0),
			Crtime: time.Time{},
			Uid:    child.uid,
			Gid:    child.gid,
		},
		AttributesExpiration: time.Now().Add(AttrValidDuration),
		EntryExpiration:      time.Now().Add(LookupValidDuration),
	}
}
