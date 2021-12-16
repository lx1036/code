package raftstore

import (
	"k8s.io/klog/v2"
	"testing"
	"time"
)

func TestRaftStore(test *testing.T) {
	type Msg struct {
		msg string
	}
	stopCh := make(chan struct{})
	msgCh := make(chan Msg, 5)
	var msgs []Msg
	go func() {
		readyCh := make(chan struct{}, 1) // 10
		for {
			if len(msgs) > 0 {
				readyCh <- struct{}{}
				klog.Info("readyCh")
			}
			select {
			case <-stopCh:
				return
			case msg := <-msgCh:
				msgs = append(msgs, msg)
				klog.Info("msgCh")
			case <-readyCh:
				if len(msgs) > 0 {
					klog.Info(len(msgs))
				}
				klog.Info("adfadf")
				msgs = msgs[:0]
			}
		}
	}()

	time.Sleep(time.Second)

	go func() {
		msgCh <- Msg{msg: "123"}
		klog.Info("123")
	}()

	go func() {
		msgCh <- Msg{msg: "456"}
		klog.Info("456")
	}()

	<-stopCh
}
