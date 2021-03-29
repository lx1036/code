package util

import (
	"fmt"
	"testing"
)

func TestPool(test *testing.T) {
	stop := make(chan bool, 10)

	select {
	case stop <- true:
		fmt.Println("write")
	case <-stop:
		fmt.Println("stop")
	default:
		fmt.Println("default")
	}
}
