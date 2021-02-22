package pkg

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

// @see staging/src/k8s.io/apimachinery/pkg/util/wait/wait_test.go

func TestTestJitterUntil(test *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	wait.JitterUntil(func() {
		klog.Info("TestTestJitterUntil ...")

		cancel() // 定时任务被cancel了，只会执行一次
	}, time.Second, 1.2, true, ctx.Done())
}
