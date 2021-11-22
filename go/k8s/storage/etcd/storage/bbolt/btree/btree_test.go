package btree

import (
	"fmt"
	"github.com/google/btree"
	"testing"
)

type foo struct {
	value int64
}

// btree 存放的东西必须实现 Less()，即 Item 接口
func (i *foo) Less(b btree.Item) bool {
	return i.value < b.(*foo).value
}

// btree 的使用: https://zhengyinyong.com/post/btree-usage/
func TestReplaceOrInsert(test *testing.T) {
	// 创建一颗 btree
	tree := btree.New(32)

	// 创建 3 个测试数据
	f1 := foo{value: 123}
	f2 := foo{value: 456}
	f3 := foo{value: 789}

	// 插入到 btree 中
	tree.ReplaceOrInsert(&f1)
	tree.ReplaceOrInsert(&f2)
	tree.ReplaceOrInsert(&f3)

	// 按照升序打印 >= 0 的数据
	// 此时应该一次打印 123,456,789
	tree.AscendGreaterOrEqual(&foo{value: 0}, func(item btree.Item) bool {
		f := item.(*foo)
		fmt.Println(f.value)
		return true
	})

	// 此时应该什么都打印不出来
	tree.AscendGreaterOrEqual(&foo{value: 999}, func(item btree.Item) bool {
		f := item.(*foo)
		fmt.Println(f.value)
		return true
	})
}
