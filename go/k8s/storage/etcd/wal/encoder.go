package wal

import (
	"go.etcd.io/etcd/pkg/crc"
	"go.etcd.io/etcd/pkg/ioutil"
	"hash"
	"sync"
	"io"
)

type encoder struct {
	mu sync.Mutex
	bw *ioutil.PageWriter
	
	crc       hash.Hash32
	buf       []byte
	uint64buf []byte
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
