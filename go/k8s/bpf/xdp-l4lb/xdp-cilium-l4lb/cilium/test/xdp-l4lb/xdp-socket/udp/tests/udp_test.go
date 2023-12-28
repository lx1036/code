package tests

import (
	"fmt"
	"testing"
)

func TestUDP(test *testing.T) {
	payload := make([]byte, 10)
	for i := 0; i < len(payload); i++ {
		payload[i] = byte(i)
	}

	fmt.Println(payload) // [0 1 2 3 4 5 6 7 8 9]
}
