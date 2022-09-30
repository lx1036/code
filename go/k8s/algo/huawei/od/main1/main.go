package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// https://www.peiluming.com/article/95 : 【算法题】构成的正方形数量

func main() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size := input.Text()
	l, _ := strconv.Atoi(size)
	if l < 4 {
		fmt.Printf("%d\n", 0)
		return
	}
	var points [][]int
	for i := 0; i < l; i++ {
		input.Scan()
		data := strings.Split(input.Text(), " ")
		if len(data) != 2 {
			continue
		}
		x, _ := strconv.Atoi(data[0])
		y, _ := strconv.Atoi(data[1])
		points = append(points, []int{x, y})
	}
	result := 0
	for i := 0; i < l-3; i++ {
		for j := i + 1; j < l-2; j++ {
			for k := j + 1; k < l-1; k++ {
				for m := k + 1; m < l; m++ {
					if isSquare(points[i], points[j], points[k], points[m]) {
						result++
					}
				}
			}
		}
	}
	fmt.Printf("%d\n", result)
}
func isSquare(i, j, k, m []int) bool {
	return isZeroVector([]int{i[0] - j[0], i[1] - j[1]}, []int{k[0] - m[0], k[1] - m[1]}) ||
		isZeroVector([]int{i[0] - k[0], i[1] - k[1]}, []int{j[0] - m[0], j[1] - m[1]}) ||
		isZeroVector([]int{i[0] - m[0], i[1] - m[1]}, []int{j[0] - k[0], j[1] - k[1]})
}
func isZeroVector(l1, l2 []int) bool { // 内积为0且长度相等
	return (l1[0]*l2[0]+l1[1]*l2[1]) == 0 && (l1[0]*l1[0]+l1[1]*l1[1] == l2[0]*l2[0]+l2[1]*l2[1])
}
