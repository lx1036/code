package wal

import (
	"io"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"testing"
)

func TestFileRead(test *testing.T) {
	file, err := os.OpenFile("META", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return
	}

	meta := make([]byte, 30)
	n, err := file.Read(meta)
	if err != nil && err != io.EOF {
		klog.Error(err)
	}
	klog.Info(n, strings.TrimSpace(string(meta)))
}
