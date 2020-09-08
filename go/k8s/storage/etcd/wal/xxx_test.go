package wal

import (
	"fmt"
	"go.etcd.io/etcd/wal"
	"go.etcd.io/etcd/wal/walpb"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewXXX(test *testing.T) {
	p, err := ioutil.TempDir(os.TempDir(), "waltest")
	if err != nil {
		test.Fatal(err)
	}
	defer os.RemoveAll(p)

	w, err := wal.Create(zap.NewExample(), p, []byte("somedata"))
	if err != nil {
		test.Fatalf("err = %v, want nil", err)
	}
	defer w.Close()
	t := &walpb.Record{Type: 0, Crc: 0}
	fmt.Println(t)
}
