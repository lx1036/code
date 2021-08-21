package fs

import (
	"io"
	"sync"
	"sync/atomic"

	"k8s-lx1036/k8s/storage/fusefs/pkg/backend"
)

// INFO: buffer 就是封装了下 s3 api 调用
type Buffer struct {
	sync.RWMutex

	key string // filename

	inodeID uint64
	ref     int32

	backend backend.Backend

	lastError error
}

func (buffer *Buffer) SetFilename(filename string) {
	buffer.key = filename
}

func (buffer *Buffer) IncRef() int32 {
	return atomic.AddInt32(&buffer.ref, 1)
}

// INFO: 读取 S3 文件，读取到 data 中
func (buffer *Buffer) ReadFile(offset int64, data []byte, max int64, direct bool) (nRead int,
	err error) {

	nNeed := len(data)
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

	}

	return
}

func (buffer *Buffer) readDirect(offset int64, data []byte) (int, error) {
	rsize, err := buffer.backend.Read(buffer.key, offset, data)
	if err != nil && err != io.EOF {
		return 0, err
	}

	return rsize, nil
}

func NewBuffer() *Buffer {

}
