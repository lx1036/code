package apimachinery

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"testing"
	"time"
)

type Task struct {
	stop chan struct{}
	period time.Duration
}
func (task *Task) process() {
	fmt.Println("a")
}

func TestWaitUtil(test *testing.T) {
	task := &Task{
		stop: make(chan struct{}),
		period: time.Second * 2,
	}
	
	go wait.Until(task.process, task.period, task.stop)
	
	<-task.stop
}
