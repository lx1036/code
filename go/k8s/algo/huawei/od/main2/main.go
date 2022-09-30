package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	//result := make(map[string]int)
	var result []string

	// 	for i := 0; i < len(data); i++ {
	// 		if !isValid(data[i])  {
	// 			fmt.Print("!error")
	// 			return
	// 		}
	// 		if isNumber(data[i]) && data[i] < '3'  {
	// 			fmt.Print("!error")
	// 			return
	// 		}

	// 		if isNumber(data[i]) {
	// 			l , _ := strconv.Atoi(string(data[i]))
	// 			for j := 0; j < l; j++ {
	// 				result = append(result, string(data[i+1]))
	// 			}
	// 			i++
	// 		} else {
	// 			result = append(result, string(data[i]))
	// 		}
	// 		//result[string(data[i])] = 1
	// 	}

	tmp2 := make(map[byte]int)
	for i := 0; i < len(data); i++ {
		if !isValid(data[i]) {
			fmt.Print("!error")
			return
		}
		if isNumber(data[i]) && data[i] < '3' {
			fmt.Print("!error")
			return
		}
		if i == len(data) && isNumber(data[i]) {
			fmt.Print("!error")
			return
		}
		if !isNumber(data[i]) {
			tmp2[data[i]]++
			if tmp2[data[i]] > 2 {
				fmt.Print("!error")
				return
			}
		}

		if isNumber(data[i]) {
			l, _ := strconv.Atoi(string(data[i]))
			for j := 0; j < l; j++ {
				result = append(result, string(data[i+1]))
			}
			i++
		} else {
			result = append(result, string(data[i]))
		}
	}

	tmp := strings.Join(result, "")
	fmt.Printf("%s", tmp)
}
func isValid(b byte) bool {
	return isNumber(b) || (b < 'z' && b > 'a')
}
func isNumber(b byte) bool {
	return b < '9' && b > '0'
}
