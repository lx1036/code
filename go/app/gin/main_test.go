package main

import (
	"fmt"
	"testing"
)

/**
???: Gin 的 call http test helper 方法，类似 laravel 的 app->call() 直接进入 application 内
*/
func TestName(test *testing.T) {
	paths := map[string]struct{}{
		"json": {},
	}
	value, bools := paths["json"]
	fmt.Println(value, bools)
}
