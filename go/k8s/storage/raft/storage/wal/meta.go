package wal

import (
	"encoding/binary"
	"fmt"
	"github.com/tiglabs/raft/util/bufalloc"
	"io"
	"os"
	"path"
	
	"github.com/tiglabs/raft/proto"

)

type truncateMeta struct {
	truncateIndex uint64
	truncateTerm uint64
}

func (meta *truncateMeta) Size() uint64 {
	return 16
}

func (meta *truncateMeta) Decode(buf []byte)  {
	meta.truncateIndex = binary.BigEndian.Uint64(buf)
	meta.truncateTerm = binary.BigEndian.Uint64(buf[8:])
}

// META文件的对象
type metaFile struct {
	file *os.File // META file
	
	truncateOffset int64
}

// 加载META文件内容
func (mf *metaFile) load() (hardState proto.HardState, meta truncateMeta, err error) {
	hardStateSize := int(hardState.Size())
	buffer := bufalloc.AllocBuffer(hardStateSize)
	defer bufalloc.FreeBuffer(buffer)
	
	// 读取META文件内容
	buf := buffer.Alloc(hardStateSize)
	n , err := mf.file.Read(buf)
	if err != nil {
		if err == io.EOF {
			err = nil
			return
		}
		return
	}
	if n != hardStateSize {
		err = fmt.Errorf("wrong hardstate data size from META file")
	}
	hardState.Decode(buf)
	
	buffer.Reset()
	metaSize := int(meta.Size())
	buf = buffer.Alloc(metaSize)
	n , err = mf.file.Read(buf)
	if err != nil {
		if err == io.EOF {
			err = nil
			return
		}
		return
	}
	if n != metaSize {
		err = fmt.Errorf("wrong truncate meta from META file")
	}
	meta.Decode(buf)
	
	return
}

// dir/META 文件
func openMetaFile(dir string) (mf *metaFile, hardState proto.HardState, meta truncateMeta, err error) {
	file, err := os.OpenFile(path.Join(dir, "META"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return
	}
	
	mf = &metaFile{
		file: file,
		truncateOffset: int64(hardState.Size()), // INFO: 注意这里的 hardState 使用
	}
	
	hardState, meta, err = mf.load()
	return
}

