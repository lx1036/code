package reflector

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/clock"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestName(test *testing.T) {
	stopCh := make(chan struct{}, 1)
	Run(stopCh)
}

func Run(stopCh <-chan struct{}) {
	realClock := &clock.RealClock{}
	backoffManager := wait.NewExponentialBackoffManager(800*time.Millisecond, 30*time.Second, 2*time.Minute, 2.0, 1.0, realClock)

	wait.BackoffUntil(func() {
		if err := ListAndWatch(stopCh); err != nil {
			utilruntime.HandleError(err)
		}
	}, backoffManager, true, stopCh)
}

func ListAndWatch(stopCh <-chan struct{}) error {
	listCh := make(chan struct{}, 1)
	var err error

	go func() {
		// do something
		fmt.Println("test")
		time.Sleep(time.Second)

		close(listCh)
	}()

	select {
	case <-listCh:
	}

	return fmt.Errorf("failed to list %v", err)
}
