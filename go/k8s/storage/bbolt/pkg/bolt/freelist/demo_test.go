package freelist

import (
	"sort"
	"testing"

	"k8s.io/klog/v2"
)

func TestSearch(test *testing.T) {
	lead := []int{2, 5, 8}
	follow := []int{3, 4, 6}
	dst := make([]int, len(lead)+len(follow))
	merged := dst[:0]
	klog.Info(merged)

	for len(lead) > 0 {
		n := sort.Search(len(lead), func(i int) bool {
			return lead[i] > follow[0]
		})
		klog.Info(n)
		merged = append(merged, lead[:n]...)
		klog.Info(merged)
		if n >= len(lead) {
			break
		}
		// Swap lead and follow.
		lead, follow = follow, lead[n:]
		klog.Info(lead, follow)
	}

	klog.Info(merged)
	// Append what's left in follow.
	merged = append(merged, follow...)
	klog.Info(merged, dst)
}
