package bbolt

import (
	"fmt"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
	"k8s.io/klog/v2"
)

var testBucket = []byte("test-bucket")

func TestBasic(test *testing.T) {
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		klog.Fatal(err)
	}
	// 参数true表示创建一个写事务，false读事务
	tx, err := db.Begin(true)
	if err != nil {
		klog.Fatal(err)
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte("key"))
	if err != nil {
		klog.Fatal(err)
	}
	// 使用bucket对象更新一个key
	if err = bucket.Put([]byte("hello"), []byte("world")); err != nil {
		klog.Fatal(err)
	}
	// 提交事务
	if err := tx.Commit(); err != nil {
		klog.Fatal(err)
	}
	db.Close()

	stopChan := make(chan struct{})
	readView := func() {
		db2, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			klog.Fatal(err)
		}
		defer db2.Close()
		tx2, err := db2.Begin(false)
		if err != nil {
			klog.Fatal(err)
		}
		defer tx2.Rollback()
		bucket2 := tx2.Bucket([]byte("key"))
		if bucket2 == nil {
			klog.Fatal(fmt.Sprintf("bucket (key) is not existed in db"))
		}
		value := bucket2.Get([]byte("hello"))
		klog.Infof(fmt.Sprintf("value=%s", string(value)))
	}

	tick := time.Tick(time.Second * 3)
	for {
		select {
		case <-tick:
			readView()
		case <-stopChan:
			return
		}
	}
}

// https://zhengyinyong.com/post/bbolt-first-experience/
func TestBbolt(test *testing.T) {
	// 在当前目录下打开 my.db 这个文件
	// 如果文件不存在，将会自动创建
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		klog.Fatal(err)
	}
	defer db.Close()

	key := []byte("hello")
	value := []byte("world")

	// 创建一个 read-write transaction 来进行写操作
	err = db.Update(func(tx *bolt.Tx) error {
		// !!! 这个lambda表达式内commit的数据不会落入disk

		// 如果 bucket 不存在则，创建一个 bucket
		bucket, err := tx.CreateBucketIfNotExists(testBucket)
		if err != nil {
			return err
		}

		// 将 key-value 写入到 bucket 中
		return bucket.Put(key, value)
	})
	if err != nil {
		klog.Fatal(err)
	}

	// 创建一个 read-only transaction 来获取数据
	err = db.View(func(tx *bolt.Tx) error {
		// 获取对应的 bucket
		bucket := tx.Bucket(testBucket)
		// 如果 bucket 返回为 nil，则说明不存在对应 bucket
		if bucket == nil {
			return fmt.Errorf("bucket %q is not found", testBucket)
		}
		// 从 bucket 中获取对应的 key（即上面写入的 key-value）
		value := bucket.Get(key)
		fmt.Printf("%s: %s\n", string(key), string(value))

		// read-only txn 内不能写操作
		/*err = bucket.Put(key, []byte("world2"))
		if err != nil { // error: "tx not writable"
			return err
		}*/

		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			fmt.Printf("key=%s, value=%s\n", key, value)
		}

		return nil
	})
	if err != nil {
		klog.Fatal(err)
	}
}
