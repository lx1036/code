package bbolt

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"testing"
	"time"
)

var testBucket = []byte("test-bucket")

// https://zhengyinyong.com/post/bbolt-first-experience/
func TestBbolt(test *testing.T) {
	// 在当前目录下打开 my.db 这个文件
	// 如果文件不存在，将会自动创建
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
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
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
