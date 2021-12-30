package client

import (
	"container/list"
	"os"
	"sync"
	"time"

	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
)

var (
	// The following two are used in the FUSE cache
	// every time the lookup will be performed on the fly, and the result will not be cached
	LookupValidDuration = 5 * time.Second
	// the expiration duration of the attributes in the FUSE cache
	AttrValidDuration = 30 * time.Second
)

type Inode struct {
	inodeID     uint64
	parentInode uint64
	size        uint64
	nlink       uint32
	uid         uint32
	gid         uint32
	gen         uint64
	createTime  int64 // time of last inode change
	modifyTime  int64 // time of last modification
	accessTime  int64 // time of last access
	mode        os.FileMode
	target      []byte

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
		inodeID:      inodeInfo.Inode,
		parentInode:  inodeInfo.PInode,
		size:         inodeInfo.Size,
		nlink:        inodeInfo.Nlink,
		uid:          inodeInfo.Uid,
		gid:          inodeInfo.Gid,
		gen:          inodeInfo.Generation,
		createTime:   inodeInfo.CreateTime,
		modifyTime:   inodeInfo.ModifyTime,
		accessTime:   inodeInfo.AccessTime,
		mode:         os.FileMode(inodeInfo.Mode),
		target:       inodeInfo.Target,
		fullPathName: "",
	}

	if proto.IsDir(inodeInfo.Mode) {
		inode.dentryCache = NewDentryCache()
	}

	return inode
}

// INFO: 这里使用了 LRU 数据结构
type InodeCache struct {
	sync.RWMutex

	// INFO: LRU = Map + DoublyLinkedList
	cache       map[uint64]*list.Element // map[inodeID]Element
	lruList     *list.List               // a doubly linked list
	maxElements int                      // 双向链表元素最大数量

	expiration time.Duration // 过期时间

}

// INFO: 把 inode 存入 LRU 数据结构中
func (inodeCache *InodeCache) Put(inode *Inode) {
	inodeCache.Lock()
	defer inodeCache.Unlock()

	old, ok := inodeCache.cache[inode.inodeID]
	if ok {
		inodeCache.lruList.Remove(old)
		delete(inodeCache.cache, inode.inodeID)
	}

	if inodeCache.lruList.Len() >= inodeCache.maxElements {
		inodeCache.evict(true)
	}

	inode.setExpiration(inodeCache.expiration)
	element := inodeCache.lruList.PushFront(inode)
	inodeCache.cache[inode.inodeID] = element
}

// INFO: 从 map 中取，这是 LRU 性能好的重要一个原因
func (inodeCache *InodeCache) Get(inodeID uint64) *Inode {
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
