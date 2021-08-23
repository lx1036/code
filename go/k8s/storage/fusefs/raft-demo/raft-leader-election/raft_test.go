package main

import (
	"k8s.io/klog"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func call() int {
	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)

	return rand.Intn(10)
}

func TestLock(test *testing.T) {
	mu := sync.Mutex{}
	votesReceived := 1
	peerIds := []int{1, 2, 3}

	// INFO: 这里由于有 lock 存在，所以 votesReceived 是依次累加的，比如
	//  reply=7,votesReceived=8;reply=1;votesReceived=8+1=9;reply=8,votesReceived=9+8=17
	for _, peerId := range peerIds {
		go func(peerId int) {
			reply := call()
			klog.Infof("reply: %d", reply)
			mu.Lock()
			defer mu.Unlock()

			votesReceived += reply
			if votesReceived > 5 {
				klog.Infof("votesReceived: %d, is a leader", votesReceived)
				return
			}
			klog.Infof("votesReceived: %d, is a candidate", votesReceived)
		}(peerId)
	}

	time.Sleep(20 * time.Second)
}
