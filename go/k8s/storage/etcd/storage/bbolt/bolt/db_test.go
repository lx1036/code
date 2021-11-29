package bolt

import (
	"fmt"
	"k8s.io/klog/v2"
	"testing"
)

func TestDBOpen(test *testing.T) {
	db, err := Open("./tmp/my.db", 0666, nil)
	if err != nil {
		klog.Fatal(err)
	}
	defer db.close()

	klog.Info(db.filesz / 1024) // 16
	klog.Infof(fmt.Sprintf("%+v", db.meta0))
	klog.Infof(fmt.Sprintf("%+v", db.meta1))

	tx := &Tx{writable: true}
	tx.init(db)
	foo, err := tx.CreateBucket([]byte("foo"))
	if err != nil {
		klog.Fatal(err)
	}
	klog.Info(fmt.Sprintf("%+v", *foo))
}
