package main

import "fmt"

func main() {
	p0 := new(int)
	fmt.Println(p0)
	fmt.Println(*p0)

	x := *p0
	p1, p2 := &x, &x
	fmt.Println(p1 == p2) // true
	fmt.Println(p1 == p0) // false

	p3 := &*p0
	fmt.Println(p3 == p0) // true
	*p0, *p1 = 123, 789
	fmt.Println(*p2, x, *p3) //  789, 789,123

	fmt.Printf("%T %T\n", *p0, x) // 123, 789
	fmt.Printf("%T %T\n", p0, p1) // 0x, 0x
}
