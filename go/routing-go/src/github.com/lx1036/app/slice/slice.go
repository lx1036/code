package main

import "fmt"

func main()  {
	//s1 := []int{0, 1, 2, 3, 8:100}
	s2 := make([]int, 5, 10)
	s2[2] = 3

	fmt.Println(s2)
	//fmt.Println(s1, len(s1), cap(s1), s2)
}
