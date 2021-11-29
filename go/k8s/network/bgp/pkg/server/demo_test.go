package server

import (
	"k8s.io/klog/v2"
	"testing"
	"time"
)

func TestTimer(test *testing.T) {
	holdTimer := &time.Timer{}
	idleHoldTimer := time.NewTimer(time.Second * time.Duration(0)) // INFO: 起始为0，<-idleHoldTimer.C 会先走

	for {
		select {
		case <-holdTimer.C:
			klog.Infof("abcd")
			return
		case <-idleHoldTimer.C: // 走这个逻辑
			klog.Infof("1234")
			return
		}
	}
}
