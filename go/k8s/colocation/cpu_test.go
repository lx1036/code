package colocation

import (
	"math"
	"testing"

	"k8s.io/klog/v2"
)

func TestCPU(test *testing.T) {
	onlineOnly := 0.001575515
	onlineWithOffline := 0.001574852
	klog.Info(float64(math.Abs((onlineWithOffline - onlineOnly) / onlineOnly))) // 干扰率是 0.9%，在 5% 以内
}
