package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	inputs := strings.Split(input.Text(), " ")

	fmt.Println(inputs)
	last := inputs[len(inputs)-1]

	fmt.Printf("%d\n", len(last))*/

	/*a := 'A' - 'a'
	fmt.Println(a)
	fmt.Println(strings.ToLower(string('A')))*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	input.Scan()
	key := input.Bytes()
	result := 0
	for i := 0; i < len(data); i++ {
		if strings.ToLower(string(data[i])) == strings.ToLower(string(key[0])) {
			result++
		}
	}
	fmt.Println(result)*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	fmt.Println(data)
	for i := 0; i < len(data); i=i+8 {
		tmp := make([]string, 8)
		for j := i; j < i+8; j++ {
			if j >= len(data) {
				tmp = append(tmp, string('0'))
			} else {
				tmp = append(tmp, string(data[j]))
			}
		}
		str := strings.Join(tmp, "")
		fmt.Printf("%s\n", str)
	}*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	temp := input.Text()
	// 0:代表系统自己判断，0x:16进制、0：8进制、其他十进制，32：接收为int32类型
	res, _ := strconv.ParseInt(temp, 0, 32)
	fmt.Print(res)*/

	/*var value int
	fmt.Scanf("%d", &value)
	factor(value)*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size := input.Text()
	l , _ := strconv.Atoi(size)
	result := make(map[int]int)
	for i := 0; i < l; i++ {
		input.Scan()
		data := strings.Split(input.Text(), " ")
		if len(data) != 2 {
			continue
		}

		key, _ := strconv.Atoi(data[0])
		value, _ := strconv.Atoi(data[1])
		result[key] += value
	}

	type KeyValue struct {
		key int
		value string
	}
	var keyValues []KeyValue
	for key, value := range result {
		keyValues = append(keyValues, KeyValue{
			key:   key,
			value: fmt.Sprintf("%d %d\n", key, value),
		})
	}
	sort.Slice(keyValues, func(i, j int) bool {
		return keyValues[i].key < keyValues[j].key
	})

	for _, value := range keyValues {
		fmt.Printf("%s", value.value)
	}*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	tmp := make(map[byte]byte)
	result := 0
	for i := 0; i < len(data); i++ {
		if _, ok := tmp[data[i]]; !ok {
			result++
			tmp[data[i]] = data[i]
		}
	}
	fmt.Printf("%d", result)*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	var result []string
	for i := len(data)-1; i >=0 ; i-- {
		result = append(result, string(data[i]))
	}
	fmt.Printf("%s", strings.Join(result, ""))*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	var result []string
	tmp := strings.Split(data, " ")
	for i := len(tmp)-1; i >=0 ; i-- {
		result = append(result, tmp[i])
	}
	fmt.Printf("%s", strings.Join(result, " "))*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size := input.Text()
	l , _ := strconv.Atoi(size)
	var result []string
	for i := 0; i < l; i++ {
		input.Scan()
		result = append(result, input.Text())
	}
	sort.Strings(result)
	for _, value := range result {
		fmt.Printf("%s\n", value)
	}*/

	/*input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size := input.Text()
	l , _ := strconv.Atoi(size)
	result := 0
	for l > 0 {
		value := l % 2
		if value != 0 {
			result++
		}
		l = l / 2
	}
	fmt.Printf("%d\n", result)*/

	/*input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		// 第一遍：数据存入map
		map1 := make(map[byte]int)
		data := input.Text()
		min := len(data)
		for i := 0; i < len(data); i++ {
			map1[data[i]]++
		}
		// 第二遍：找到出现最少的次数
		for _, value := range map1 {
			if value < min {
				min = value
			}
		}
		// 第三遍：顺序输出出现次数不是最少的
		var result []string
		for i := 0; i < len(data); i++ {
			if map1[data[i]] == min {
				continue
			}
			result = append(result, string(data[i]))
		}
		fmt.Printf("%s", strings.Join(result, ""))
	}*/

	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		s := input.Text()
		l1 := make([]byte, 0)
		for i := 0; i < len(s); i++ {
			if s[i] >= 'a' && s[i] <= 'z' {
				l1 = append(l1, s[i])
			} else if s[i] >= 'A' && s[i] <= 'Z' {
				l1 = append(l1, s[i])
			} else {
				l1 = append(l1, ' ')
			}
		}
		s = string(l1)
		li := strings.Split(s, " ")
		l := len(li)
		for i := 0; i < len(li); i++ {
			fmt.Printf("%s ", li[l-i-1])
		}
		fmt.Printf("\n")
	}
}

func isWord(data byte) bool {
	return (data > 'a' && data < 'z') || (data > 'A' && data < 'Z')
}

func factor(value int) {
	for i := 2; i*i <= value; i++ {
		if (value % i) == 0 {
			fmt.Printf("%d ", i)
			factor(value / i)
			return
		}
	}

	fmt.Printf("%d", value)
}

func less(a, b string) bool {
	l1 := len(a)
	l2 := len(b)
	var min int
	if l1 < l2 {
		min = l1
	} else {
		min = l2
	}

	for i := 0; i < min; i++ {
		if a[i] < b[i] {
			return true
		}
	}

	return false
}
