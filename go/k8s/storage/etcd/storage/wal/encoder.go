package wal

import (
	"encoding/binary"
	"hash"
	"io"
	"os"
	"sync"

	"go.etcd.io/etcd/pkg/v3/crc"
	"go.etcd.io/etcd/pkg/v3/ioutil"
	"go.etcd.io/etcd/server/v3/wal/walpb"
)

// walPageBytes is the alignment for flushing records to the backing Writer.
// It should be a multiple of the minimum sector size so that WAL can safely
// distinguish between torn writes and ordinary data corruption.
const walPageBytes = 8 * minSectorSize

type encoder struct {
	mu sync.Mutex
	bw *ioutil.PageWriter

	crc       hash.Hash32
	buf       []byte
	uint64buf []byte
}

//
func (e *encoder) encode(rec *walpb.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.crc.Write(rec.Data)
	rec.Crc = e.crc.Sum32()
	var (
		data []byte
		err  error
		n    int
	)

	if rec.Size() > len(e.buf) {
		data, err = rec.Marshal()
		if err != nil {
			return err
		}
	} else {
		n, err = rec.MarshalTo(e.buf)
		if err != nil {
			return err
		}
		data = e.buf[:n]
	}

	lenField, padBytes := encodeFrameSize(len(data))
	if err = writeUint64(e.bw, lenField, e.uint64buf); err != nil {
		return err
	}

	if padBytes != 0 {
		data = append(data, make([]byte, padBytes)...)
	}
	n, err = e.bw.Write(data)
	walWriteBytes.Add(float64(n))
	return err
}

func (e *encoder) flush() error {
	e.mu.Lock()
	n, err := e.bw.FlushN()
	e.mu.Unlock()
	walWriteBytes.Add(float64(n))
	return err
}

func writeUint64(w io.Writer, n uint64, buf []byte) error {
	// http://golang.org/src/encoding/binary/binary.go
	binary.LittleEndian.PutUint64(buf, n)
	nv, err := w.Write(buf)
	walWriteBytes.Add(float64(nv))
	return err
}

func encodeFrameSize(dataBytes int) (lenField uint64, padBytes int) {
	lenField = uint64(dataBytes)
	// force 8 byte alignment so length never gets a torn write
	padBytes = (8 - (dataBytes % 8)) % 8
	if padBytes != 0 {
		lenField |= uint64(0x80|padBytes) << 56
	}
	return lenField, padBytes
}

func newEncoder(w io.Writer, prevCrc uint32, pageOffset int) *encoder {
	return &encoder{
		bw:  ioutil.NewPageWriter(w, walPageBytes, pageOffset),
		crc: crc.New(prevCrc, crcTable),
		// 1MB buffer
		buf:       make([]byte, 1024*1024),
		uint64buf: make([]byte, 8),
	}
}

// newFileEncoder creates a new encoder with current file offset for the page writer.
func newFileEncoder(f *os.File, prevCrc uint32) (*encoder, error) {
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	return newEncoder(f, prevCrc, int(offset)), nil
}
