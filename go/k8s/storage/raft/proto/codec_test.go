package proto

import (
	"testing"

	"k8s.io/klog/v2"
)

func TestDecodeHBContext(test *testing.T) {
	buf := "hello world"
	heartbeatContext := DecodeHBContext([]byte(buf))
	for _, value := range heartbeatContext {
		klog.Info(value)
	}
}
