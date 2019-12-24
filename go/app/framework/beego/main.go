package main

import (
	"fmt"
	"github.com/astaxie/beego"
	_ "k8s-lx1036/app/framework/beegowork/beego/routers"
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

	beego.SetStaticPath("/md", "assets/md")
	// 1. resolve /conf
	// 2. user-defined hook func
	// 3. start session
	// 4. compile template
	// 5. monitor serving in 8088
	// 6. listen and serve 8081
	beego.Run("localhost:8081")
}
