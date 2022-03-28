package master

import (
	"fmt"
	"k8s.io/klog/v2"
	"testing"
)

func TestMetaPartitionCount(test *testing.T) {
	klog.Infof(fmt.Sprintf("%d", defaultMetaPartitionInodeIDStep))
}
