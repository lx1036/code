package wal

import (
	"bytes"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"k8s-lx1036/k8s/storage/etcd/wal/walpb"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(test *testing.T) {
	p, err := ioutil.TempDir(os.TempDir(), "waltest")
	if err != nil {
		test.Fatal(err)
	}
	defer os.RemoveAll(p)

	w, err := Create(zap.NewExample(), p, []byte("somedata"))
	if err != nil {
		test.Fatalf("err = %v, want nil", err)
	}
	if g := filepath.Base(w.tail().Name()); g != walName(0, 0) {
		test.Errorf("name = %+v, want %+v", g, walName(0, 0))
	}
	defer w.Close()

	// file is preallocated to segment size; only read data written by wal
	off, err := w.tail().Seek(0, io.SeekCurrent)
	if err != nil {
		test.Fatal(err)
	}
	gd := make([]byte, off)
	f, err := os.Open(filepath.Join(p, filepath.Base(w.tail().Name())))
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()
	if _, err = io.ReadFull(f, gd); err != nil {
		test.Fatalf("err = %v, want nil", err)
	}

	var wb bytes.Buffer
	e := newEncoder(&wb, 0, 0)
	err = e.encode(&walpb.Record{Type: crcType, Crc: 0})
	if err != nil {
		test.Fatalf("err = %v, want nil", err)
	}
	err = e.encode(&walpb.Record{Type: metadataType, Data: []byte("somedata")})
	if err != nil {
		test.Fatalf("err = %v, want nil", err)
	}
	r := &walpb.Record{
		Type: snapshotType,
		Data: pbutil.MustMarshal(&walpb.Snapshot{}),
	}
	if err = e.encode(r); err != nil {
		test.Fatalf("err = %v, want nil", err)
	}
	e.flush()
	if !bytes.Equal(gd, wb.Bytes()) {
		test.Errorf("data = %v, want %v", gd, wb.Bytes())
	}
}
