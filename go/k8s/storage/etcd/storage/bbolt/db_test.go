package bbolt

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"

	bolt "go.etcd.io/bbolt"

	"k8s.io/klog/v2"
)

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}

// Ensure that a database can be opened without error.
func TestOpen(t *testing.T) {
	path := tempfile()
	defer os.RemoveAll(path)

	klog.Infof("path %s", path)
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		t.Fatal(err)
	} else if db == nil {
		t.Fatal("expected db")
	}

	if s := db.Path(); s != path {
		t.Fatalf("unexpected path: %s", s)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

// Regression validation for https://github.com/etcd-io/bbolt/pull/122.
// Tests multiple goroutines simultaneously opening a database.
func TestOpen_MultipleGoroutines(t *testing.T) {
	const (
		instances  = 3
		iterations = 3
	)
	path := tempfile()
	defer os.RemoveAll(path)
	var wg sync.WaitGroup
	errCh := make(chan error, iterations*instances) // buffer channel 3*3
	for iteration := 0; iteration < iterations; iteration++ {
		for instance := 0; instance < instances; instance++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db, err := bolt.Open(path, 0600, nil)
				if err != nil {
					errCh <- err
					return
				}
				if err := db.Close(); err != nil {
					errCh <- err
					return
				}
			}()
		}
		wg.Wait()
	}
	// 这里为啥关闭channel，关闭channel还能range么？
	// close channel 是说该channel不在接收新数据，该channel里的已有数据还可以继续range
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("error from inside goroutine: %v", err)
		}
	}
}

func TestName(test *testing.T) {
	const (
		instances  = 30
		iterations = 30
	)
	errCh := make(chan error, iterations*instances)
	//errCh <- fmt.Errorf("test error")
	//defer close(errCh) // defer会hang住
	close(errCh)
	//errCh = nil // range nil channel 不会报错，会一直block
	//errCh <- fmt.Errorf("test error") // 给 nil channel 发数据，会block
	for err := range errCh {
		if err != nil {
			test.Fatalf("error from inside goroutine: %v", err)
		}
	}
}
