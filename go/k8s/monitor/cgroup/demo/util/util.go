package util

/*
#include "util.c"
*/
import "C"

import "fmt"

func GoSum(a, b int) {
	s := C.sumlx(C.int(a), C.int(b))
	fmt.Println(s)
}
