package lru

import (
	"fmt"
	"testing"

	tslru "github.com/hashicorp/golang-lru"
	"github.com/hashicorp/golang-lru/simplelru"
)

// https://github.com/hashicorp/golang-lru/blob/master/simplelru/lru_test.go

func TestLRU(test *testing.T) {
	l, _ := tslru.New(128)
	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	if l.Len() != 128 {
		panic(fmt.Sprintf("bad len: %v", l.Len()))
	}
}

func TestLRU2(test *testing.T) {
	l, _ := NewLRU(128)
	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	if l.Len() != 128 {
		panic(fmt.Sprintf("bad len: %v", l.Len()))
	}
}

func TestSimpleLRU2(test *testing.T) {
	l, _ := NewLRU(128)
	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	if l.Len() != 128 {
		panic(fmt.Sprintf("bad len: %v", l.Len()))
	}

	// 这里v==i+128才正确，0-127已经被删除了
	for i, k := range l.Keys() {
		if v, ok := l.Get(k); !ok || v != k || v != i+128 {
			test.Fatalf("bad key: %v", k)
		}
	}
}

func TestSimpleLRU(test *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}) {
		if k != v {
			test.Fatalf("Evict values not equal (%v!=%v)", k, v)
		}

		evictCounter++
	}

	l, err := simplelru.NewLRU(128, onEvicted)
	if err != nil {
		test.Fatalf("err: %v", err)
	}

	for i := 0; i < 256; i++ {
		l.Add(i, i)
	}
	if l.Len() != 128 {
		test.Fatalf("bad len: %v", l.Len())
	}

	if evictCounter != 128 {
		test.Fatalf("bad evict count: %v", evictCounter)
	}

	for i, k := range l.Keys() {
		if v, ok := l.Get(k); !ok || v != k || v != i+128 {
			test.Fatalf("bad key: %v", k)
		}
	}

	for i := 0; i < 128; i++ {
		_, ok := l.Get(i)
		if ok {
			test.Fatalf("should be evicted")
		}
	}
	for i := 128; i < 256; i++ {
		_, ok := l.Get(i)
		if !ok {
			test.Fatalf("should not be evicted")
		}
	}

	for i := 128; i < 192; i++ {
		ok := l.Remove(i)
		if !ok {
			test.Fatalf("should be contained")
		}
		ok = l.Remove(i)
		if ok {
			test.Fatalf("should not be contained")
		}
		_, ok = l.Get(i)
		if ok {
			test.Fatalf("should be deleted")
		}
	}
	l.Get(192) // expect 192 to be last key in l.Keys()

	for i, k := range l.Keys() {
		if (i < 63 && k != i+193) || (i == 63 && k != 192) {
			test.Fatalf("out of order key: %v", k)
		}
	}

	l.Purge()
	if l.Len() != 0 {
		test.Fatalf("bad len: %v", l.Len())
	}
	if _, ok := l.Get(200); ok {
		test.Fatalf("should contain nothing")
	}
}
