package source

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"testing"
	"time"
)

func TestSimpleQueue(t *testing.T) {
	fakeClock := clock.NewFakeClock(time.Now())
	q := workqueue.NewDelayingQueueWithCustomClock(fakeClock, "")

	first := "foo"

	q.AddAfter(first, 50*time.Millisecond)
	if err := waitForWaitingQueueToFill(q); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if q.Len() != 0 {
		t.Errorf("should not have added")
	}

	fakeClock.Step(60 * time.Millisecond)

	if err := waitForAdded(q, 1); err != nil {
		t.Errorf("should have added")
	}
	item, _ := q.Get()
	q.Done(item)

	// step past the next heartbeat
	fakeClock.Step(10 * time.Second)

	err := wait.Poll(1*time.Millisecond, 30*time.Millisecond, func() (done bool, err error) {
		if q.Len() > 0 {
			return false, fmt.Errorf("added to queue")
		}

		return false, nil
	})
	if err != wait.ErrWaitTimeout {
		t.Errorf("expected timeout, got: %v", err)
	}

	if q.Len() != 0 {
		t.Errorf("should not have added")
	}
}
