package main

import (
	"fmt"
)

/*
https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away
*/
func main() {
	array := [5]int{1, 2, 3, 4, 5}
	slice := []string{"a", "b", "c"}
	maps := map[string]int{"a": 1, "b": 2}
	makeSlice := make([]int, 5)
	for i := 0; i < 5; i++ {
		makeSlice[i] = i + 1
	}
	makeSlice = append(makeSlice, 10)

	fmt.Println(array, slice, maps, makeSlice, makeSlice[1:3], makeSlice[1:])

	//beego.Run()
}
