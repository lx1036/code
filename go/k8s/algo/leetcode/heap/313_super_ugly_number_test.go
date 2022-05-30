package heap

import (
	"testing"

	"k8s.io/klog/v2"
)

// https://leetcode.cn/problems/super-ugly-number/

func nthSuperUglyNumber(n int, primes []int) int {
	pq := NewPriorityQueue264()
	pq.Push(&Item264{value: 1})

	var result int
	results := map[int]bool{1: true}
	for i := 1; ; i++ {
		item := pq.Pop()
		if i == n {
			result = item.value
			break
		}

		for _, factor := range primes {
			data := factor * item.value
			if _, ok := results[data]; !ok { // 去重再push
				results[data] = true
				pq.Push(&Item264{value: data})
			}
		}
	}

	return result
}

func TestNthUglyNumber313(test *testing.T) {
	primes := []int{2, 7, 13, 19}
	n := 12
	klog.Info(nthSuperUglyNumber(n, primes))

	primes = []int{2, 3, 5}
	n = 1
	klog.Info(nthSuperUglyNumber(n, primes))
}
