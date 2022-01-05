package client

import (
	"container/list"
	"fmt"
	"io"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/pkg/proto"
	"strconv"
	"sync"
	"sync/atomic"

	"k8s.io/klog/v2"
)

// INFO: buffer 就是封装了下 s3 api 调用
type Buffer struct {
	sync.RWMutex

	key     string // filename
	backend Backend

	inodeID uint64
	ref     int32 // 这个字段有何设计目的???

	//backend backend.Backend
	fs *FuseFS

	stopC   chan struct{}
	LRUList *list.List

	lastError error
}

// NewBuffer INFO: 该 Buffer 对象其实就是对 S3 数据的缓存
func NewBuffer(key string, backend Backend) *Buffer {
	return &Buffer{
		backend: backend,
		key:     key,
	}

	/*buffer := &Buffer{
		inodeID: inodeID,
		//blockSize:     fs.blockSize,
		//flushInterval: fs.bufFlushInterval,
		//blocks:        make(map[int64]*list.Element, 0),
		LRUList: list.New(),
		//dirtyBlocks:   make(map[int64]*list.Element, 0),
		//dirtyList:     list.New(),
		fs: fs,
		//gbuf:          gbuf,
		backend: fs.s3Client,
		//mergeBlock:    fs.mergeBlock,
		//flushC:        make(chan struct{}, 1),
		stopC: make(chan struct{}, 1),
	}

	//buffer.bgFlushWait = DefaultFlushWait
	//atomic.StoreInt64(&buffer.nextReadOffset, 0)
	//atomic.StoreInt64(&buffer.seqReadAmount, 0)
	//atomic.StoreInt64(&buffer.nextReadAheadOffset, 0)
	//atomic.StoreInt64(&buffer.dirtyAmount, 0)
	//atomic.StoreInt32(&buffer.bgStarted, 0)
	//atomic.StoreInt32(&buffer.flushing, 0)

	return buffer*/
}

// SetFilename INFO: @see FuseFS::newFileHandle() 里设置的 key，其实就是文件的 inodeID
func (buffer *Buffer) SetFilename(filename string) {
	buffer.key = filename
}

func (buffer *Buffer) IncRef() int32 {
	return atomic.AddInt32(&buffer.ref, 1)
}

// ReadFile INFO: 读取 S3 文件，读取到 data 中
func (buffer *Buffer) ReadFile(offset int64, data []byte, filesize uint64, direct bool) (bytesRead int, err error) {
	defer func() {
		if err == io.EOF {
			err = nil
		}
	}()

	// read data from backend if direct flag is set
	if direct {
		bytesRead, err = buffer.readDirect(offset, data)
		if err != nil && err != io.EOF {
			return 0, err
		}
	}

	// TODO: read data from buffer

	/*nNeed := len(data)
	aOffset := buffer.alignOffset(offset)
	for nNeed > 0 {
		blk := buffer.getBlock(aOffset)
		if blk != nil {

		} else {
			err = buffer.FlushFile()
			if err != nil {

			}

			r, err = buffer.readDirect(offset, data[nRead:])
			if err != nil && err != io.EOF {
				return
			}
		}

	}*/

	return bytesRead, nil
}

func (buffer *Buffer) readDirect(offset int64, data []byte) (int, error) {
	bytesRead, err := buffer.backend.Read(buffer.key, offset, data)
	if err != nil && err != io.EOF {
		return 0, err
	}

	return bytesRead, nil
}

func (buffer *Buffer) WriteFile(offset int64, data []byte, direct bool) (bytesWrite int, err error) {

	if direct {
		bytesWrite, err = buffer.writeDirect(offset, data)
		if err != nil {
			return 0, err
		}

		return
	}

	return
}

func (buffer *Buffer) writeDirect(offset int64, data []byte) (bytesWrite int, err error) {
	return buffer.backend.Write(buffer.key, offset, data)
}

func (buffer *Buffer) FlushFile() {

}

// fullPathName=true, s3 key 则是 path；否则是 inodeID
func (fs *FuseFS) getS3Key(inodeID fuseops.InodeID) (string, error) {
	if !fs.fullPathName {
		return strconv.FormatUint(uint64(inodeID), 10), nil
	}

	if uint64(inodeID) == proto.RootInode {
		return "", nil
	}

	inode, err := fs.GetInode(inodeID)
	if err != nil {
		return "", err
	}
	if len(inode.fullPathName) != 0 {
		return inode.fullPathName, nil
	}

	pInode := inode
	currentInodeID := pInode.inodeID
	for parentInodeID := pInode.parentInodeID; uint64(pInode.inodeID) != proto.RootInode; parentInodeID = pInode.parentInodeID {
		pInode, err = fs.GetInode(parentInodeID)
		if err != nil {
			klog.Errorf(fmt.Sprintf("[getS3Key]get inodeID:%d err:%v", parentInodeID, err))
			return "", err
		}

		name, ok := pInode.dentryCache.GetByInode(currentInodeID)
		if !ok {
			name, err = fs.metaClient.LookupName(parentInodeID, currentInodeID)
			if err != nil {
				klog.Errorf(fmt.Sprintf("[getS3Key]get inodeID:%d LookupName err:%v", parentInodeID, err))
				return "", err
			}
		}

		inode.fullPathName = name
	}

	return inode.fullPathName, nil
}
