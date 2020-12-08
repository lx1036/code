package channel

import (
	"fmt"
	"runtime"
	"testing"
)

func TestProcess(test *testing.T) {
	fmt.Println(runtime.GOMAXPROCS(runtime.NumCPU()))
}






