package main

import (
	"fmt"
	"testing"
)

func TestSum(t *testing.T) {
	array := [5]int{1, 2, 3, 4, 5}
	slice := []string{"a", "b", "c"}
	maps := map[string]int{"a": 1, "b": 2}

	fmt.Println(array, slice, maps)
}
