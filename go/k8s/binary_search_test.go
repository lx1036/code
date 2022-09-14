package k8s

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// https://leetcode-cn.com/problems/binary-search/ 704

// 二分查找, 查找范围是 [left,right]，left<=right

// 题目：对于一个有序的 []int，查找 target。时间复杂度是 logN

func search(nums []int, target int) int {
	left, right := 0, len(nums)-1

	for left <= right {
		mid := (right-left)/2 + left
		if nums[mid] == target {
			return mid
		} else if nums[mid] > target {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}

	return -1
}

func TestSearch(test *testing.T) {
	nums := []int{-1, 0, 3, 5, 9, 12}
	target := 9
	ans := search(nums, target)
	assert.Equal(test, 4, ans)

	nums = []int{-1, 0, 3, 5, 9, 12}
	target = 2
	ans = search(nums, target)
	assert.Equal(test, -1, ans)
}

// 协程池
type ThreadPool struct {
	size   int
	worker chan func()
}

func NewThreadPool(size int) *ThreadPool {
	return &ThreadPool{
		size:   size,
		worker: make(chan func(), 100),
	}
}

func (p *ThreadPool) Add(task func()) {
	select {
	case p.worker <- task:
	default:
	}
}

func (p *ThreadPool) Run() {
	var wg sync.WaitGroup
	for i := 0; i < p.size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case t := <-p.worker:
					t()
				}
			}
		}()
	}

	wg.Wait()
}

func TestThreadPool(test *testing.T) {
	stopCh := make(chan struct{})

	p := NewThreadPool(10)
	go p.Run()

	j := 0
	for i := 0; i < 5; i++ {
		j++
		p.Add(func() {
			fmt.Println(j)
		})
	}

	time.Sleep(time.Second)

	p.Add(func() {
		fmt.Println(6)
	})

	<-stopCh
}

type Pool struct {
	taskCh chan func()
}

func NewPool(size int) *Pool {
	p := &Pool{
		taskCh: make(chan func(), 100),
	}
	
	for i := 0; i < size; i++ {
		go func() {
			for task := range p.taskCh {
				task()
			}
		}()
	}
	
	return p
}

func (p *Pool) Add(task func())  {
	p.taskCh <- task
}