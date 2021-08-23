package fs

import (
	"container/list"
	"io"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/fusefs/pkg/backend"
)

// INFO: buffer 就是封装了下 s3 api 调用
type Buffer struct {
	sync.RWMutex

	key     string // filename
	inodeID uint64
	ref     int32

	backend backend.Backend

	stopC   chan struct{}
	LRUList *list.List

	lastError error
}

func (buffer *Buffer) SetFilename(filename string) {
	buffer.key = filename
}

func (buffer *Buffer) IncRef() int32 {
	return atomic.AddInt32(&buffer.ref, 1)
}

// INFO: 读取 S3 文件，读取到 data 中
func (buffer *Buffer) ReadFile(offset int64, data []byte, max int64, direct bool) (int, error) {
	var err error
	var nRead int

	defer func() {
		if err == io.EOF {
			err = nil
		}
	}()

	// read data from backend if direct flag is set
	if direct {
		nRead, err = buffer.readDirect(offset, data)
		if err != nil && err != io.EOF {
			return 0, err
		}
	}

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

	return nRead, nil
}

func (buffer *Buffer) readDirect(offset int64, data []byte) (int, error) {
	rsize, err := buffer.backend.Read(buffer.key, offset, data)
	if err != nil && err != io.EOF {
		return 0, err
	}

	return rsize, nil
}

func NewBuffer(inodeID uint64, backend backend.Backend) *Buffer {
	buffer := &Buffer{
		inodeID: inodeID,
		//blockSize:     fs.blockSize,
		//flushInterval: fs.bufFlushInterval,
		//blocks:        make(map[int64]*list.Element, 0),
		LRUList: list.New(),
		//dirtyBlocks:   make(map[int64]*list.Element, 0),
		//dirtyList:     list.New(),
		//fs:            fs,
		//gbuf:          gbuf,
		backend: backend,
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

	return buffer
}