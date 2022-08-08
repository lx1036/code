package queuesort

import (
	"testing"
	"time"

	framework "k8s-lx1036/k8s/scheduler/pkg/framework"

	v1 "k8s.io/api/core/v1"
)

func TestLess(test *testing.T) {
	prioritySort := &PrioritySort{}
	var lowPriority, highPriority = int32(10), int32(100)
	t1 := time.Now()
	t2 := t1.Add(time.Second)
	for _, tt := range []struct {
		name     string
		p1       *framework.QueuedPodInfo
		p2       *framework.QueuedPodInfo
		expected bool
	}{
		{
			name: "p1.priority less than p2.priority",
			p1: &framework.QueuedPodInfo{
				Pod: &v1.Pod{
					Spec: v1.PodSpec{
						Priority: &lowPriority,
					},
				},
			},
			p2: &framework.QueuedPodInfo{
				Pod: &v1.Pod{
					Spec: v1.PodSpec{
						Priority: &highPriority,
					},
				},
			},
			expected: false, // p2 should be ahead of p1 in the queue
		},
		{
			name: "p1.priority greater than p2.priority",
			p1: &framework.QueuedPodInfo{
				Pod: &v1.Pod{
					Spec: v1.PodSpec{
						Priority: &highPriority,
					},
				},
			},
			p2: &framework.QueuedPodInfo{
				Pod: &v1.Pod{
					Spec: v1.PodSpec{
						Priority: &lowPriority,
					},
				},
			},
			expected: true, // p1 should be ahead of p2 in the queue
		},
		{
			name: "equal priority. p1 is added to schedulingQ earlier than p2",
			p1: &framework.QueuedPodInfo{
				Pod: &v1.Pod{
					Spec: v1.PodSpec{
						Priority: &highPriority,
					},
				},
				Timestamp: t1,
			},
			p2: &framework.QueuedPodInfo{
				Pod: &v1.Pod{
					Spec: v1.PodSpec{
						Priority: &highPriority,
					},
				},
				Timestamp: t2,
			},
			expected: true, // p1 should be ahead of p2 in the queue
		},
		{
			name: "equal priority. p2 is added to schedulingQ earlier than p1",
			p1: &framework.QueuedPodInfo{
				Pod: &v1.Pod{
					Spec: v1.PodSpec{
						Priority: &highPriority,
					},
				},
				Timestamp: t2,
			},
			p2: &framework.QueuedPodInfo{
				Pod: &v1.Pod{
					Spec: v1.PodSpec{
						Priority: &highPriority,
					},
				},
				Timestamp: t1,
			},
			expected: false, // p2 should be ahead of p1 in the queue
		},
	} {
		test.Run(tt.name, func(t *testing.T) {
			if got := prioritySort.Less(tt.p1, tt.p2); got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
