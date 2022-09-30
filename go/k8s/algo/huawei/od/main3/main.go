package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// https://leetcode.cn/circle/discuss/uZdKzf/ : 篮球比赛分组问题

func main() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := strings.Split(input.Text(), " ")
	var group []int
	for i := 0; i < len(data); i++ {
		tmp, _ := strconv.Atoi(data[i])
		group = append(group, tmp)
	}
	res := cal(group)
	fmt.Printf("%d", res)
}

func cal(group []int) int {
	sort.Ints(group)
	ans := 1<<32 - 1
	sum := 0
	for i := 0; i < len(group); i++ {
		sum += group[i]
	}

	for j := 0; j < 7; j++ {
		for k := j + 1; k < 8; k++ {
			begin := k + 1
			end := 9
			for begin < end {
				bal := group[0] + group[j] + group[k] + group[begin] + group[end]
				if sum%2 == 0 && sum-bal*2 == 0 {
					return 0
				} else if sum%2 == 1 && abs(sum-bal*2) == 1 {
					return 1
				}
				if abs(sum-bal*2) < ans {
					ans = abs(sum - bal*2)
				}
				if (sum - 2*bal) > 0 {
					begin++
				} else {
					end--
				}
			}
		}
	}
	return ans
}
func abs(a int) int {
	if a < 0 {
		return a * -1
	}
	return a
}
