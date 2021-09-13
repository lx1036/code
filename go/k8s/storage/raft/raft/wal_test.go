// INFO: https://github.com/tidwall/wal/blob/master/README.md

package raft

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/wal"
	"k8s.io/klog/v2"
)

// rm -rf ./log
func TestWalWriteRead(test *testing.T) {
	// INFO: basic write/read
	log, err := wal.Open("log", &wal.Options{
		//NoSync:           false,
		//SegmentSize:      0,
		LogFormat: wal.JSON, // 这里为了方便debug
		//SegmentCacheSize: 0,
		//NoCopy:           false,
	})
	if err != nil {
		klog.Fatal(err)
	}
	defer log.Close()

	type Person struct {
		Age  int    `json:"age"`
		Name string `json:"name"`
	}
	person1 := Person{
		Age:  18,
		Name: "lx1036",
	}
	data1, _ := json.Marshal(person1)
	err = log.Write(1, data1)
	person2 := Person{
		Age:  19,
		Name: "lx1037",
	}
	data2, _ := json.Marshal(person2)
	err = log.Write(2, data2)
	if err != nil {
		klog.Fatal(err)
	}

	data, err := log.Read(1)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Info(string(data))

	// INFO: batch write
	// write three entries as a batch
	batch := new(wal.Batch)
	//batch.Write(1, []byte("first entry")) // out of order
	batch.Write(3, []byte("second entry"))
	batch.Write(4, []byte("third entry"))
	err = log.WriteBatch(batch)
	if err != nil {
		klog.Fatal(err)
	}
	data, err = log.Read(4)
	if err != nil {
		klog.Fatal(err)
	}
	klog.Info(string(data))

	// 把index=2之前的log，和index=3之后的log都给truncate掉
	// 其实就是保留index=2~3的log
	err = log.TruncateFront(2)
	err = log.TruncateBack(3)
	data, err = log.Read(2)
	if err != nil {
		klog.Error(err)
	} else {
		klog.Info(string(data))
	}
}
