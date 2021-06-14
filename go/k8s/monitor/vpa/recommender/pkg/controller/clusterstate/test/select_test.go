package test

import (
	"k8s.io/klog/v2"
	"testing"
)

func TestSelect(test *testing.T) {
	stopCh := make(chan struct{})

Loop:
	for {
		select {
		case oom := <-stopCh:
			klog.Infof("oom %+v", oom)
		default:
			klog.Info("default")
			break Loop
		}
	}

	klog.Info("success")
	// output: default success
}
