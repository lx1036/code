package heap

import (
	"k8s.io/klog/v2"
	"math"
	"testing"
)

// https://leetcode.cn/problems/largest-number-after-digit-swaps-by-parity/solution/xiang-tong-de-by-lx1036-u4kn/
// 这里不需要使用优先级队列

func largestInteger2(num int) int {
	var value []int
	for num > 0 {
		value = append(value, num%10)
		num = num / 10
	}

	l := len(value) - 1
	for i := 0; i <= l; i++ {
		for j := i + 1; j <= l; j++ {
			if (value[i]-value[j])%2 == 0 && (value[i] > value[j]) { // 具有相同的奇偶性，且 i>j，则 swap
				value[j], value[i] = value[i], value[j]
			}
		}
	}

	result := 0
	for index, data := range value {
		result += int(math.Pow(10, float64(index)) * float64(data))
	}

	return result
}

func TestLargestInteger(test *testing.T) {
	klog.Info(largestInteger2(1234))
	klog.Info(largestInteger2(65875))
	klog.Info(largestInteger2(247))
}
