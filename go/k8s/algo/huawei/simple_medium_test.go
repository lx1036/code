package huawei

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// https://www.nowcoder.com/practice/3ab09737afb645cc82c35d56a5ce802a?tpId=37&tqId=21230&rp=1&ru=/exam/oj/ta&qru=/exam/oj/ta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D1%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=1&judgeStatus=undefined&tags=&title=
// 取近似值
func hj7() {
	var f float64
	n, err := fmt.Scan(&f)
	if err != nil {
		return
	}

	if n == 0 {
		return
	} else {
		fmt.Printf("%d\n", getInt(f))
	}
}
func getInt(f float64) int {
	return int(math.Floor(f))
}

// https://www.nowcoder.com/practice/253986e66d114d378ae8de2e6c4577c1?tpId=37&tqId=21232&rp=1&ru=/exam/oj/ta&qru=/exam/oj/ta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D1%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=1&judgeStatus=undefined&tags=&title=
// 提取不重复的整数
func hj9() {
	var in int
	fmt.Scanf("%d", &in)
	fmt.Printf("%d", getNonDuplicateInt(in))
}
func getNonDuplicateInt(f int) int {
	result := 0
	tmp := make(map[int]int)
	cur := f
	for cur > 0 {
		value := cur % 10
		if _, ok := tmp[value]; ok {
			cur = cur / 10
			continue
		} else {
			tmp[value] = value
		}

		result = result*10 + value
		cur = cur / 10
	}

	return result
}
func TestHJ9(test *testing.T) {
	ans := getNonDuplicateInt(9876673)
	assert.Equal(test, 37689, ans)
}

// https://www.nowcoder.com/practice/a30bbc1a0aca4c27b86dd88868de4a4a?tpId=37&tqId=21232&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D1%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=1&judgeStatus=undefined&tags=&title=
// 截取字符串
func hj46() {
	var in string
	var k int
	fmt.Scanf("%s", &in)
	fmt.Scanf("%d", &k)
	fmt.Printf("%s", getSubstring(in, k))
}
func getSubstring(in string, k int) string {
	var result []string
	for i := 0; i < k; i++ {
		result = append(result, string(in[i]))
	}

	return strings.Join(result, "")
}
func TestHJ46(test *testing.T) {
	ans := getSubstring("abABCcDEF", 6)
	assert.Equal(test, "abABCc", ans)
}

// https://www.nowcoder.com/practice/69ef2267aafd4d52b250a272fd27052c?tpId=37&tags=&title=&difficulty=1&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D1%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// 输入n个整数，输出其中最小的k个
func hj58() {
	var num []int
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	inputs := strings.Split(input.Text(), " ")
	b, _ := strconv.Atoi(inputs[1])
	input.Scan()
	inputss := strings.Split(input.Text(), " ")
	for _, v := range inputss {
		temp, _ := strconv.Atoi(v)
		num = append(num, temp)
	}

	sort.Ints(num)
	for i := 0; i < b; i++ {
		fmt.Printf("%d ", num[i])
	}
}

// https://www.nowcoder.com/practice/dd0c6b26c9e541f5b935047ff4156309?tpId=37&tags=&title=&difficulty=1&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D1%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// 输入整型数组和排序标识，对其元素按照升序或降序
func hj101() {
	var n int
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	n, _ = strconv.Atoi(input.Text())

	nums := make([]int, 0)
	input.Scan()
	inputs := strings.Split(input.Text(), " ")
	for i := 0; i < n; i++ {
		data, _ := strconv.Atoi(inputs[i])
		nums = append(nums, data)
	}

	input.Scan()
	tag, _ := strconv.Atoi(input.Text())

	if tag == 0 { // 升序
		for i := 1; i < len(nums); i++ {
			for j := i; j > 0 && nums[j] < nums[j-1]; j-- {
				Swap(nums, j, j-1)
			}
		}
	} else {
		for i := 1; i < len(nums); i++ {
			for j := i; j > 0 && nums[j] > nums[j-1]; j-- {
				Swap(nums, j, j-1)
			}
		}
	}
	for _, v := range nums {
		fmt.Printf("%d ", v)
	}
}
func Swap(a []int, b int, c int) {
	var temp int = a[b]
	a[b] = a[c]
	a[c] = temp
}

// https://www.nowcoder.com/practice/8c949ea5f36f422594b306a2300315da?tpId=37&tqId=21224&rp=1&ru=/exam/oj/ta&qru=/exam/oj/ta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// 字符串最后一个单词的长度
func hj1() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	inputs := strings.Split(input.Text(), " ")

	//fmt.Println(inputs)
	last := inputs[len(inputs)-1]
	fmt.Printf("%d\n", len(last))

}

// https://www.nowcoder.com/practice/a35ce98431874e3a820dbe4b2d0508b1?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// 计算某字符出现次数
func hj2() {
	input := bufio.NewScanner(os.Stdin)
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

	fmt.Println(result)
}

// https://www.nowcoder.com/practice/d9162298cb5a437aad722fccccaae8a7?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// 字符串分隔
func hj4() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()

	for i := 0; i < len(data); i = i + 8 {
		tmp := make([]string, 8)
		for j := i; j < 8; j++ {
			if j >= len(data) {
				tmp = append(tmp, string('0'))
			} else {
				tmp = append(tmp, string(data[j]))
			}
		}
		str := strings.Join(tmp, "")
		fmt.Printf("%s\n", str)
	}
}

// https://www.nowcoder.com/practice/8f3df50d2b9043208c5eed283d1d4da6?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// 进制转换
func hj5() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	temp := input.Text()
	// 0:代表系统自己判断，0x:16进制、0：8进制、其他十进制，32：接收为int32类型
	res, _ := strconv.ParseInt(temp, 0, 32)
	fmt.Print(res)
}

// https://www.nowcoder.com/practice/196534628ca6490ebce2e336b47b3607?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ6 质数因子
func hj6() {
	var value int
	fmt.Scanf("%d", &value)
	factor(value)
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

// https://www.nowcoder.com/practice/de044e89123f4a7482bd2b214a685201?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ8 合并表记录
func hj8() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size := input.Text()
	l, _ := strconv.Atoi(size)
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
		key   int
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
	}
}

// https://www.nowcoder.com/practice/eb94f6a5b2ba49c6ac72d40b5ce95f50?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ10 字符个数统计
func hj10() {
	input := bufio.NewScanner(os.Stdin)
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
	fmt.Printf("%d", result)
}

// https://www.nowcoder.com/practice/ae809795fca34687a48b172186e3dafe?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ11 数字颠倒
func hj11() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	var result []string
	for i := len(data) - 1; i >= 0; i-- {
		result = append(result, string(data[i]))
	}
	fmt.Printf("%s", strings.Join(result, ""))
}

// https://www.nowcoder.com/practice/e45e078701ab4e4cb49393ae30f1bb04?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ12 字符串反转
func hj12() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	var result []string
	for i := len(data) - 1; i >= 0; i-- {
		result = append(result, string(data[i]))
	}
	fmt.Printf("%s", strings.Join(result, ""))
}

// https://www.nowcoder.com/practice/48b3cb4e3c694d9da5526e6255bb73c3?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ13 句子逆序
func hj13() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	var result []string
	tmp := strings.Split(data, " ")
	for i := len(tmp) - 1; i >= 0; i-- {
		result = append(result, tmp[i])
	}
	fmt.Printf("%s", strings.Join(result, " "))
}

// https://www.nowcoder.com/practice/5af18ba2eb45443aa91a11e848aa6723?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ14 字符串排序
func hj14() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size := input.Text()
	l, _ := strconv.Atoi(size)
	var result []string
	for i := 0; i < l; i++ {
		input.Scan()
		result = append(result, input.Text())
	}
	sort.Strings(result)
	for _, value := range result {
		fmt.Printf("%s\n", value)
	}
}

// https://www.nowcoder.com/practice/440f16e490a0404786865e99c6ad91c9?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ15 求int型正整数在内存中存储时1的个数
func hj15() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size := input.Text()
	l, _ := strconv.Atoi(size)
	result := 0
	for l > 0 {
		value := l % 2
		if value != 0 {
			result++
		}
		l = l / 2
	}
	fmt.Printf("%d\n", result)
}

// https://www.nowcoder.com/practice/7960b5038a2142a18e27e4c733855dac?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ21 简单密码
func hj21() {
	var str string
	H := make(map[string]string)
	fmt.Scanf("%s\n", &str)
	H["abc"] = "2"
	H["def"] = "3"
	H["ghi"] = "4"
	H["jkl"] = "5"
	H["mno"] = "6"
	H["pqrs"] = "7"
	H["tuv"] = "8"
	H["wxyz"] = "9"
	for _, v := range str {
		for k, v2 := range H {
			if strings.Contains(k, string(v)) {
				fmt.Printf("%s", v2)
			}
		}
		if v >= 'A' && v < 'Z' {
			var str1 string = strings.ToLower(string(v + 1))
			fmt.Printf("%s", str1)
		} else if v == 'Z' {
			fmt.Printf("%s", "a")
		} else if v >= '0' && v <= '9' {
			fmt.Printf("%s", string(v))
		}
	}
}

// https://www.nowcoder.com/practice/fe298c55694f4ed39e256170ff2c205f?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ22 汽水瓶
func hj22() {
	input := bufio.NewScanner(os.Stdin)
	for {
		input.Scan()
		str, _ := strconv.Atoi(input.Text())
		if str == 0 {
			break
		}
		fmt.Println(str / 2)
	}
}

// https://www.nowcoder.com/practice/05182d328eb848dda7fdd5e029a56da9?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ23 删除字符串中出现次数最少的字符
func hj23() {
	input := bufio.NewScanner(os.Stdin)
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
	}
}

// https://www.nowcoder.com/practice/81544a4989df4109b33c2d65037c5836?tpId=37&tqId=21224&rp=1&ru=%2Fexam%2Foj%2Fta&qru=%2Fexam%2Foj%2Fta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=2&judgeStatus=undefined&tags=&title=
// HJ31 单词倒排
func hj31() {
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

// https://www.nowcoder.com/practice/2de4127fda5e46858aa85d254af43941?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ34 图片整理
func hj34() {
	var inputSlice []string
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		s := input.Text()
		for i := 0; i < len(s); i++ {
			inputSlice = append(inputSlice, string(s[i]))
		}
		sort.Strings(inputSlice)
		res := ""
		for _, v := range inputSlice {
			res += string(v)
		}
		fmt.Println(res)
	}
}

// https://www.nowcoder.com/practice/649b210ef44446e3b1cd1be6fa4cab5e?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ35 蛇形矩阵
func hj35() {
	var n int
	fmt.Scan(&n)
	a := make([][]int, n)
	for i := 0; i < n; i++ {
		a[i] = make([]int, n)
	}
	temp := 0
	for i := 0; i < n; i++ {
		for j := i; j >= 0; j-- {
			temp++
			a[j][i-j] = temp
		}
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if a[i][j] != 0 {
				fmt.Printf("%d ", a[i][j])
			}
		}
		fmt.Println()
	}
}

// https://www.nowcoder.com/practice/1221ec77125d4370833fd3ad5ba72395?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ37 统计每个月兔子的总数
func hj37() {
	var n int
	temp := 0
	fmt.Scan(&n)
	for i := 1; i < n+1; i++ {
		temp = f(i)
	}
	fmt.Println(temp)
}
func f(n int) int {
	if n < 2 {
		return n
	}
	return f(n-2) + f(n-1)
}

// https://www.nowcoder.com/practice/539054b4c33b4776bc350155f7abd8f5?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ40 统计字符
func hj40() {
	var a, b, c, d int = 0, 0, 0, -1
	inputReader := bufio.NewReader(os.Stdin)
	input, err := inputReader.ReadString('\n')
	if err != nil {
		return
	}
	for _, v := range input {
		if (v >= 'a' && v <= 'z') || (v >= 'A' && v <= 'Z') {
			a++
		} else if v == ' ' {
			b++
		} else if v >= '0' && v <= '9' {
			c++
		} else {
			d++
		}
	}
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(d)
}

// https://www.nowcoder.com/practice/54404a78aec1435a81150f15f899417d?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ51 输出单向链表中倒数第k个结点
func hj51() {
	for {
		var num int
		_, e := fmt.Scanf("%d", &num)
		if e != nil {
			break
		}
		slice := make([]int, num)
		for i := 0; i < num; i++ {
			fmt.Scanf("%d", &slice[i])
		}
		var index int
		fmt.Scanf("%d", &index)
		if index > 0 {
			fmt.Println(slice[num-index])
		} else {
			fmt.Println(0)
		}
	}
}

// https://www.nowcoder.com/practice/8ef655edf42d4e08b44be4d777edbf43?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ53 杨辉三角的变形
func hj53() {
	var n int
	fmt.Scan(&n)
	if n < 3 {
		fmt.Print(-1)

	} else {
		switch (n - 2) % 4 {
		case 1:
			fmt.Print(2)
		case 2:
			fmt.Print(3)
		case 3:
			fmt.Print(2)
		case 0:
			fmt.Print(4)
		}
	}
}

// https://www.nowcoder.com/practice/9566499a2e1546c0a257e885dfdbf30d?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ54 表达式求值
type node struct {
	NType byte // 类型 0 数字 别的符号
	Val   int  //
}

func hj54() {
	var s string
	fmt.Scanln(&s)

	var data []node
	var num []byte
	for i := 0; i < len(s); i++ { // 先把数字(包括负数)、符号、括号 解析出来
		if s[i] == '-' && (i == 0 || (i > 0 && s[i-1] == '(')) {
			num = append(num, '-') // 负数的情况，暂存负号
			continue
		} else if '0' <= s[i] && s[i] <= '9' {
			num = append(num, s[i]) // 数字的情况，暂存数字
			continue
		}

		if len(num) > 0 { // 各种符号的情况
			data = append(data, node{0, bsToNum(num)}) // 将数字解析出来
			num = []byte{}
		}

		if s[i] == ')' { // 遇到一堆完整的括号了，先计算这个括号里面的额值，然后替换表达式
			j := len(data) - 2 // 找到最近的左括号 来做计算
			for ; data[j].NType != '('; j-- {
			}

			data[j] = node{0, calc(data[j+1:])}
			data = data[:j+1]
		} else {
			data = append(data, node{s[i], 0})
		}
	}

	if len(num) > 0 {
		data = append(data, node{0, bsToNum(num)})
	}

	fmt.Println(calc(data))
}

func calc(data []node) int { // 无括号表达式的计算
	var afterMulAndDivide []node
	for i := 0; i < len(data); { // 先乘除
		if data[i].NType == '*' {
			afterMulAndDivide[len(afterMulAndDivide)-1].Val *= data[i+1].Val
			i += 2
		} else if data[i].NType == '/' {
			afterMulAndDivide[len(afterMulAndDivide)-1].Val /= data[i+1].Val
			i += 2
		} else {
			afterMulAndDivide = append(afterMulAndDivide, data[i])
			i++
		}
	}

	result := afterMulAndDivide[0].Val
	for i := 1; i < len(afterMulAndDivide); { // 后加减
		if afterMulAndDivide[i].NType == '+' {
			result += afterMulAndDivide[i+1].Val
		} else if afterMulAndDivide[i].NType == '-' {
			result -= afterMulAndDivide[i+1].Val
		}
		i += 2
	}

	return result
}
func bsToNum(bs []byte) int {
	num, _ := strconv.Atoi(string(bs))
	return num
}

// https://www.nowcoder.com/practice/7299c12e6abb437c87ad3e712383ff84?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ56 完全数计算
func hj56() {
	for {
		n := 0
		_, err := fmt.Scan(&n)
		n1 := 0
		if err != nil {
			break
		}
		for i := 1; i <= n; i++ {
			li := make([]int, 0)
			sum := 0
			for j := 1; j < i; j++ {
				if i%j == 0 {
					li = append(li, j)
				}
			}
			for _, k := range li {
				sum += k
			}
			if sum == i {
				n1++
			}
		}
		fmt.Println(n1)
	}
}

// https://www.nowcoder.com/practice/f8538f9ae3f1484fb137789dec6eedb9?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ60 查找组成一个偶数最接近的两个素数
func hj60() {
	for {
		num := 0
		_, err := fmt.Scan(&num)
		if err != nil {
			break
		}
		i, j := num/2, num/2
		for {
			if sushu(i) && sushu(j) {
				fmt.Println(i)
				fmt.Println(j)
				break
			}
			i--
			j++
		}
	}
}
func sushu(num1 int) bool {
	num := 0
	for i := 1; i <= num1; i++ {
		if num1%i == 0 {
			num++
		}
	}
	if num == 2 {
		return true
	}
	return false
}

// https://www.nowcoder.com/practice/bfd8234bb5e84be0b493656e390bdebf?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ61 放苹果
func hj61() {
	var m, n int
	fmt.Scan(&m, &n)
	res := Dp(m, n)
	fmt.Println(res)
}

func Dp(m, n int) int {
	dp := make([][]int, m+1)
	//初始化
	for i := 0; i < m+1; i++ {
		dp[i] = make([]int, n+1)
		dp[i][1] = 1
		dp[i][0] = 1
	}
	for j := 0; j < n+1; j++ {
		dp[1][j] = 1
		dp[0][j] = 1
	}
	for i := 2; i < m+1; i++ {
		for j := 2; j < n+1; j++ {
			if i < j {
				dp[i][j] = dp[i][j-1]
			} else {
				dp[i][j] = dp[i-j][j] + dp[i][j-1]
			}
		}
	}

	return dp[m][n]
}

// https://www.nowcoder.com/practice/1b46eb4cf3fa49b9965ac3c2c1caf5ad?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ62 查找输入整数二进制中1的个数
func hj62() {
	for {
		l := 0
		_, err := fmt.Scan(&l)
		if err != nil {
			break
		}
		result := 0
		for l > 0 {
			value := l % 2
			if value != 0 {
				result++
			}
			l = l / 2
		}
		fmt.Printf("%d\n", result)
	}
}

// https://www.nowcoder.com/practice/74c493f094304ea2bda37d0dc40dc85b?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ72 百钱买百鸡问题
func hj72() {
	var n int
	fmt.Scan(&n)
	for i := 0; i <= 3; i++ {
		fmt.Printf("%d %d %d\n", 4*i, 25-7*i, 75+3*i)
	}
}

// https://www.nowcoder.com/practice/769d45d455fe40b385ba32f97e7bcded?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ73 计算日期到天数转换
func hj73() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		inputs := strings.Split(input.Text(), " ")
		year, _ := strconv.Atoi(inputs[0])
		month, _ := strconv.Atoi(inputs[1])
		day, _ := strconv.Atoi(inputs[2])
		res := 0
		m := []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
		if year%400 == 0 || year%100 != 0 && year%4 == 0 { // 闰年
			for i := 0; i < month-1; i++ {
				res += m[i]
			}
			res += day
		} else {
			m[1] = 28
			for i := 0; i < month-1; i++ {
				res += m[i]
			}
			res += day
		}
		fmt.Println(res)
	}
}

// https://www.nowcoder.com/practice/dbace3a5b3c4480e86ee3277f3fe1e85?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ76 尼科彻斯定理
func hj76() {
	var in int
	fmt.Scanln(&in)
	fmt.Println(VerifyingNicochaseTheorem(in))
}

//转化为数学问题，即等差数列求和问题，公差为2，由规律可知项数n等于输入的m
func VerifyingNicochaseTheorem(in int) string {
	sum := math.Pow(float64(in), 3)
	n := in
	a1 := int(sum)/n - (n - 1)
	var s string
	s += strconv.Itoa(a1)
	for i := 1; i < n; i++ {
		a1 = a1 + 2
		s += "+"
		s += strconv.Itoa(a1)
	}
	return s
}

// https://www.nowcoder.com/practice/c4f11ea2c886429faf91decfaf6a310b?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ80 整型数组合并
func hj80() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		inputss := make([]int, 0)
		input.Scan()
		nn := strings.Split(input.Text(), " ")
		input.Scan()
		input.Scan()
		mm := strings.Split(input.Text(), " ")
		mm = append(mm, nn...)

		for _, i := range mm {
			temp, _ := strconv.Atoi(i)
			inputss = append(inputss, temp)
		}

		sort.Ints(inputss)
		fmt.Print(inputss[0])
		for i := 1; i < len(inputss); i++ {
			if inputss[i] == inputss[i-1] {
				continue
			}
			fmt.Print(inputss[i])
		}
	}
}

// https://www.nowcoder.com/practice/22fdeb9610ef426f9505e3ab60164c93?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ81 字符串字符匹配
func hj81() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		a := input.Text()
		input.Scan()
		aaa := input.Text()
		mapa := make(map[string]bool)
		flag := true
		for _, a1 := range aaa {
			mapa[string(a1)] = true
		}
		for _, a2 := range a {
			if _, ok := mapa[string(a2)]; !ok {
				flag = false
			}
		}
		fmt.Println(flag)
	}
}

// https://www.nowcoder.com/practice/2f8c17bec47e416897ce4b9aa560b7f4?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ83 二维数组操作

// https://www.nowcoder.com/practice/434414efe5ea48e5b06ebf2b35434a9c?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ84 统计大写字母个数
func hj84() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		inputs := strings.Split(input.Text(), "")
		res := 0
		for _, s := range inputs {
			if s >= "A" && s <= "Z" {
				res++
			}
		}
		fmt.Println(res)
	}
}

// https://www.nowcoder.com/practice/12e081cd10ee4794a2bd70c7d68f5507?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ85 最长回文子串
func hj85() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		str := longestPalindrome(input.Text())
		fmt.Println(len(str))
	}
}
func longestPalindrome(s string) string {
	n := len(s)
	if n < 2 {
		return s
	}
	start, end := 0, 0
	for i := 0; i < n; {
		l, r := i, i
		//如果字符串相同则分别冲前一个和后一个开始回文
		for r < n-1 && s[r] == s[r+1] {
			r++
		}
		i = r + 1
		for l > 0 && r < n-1 && s[l-1] == s[r+1] {
			l--
			r++
		}
		if end < r-l {
			start = l
			end = r - l
		}
	}
	return s[start : start+end+1]
}

// https://www.nowcoder.com/practice/4b1658fd8ffb4217bc3b7e85a38cfaf2?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ86 求最大连续bit数
// 输入描述：
//输入一个int类型数字
//输出描述：
//输出转成二进制之后连续1的个数
func hj86() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		inputs, _ := strconv.Atoi(input.Text())
		res := 0
		max := 0
		for inputs != 0 {
			temp := inputs % 2
			inputs /= 2
			if temp == 1 {
				res++
				if res > max {
					max = res
				}
			} else {
				res = 0
			}
		}
		fmt.Println(max)
	}
}

// https://www.nowcoder.com/practice/52d382c2a7164767bca2064c1c9d5361?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ87 密码强度等级
//输入描述：
//输入一个string的密码
//输出描述：
//输出密码等级
func hj87() {
	for {
		var a string
		if _, err := fmt.Scan(&a); err != nil {
			break
		}
		solution(a)
	}
}
func solution(s string) {
	r := []rune(s)
	total := 0
	c := count(r)
	total = judgeLen(c) + judgeNum(c) + judgeAlph(c) + judgeSymbol(c) + award(c)
	grade(total)
}

// 评分
func grade(n int) {
	switch {
	case n >= 90:
		fmt.Println("VERY_SECURE")
	case n >= 80:
		fmt.Println("SECURE")
	case n >= 70:
		fmt.Println("VERY_STRONG")
	case n >= 60:
		fmt.Println("STRONG")
	case n >= 50:
		fmt.Println("AVERAGE")
	case n >= 25:
		fmt.Println("WEAK")
	default:
		fmt.Println("VERY_WEAK")
	}
}

type Counter struct {
	length, num, lower, upper, symbol int
}

// 一次遍历统计所有次数
func count(r []rune) *Counter {
	var c Counter
	c.length = len(r)
	for _, v := range r {
		switch {
		case v >= 'a' && v <= 'z':
			c.lower++
		case v >= 'A' && v <= 'Z':
			c.upper++
		case v >= '0' && v <= '9':
			c.num++
		default:
			c.symbol++
		}
	}
	return &c
}

// 长度评分
func judgeLen(c *Counter) int {
	l := c.length
	switch {
	case l <= 4:
		return 5
	case l <= 7:
		return 10
	default:
		return 25
	}
}

// 数字评分
func judgeNum(c *Counter) int {
	n := c.num
	switch {
	case n == 0:
		return 0
	case n == 1:
		return 10
	default:
		return 20
	}
}

// 字母评分
func judgeAlph(c *Counter) int {
	l, u := c.lower, c.upper
	switch {
	case l+u == 0:
		return 0
	case l == 0 || u == 0:
		return 10
	default:
		return 20
	}
}

// 符号评分
func judgeSymbol(c *Counter) int {
	s := c.symbol
	switch {
	case s == 0:
		return 0
	case s == 1:
		return 10
	default:
		return 25
	}
}

// 奖励
func award(c *Counter) int {
	l, u, n, s := c.lower, c.upper, c.num, c.symbol
	switch {
	case l > 0 && u > 0 && n > 0 && s > 0:
		return 5
	case l+u > 0 && n > 0 && s > 0:
		return 3
	case l+u > 0 && n > 0:
		return 2
	default:
		return 0
	}
}

// https://www.nowcoder.com/practice/e2a22f0305eb4f2f9846e7d644dba09b?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ91 走方格的方案数
// INFO: DP: 动态规划
// 请计算n*m的棋盘格子（n为横向的格子数，m为竖向的格子数）从棋盘左上角出发沿着边缘线从左上角走到右下角，总共有多少种走法，要求不能走回头路，即：只能往右和往下走，不能往左和往上走。
//注：沿棋盘格之间的边缘线行走
//输入描述：
//输入两个正整数n和m，用空格隔开。(1≤n,m≤8)
//输出描述：
//输出一行结果
func hj91() {
	var n, m int
	fmt.Scan(&n, &m)
	res := hj91DP(n, m)
	fmt.Println(res)
}
func hj91DP(n, m int) int {
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 0; i < n+1; i++ {
		dp[i][0] = 1
	}
	for j := 0; j < m+1; j++ {
		dp[0][j] = 1
	}
	for i := 1; i < n+1; i++ {
		for j := 1; j < m+1; j++ {
			dp[i][j] = dp[i-1][j] + dp[i][j-1]
		}
	}

	return dp[n][m]
}

// https://www.nowcoder.com/practice/3350d379a5d44054b219de7af6708894?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ94 记票统计
// 请实现一个计票统计系统。你会收到很多投票，其中有合法的也有不合法的，请统计每个候选人得票的数量以及不合法的票数。
//输入描述：
//第一行输入候选人的人数n，第二行输入n个候选人的名字（均为大写字母的字符串），第三行输入投票人的人数，第四行输入投票。
//输出描述：
//按照输入的顺序，每行输出候选人的名字和得票数量（以" : "隔开，注：英文冒号左右两边都有一个空格！），最后一行输出不合法的票数，格式为"Invalid : "+不合法的票数。
func hj94() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		input.Scan()
		s := strings.Split(input.Text(), " ")
		input.Scan()
		input.Scan()
		ss := strings.Split(input.Text(), " ")
		map1 := make(map[string]int)
		for _, ins := range s {
			map1[ins]++
		}
		for _, ins := range ss {
			if _, ok := map1[ins]; ok {
				map1[ins]++
			} else {
				map1["invalid"]++
			}
		}
		for _, ins := range s {
			fmt.Printf("%s : %d\n", ins, map1[ins]-1)
		}
		fmt.Printf("%s : %d\n", "Invalid", map1["invalid"])
	}
}

// https://www.nowcoder.com/practice/637062df51674de8ba464e792d1a0ac6?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ96 表示数字
// 将一个字符串中所有的整数前后加上符号“*”，其他字符保持不变。连续的数字视为一个整数。
//输入描述：
//输入一个字符串
//输出描述：
//字符中所有出现的数字前后加上符号“*”，其他字符保持不变
func hj96() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		ss := strings.Split(input.Text(), "")
		var res strings.Builder
		temp := ""
		for _, s := range ss {
			if (s > "9" || s < "0") && temp == "" { //符号
				res.WriteString(s)
			} else if (s > "9" || s < "0") && temp != "" {
				res.WriteString("*")
				res.WriteString(temp)
				res.WriteString("*")
				res.WriteString(s)
				temp = ""
			} else if s <= "9" && s >= "0" {
				temp = temp + s
			}
		}
		if temp != "" {
			res.WriteString("*")
			res.WriteString(temp)
			res.WriteString("*")
		}
		fmt.Println(res.String())
	}
}

// https://www.nowcoder.com/practice/6abde6ffcc354ea1a8333836bd6876b8?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ97 记负均正
// 首先输入要输入的整数个数n，然后输入n个整数。输出为n个整数中负数的个数，和所有正整数的平均值，结果保留一位小数。
// 0即不是正整数，也不是负数，不计入计算。如果没有正数，则平均值为0。
//输入描述：
//首先输入一个正整数n，
//然后输入n个整数。
//输出描述：
//输出负数的个数，和所有正整数的平均值。
func hj97() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		input.Scan()
		inputs := strings.Split(input.Text(), " ")
		sum := 0
		futag := 0
		zhengtag := 0
		for _, s := range inputs {
			data, _ := strconv.Atoi(s)
			if data > 0 {
				sum += data
				zhengtag++
			} else if data < 0 {
				futag++
			}
		}
		fmt.Printf("%d ", futag)
		if zhengtag == 0 {
			fmt.Println("0.0")
		} else {
			res := float64(sum) / float64(zhengtag)
			ress := strconv.FormatFloat(res, 'f', 1, 64)
			fmt.Println(ress)
		}
	}
}

// https://www.nowcoder.com/practice/88ddd31618f04514ae3a689e83f3ab8e?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ99 自守数
// 自守数是指一个数的平方的尾数等于该数自身的自然数。例如：25^2 = 625，76^2 = 5776，9376^2 = 87909376。请求出n(包括n)以内的自守数的个数
// 输入描述：
//int型整数
//输出描述：
//n以内自守数的数量。
func hj99() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		data, _ := strconv.Atoi(input.Text())
		res := 0
		for i := 0; i < data+1; i++ {
			// HasSuffix:判断是否有后缀字符串
			if strings.HasSuffix(strconv.Itoa(i*i), strconv.Itoa(i)) {
				res++
			}
		}
		fmt.Print(res)
	}
}

// https://www.nowcoder.com/practice/f792cb014ed0474fb8f53389e7d9c07f?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ100 等差数列
// 等差数列 2，5，8，11，14。。。。
//（从 2 开始的 3 为公差的等差数列）
//输出求等差数列前n项和
func hj100() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		n, _ := strconv.Atoi(input.Text())
		res := 2
		temp := 2
		for i := 0; i < n-1; i++ {
			temp += 3
			res += temp
		}
		fmt.Println(res)
	}
}

// https://www.nowcoder.com/practice/c1f9561de1e240099bdb904765da9ad0?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ102 字符统计
// 输入描述：
//一个只包含小写英文字母和数字的字符串。
//输出描述：
//一个字符串，为不同字母出现次数的降序表示。若出现次数相同，则按ASCII码的升序输出。
func hj102() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		text := input.Text()
		map1 := map[byte]int{}
		var bytes []byte
		for i := range text {
			if _, ok := map1[text[i]]; !ok {
				bytes = append(bytes, text[i])
			}
			map1[text[i]]++
		}
		sort.Slice(bytes, func(i, j int) bool {
			if map1[bytes[i]] == map1[bytes[j]] {
				return bytes[i] < bytes[j]
			}
			return map1[bytes[i]] > map1[bytes[j]]
		})
		fmt.Println(string(bytes))
	}
}

// https://www.nowcoder.com/practice/64f6f222499c4c94b338e588592b6a62?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ105 记负均正II
//输入 n 个整型数，统计其中的负数个数并求所有非负数的平均值，结果保留一位小数，如果没有非负数，则平均值为0
//本题有多组输入数据，输入到文件末尾。
// 输入描述：
//输入任意个整数，每行输入一个。
//输出描述：
//输出负数个数以及所有非负数的平均值
func hj105() {
	count, no := 0, 0
	var sum float64 = 0
	for {
		var num float64 = 0
		_, e := fmt.Scanf("%f", &num)
		if e != nil {
			break
		}
		if num < 0 {
			count++
		} else {
			no++
			sum += num
		}
	}
	fmt.Println(count)
	if no > 0 {
		fmt.Printf("%.1f \n", sum/float64(no))
	} else {
		fmt.Println("0.0")
	}
}

// https://www.nowcoder.com/practice/cc57022cb4194697ac30bcb566aeb47b?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ106 字符逆序
func hj106() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	inputs := strings.Split(input.Text(), "")
	for i := len(inputs) - 1; i >= 0; i-- {
		fmt.Print(inputs[i])
	}
}

// https://www.nowcoder.com/practice/22948c2cad484e0291350abad86136c3?tpId=37&tags=&title=&difficulty=2&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D2%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ108 求最小公倍数
// 输入描述：
//输入两个正整数A和B。
//输出描述：
//输出A和B的最小公倍数。
func hj108() {
	var a, b int
	fmt.Scanf("%d %d", &a, &b)
	fmt.Println(LeastCommonMultiple(a, b))
}

// 辗转相除法计算最大公约数
func GreatestCommonDivisor(a, b int) int {
	// 以 b 为判断条件，进行辗转相除
	// 如果 a % b == 0，证明 b 是 a 的一个约数
	// b 就是是 a 和 b 中的最大公约数
	// 如果 a < b 时，a % b 的结果就是 a，这里还保证了数之间的交换。
	// 使得最终一定有 b = 大数 % 小数 从而进行判断
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// 利用公式： (a, b)最大公约数 * (a, b)最小公倍数 = a * b
func LeastCommonMultiple(a, b int) int {
	return a * b / GreatestCommonDivisor(a, b)
}

///////////////////////// Medium ////////////////////////////////

// https://www.nowcoder.com/practice/f9c6f980eeec43ef85be20755ddbeaf4?tpId=37&tqId=21239&rp=1&ru=/exam/oj/ta&qru=/exam/oj/ta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=3&judgeStatus=undefined&tags=&title=
// HJ16 购物单
// 输出描述：
// 输出一个正整数，为张强可以获得的最大的满意度。
func hj16() {
	sc := bufio.NewScanner(os.Stdin)
	sc.Scan()
	str := strings.Split(sc.Text(), " ")
	money, _ := strconv.Atoi(str[0])
	n, _ := strconv.Atoi(str[1])
	items := make([]item, n)
	for i := 0; i < n; i++ { //构造物品
		sc.Scan()
		it := strings.Split(sc.Text(), " ")
		v, _ := strconv.Atoi(it[0])
		p, _ := strconv.Atoi(it[1])
		q, _ := strconv.Atoi(it[2])
		items[i] = item{
			v,
			p,
			q,
			nil,
			nil,
		}
	}
	for j, k := range items { //构造主物品
		if k.q != 0 {
			if items[k.q-1].acc1 == nil {
				//                 fmt.Println(k,"acc1")
				items[k.q-1].acc1 = &items[j]
			} else {
				items[k.q-1].acc2 = &items[j]
			}
		}
	}
	//     fmtx.Println("ist",items)
	//由于每个东西只能买一件，并且买了主件才能买附件，同时主键的附件数量是确定的不大于2
	//因此我们可以看成购买就是针对于主件，只不过主件的附件数量不同而已，因为只能买一次
	//因此主件和附件的搭配只能选一种，之后就可以看成01背包问题，只不过背包的物品是多选一的
	matrix := make([][]int, n+1)
	for j, _ := range matrix {
		matrix[j] = make([]int, money+1)
	}
	cnt := 1
	for i := 0; i < n; i++ {
		if items[i].q != 0 { //附件直接跳过
			continue
		}
		//构造各个选择的价格和满意度
		var (
			v0   = items[i].v //只有主件
			myd0 = v0 * items[i].p
			v1   = v0 //主件加附件1
			myd1 = myd0
			v2   = v0 //主件加附件2
			myd2 = myd0
			v3   = v0 //主件加两个附件
			myd3 = myd0
		)
		if items[i].acc1 != nil {
			v1 += items[i].acc1.v
			myd1 += items[i].acc1.v * items[i].acc1.p
		}
		if items[i].acc2 != nil {
			v2 += items[i].acc2.v
			myd2 += items[i].acc2.v * items[i].acc2.p
		}
		if items[i].acc1 != nil && items[i].acc2 != nil {
			v3 = v3 + items[i].acc1.v + items[i].acc2.v
			myd3 = myd3 + items[i].acc1.v*items[i].acc1.p + items[i].acc2.v*items[i].acc2.p
		}
		//         fmt.Println("gz",v0,myd0,v1,myd1,v2,myd2,v3,myd3)

		for j := 1; j <= money; j++ {
			matrix[cnt][j] = matrix[cnt-1][j]
			if j >= v0 {
				matrix[cnt][j] = max(matrix[cnt][j], matrix[cnt-1][j-v0]+myd0)
			}
			if j >= v1 && v1 > v0 {
				matrix[cnt][j] = max(matrix[cnt][j], matrix[cnt-1][j-v1]+myd1)
			}
			if j >= v2 && v2 > v0 {
				matrix[cnt][j] = max(matrix[cnt][j], matrix[cnt-1][j-v2]+myd2)
			}
			if j >= v3 && v3 > v0 {
				matrix[cnt][j] = max(matrix[cnt][j], matrix[cnt-1][j-v3]+myd3)
			}

		}
		cnt++
	}
	fmt.Println(matrix[cnt-1][money])
}

type item struct {
	v    int
	p    int
	q    int
	acc1 *item
	acc2 *item
}

func max(a int, b int) int {
	if a >= b {
		return a
	}
	return b
}

// https://www.nowcoder.com/practice/119bcca3befb405fbe58abe9c532eb29?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ17 坐标移动
// 描述
//开发一个坐标计算工具， A表示向左移动，D表示向右移动，W表示向上移动，S表示向下移动。从（0,0）点开始移动，从输入字符串里面读取一些坐标，并将最终输入结果输出到输出文件里面。
//输入：
//合法坐标为A(或者D或者W或者S) + 数字（两位以内）
//坐标之间以;分隔。
//非法坐标点需要进行丢弃。如AA10;  A1A;  $%$;  YAD; 等。
//下面是一个简单的例子 如：
//A10;S20;W10;D30;X;A1A;B10A11;;A10;
//处理过程：
//起点（0,0）
//+   A10   =  （-10,0）
//+   S20   =  (-10,-20)
//+   W10  =  (-10,-10)
//+   D30  =  (20,-10)
//+   x    =  无效
//+   A1A   =  无效
//+   B10A11   =  无效
//+  一个空 不影响
//+   A10  =  (10,-10)
//结果 （10， -10）
// 输入描述：
//一行字符串
//输出描述：
//最终坐标，以逗号分隔
func hj17() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	s := input.Text()
	x, y := removeCoord(s)
	fmt.Printf("%v,%v", x, y)
}
func removeCoord(s string) (int, int) {
	var x, y int = 0, 0
	s1 := strings.Split(s, ";")
	for i := 0; i < len(s1); i++ {
		s2 := s1[i]
		if len(s2) >= 2 {
			first_num := s2[0]
			if first_num == 'A' || first_num == 'D' || first_num == 'S' || first_num == 'W' {
				s2 = s2[1:]
				num, err := strconv.Atoi(s2)
				if err == nil {
					switch first_num {
					case 'A':
						x -= num
					case 'S':
						y -= num
					case 'W':
						y += num
					case 'D':
						x += num
					}
				}
			}
		}

		continue
	}
	return x, y
}

// https://www.nowcoder.com/practice/184edec193864f0985ad2684fbc86841?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ20 密码验证合格程序
// 描述
//密码要求:
//1.长度超过8位
//2.包括大小写字母.数字.其它符号,以上四种至少三种
//3.不能有长度大于2的包含公共元素的子串重复 （注：其他符号不含空格或换行）
//数据范围：输入的字符串长度满足 1 \le n \le 100 \1≤n≤100
//输入描述：
//一组字符串。
//输出描述：
//如果符合要求输出：OK，否则输出NG
func hj20() {
	var inputHexStr string
	for {
		n, _ := fmt.Scan(&inputHexStr)
		if n == 0 {
			break
		}
		if len(inputHexStr) < 8 {
			fmt.Println(`NG`)
			continue
		}
		var countDiffType map[string]int = make(map[string]int)
		var subStr map[string]int = make(map[string]int, 0)
		var mark int
		// INFO: 检查是否有公共子串
		for i := 0; i < len(inputHexStr)-2; i++ {
			if _, ok := subStr[inputHexStr[i:i+3]]; ok {
				// 存在公共子串
				fmt.Println(`NG`)
				mark = 1
				break
			} else {
				subStr[inputHexStr[i:i+3]] = i
			}
		}
		if mark == 1 {
			continue
		}
		for i := 0; i < len(inputHexStr); i++ {
			singlechar := inputHexStr[i]
			// 小写字母
			if singlechar >= byte('a') && singlechar <= byte('z') {
				if _, ok := countDiffType["Lower"]; !ok {
					countDiffType["Lower"] = 1
				}
				continue
			}
			// 大写字母
			if singlechar >= 65 && singlechar <= 90 {
				if _, ok := countDiffType["Upper"]; !ok {
					countDiffType["Upper"] = 1
				}
				continue
			}
			// 数字
			if singlechar >= 48 && singlechar <= 57 {
				if _, ok := countDiffType["Number"]; !ok {
					countDiffType["Number"] = 1
				}
				continue
			}
			if _, ok := countDiffType["Other"]; !ok {
				countDiffType["Other"] = 1
			}
		}
		if len(countDiffType) < 3 {
			fmt.Println("NG")
			continue
		}
		fmt.Println(`OK`)
	}
}

// https://www.nowcoder.com/practice/6d9d69e3898f45169a441632b325c7b4?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ24 合唱队
// 描述
//N 位同学站成一排，音乐老师要请最少的同学出列，使得剩下的 K 位同学排成合唱队形。
//设KK位同学从左到右依次编号为 1，2…，K ，他们的身高分别为T_1,T_2,…,T_KT
//1
// ,T
//2
// ,…,T
//K
//  ，若存在i(1\leq i\leq K)i(1≤i≤K) 使得T_1<T_2<......<T_{i-1}<T_iT
//1
// <T
//2
// <......<T
//i−1
// <T
//i
//  且 T_i>T_{i+1}>......>T_KT
//i
// >T
//i+1
// >......>T
//K
// ，则称这KK名同学排成了合唱队形。
//通俗来说，能找到一个同学，他的两边的同学身高都依次严格降低的队形就是合唱队形。
//例子：
//123 124 125 123 121 是一个合唱队形
//123 123 124 122不是合唱队形，因为前两名同学身高相等，不符合要求
//123 122 121 122不是合唱队形，因为找不到一个同学，他的两侧同学身高递减。
//
//你的任务是，已知所有N位同学的身高，计算最少需要几位同学出列，可以使得剩下的同学排成合唱队形。
//
//注意：不允许改变队列元素的先后顺序 且 不要求最高同学左右人数必须相等
// 输入描述：
//用例两行数据，第一行是同学的总数 N ，第二行是 N 位同学的身高，以空格隔开
//输出描述：
//最少需要几位同学出列
func hj24() {
	var n, scanN int
	var heights []int
	for {
		scanN, _ = fmt.Scan(&n)
		if scanN == 0 {
			break
		}
		heights = make([]int, n)
		for i := 0; i < n; i++ {
			fmt.Scan(&heights[i])
		}

		fmt.Println(hj24DP(heights))
	}
}

// INFO: 动态规划
func hj24DP(heights []int) int {
	// 两个数组分别表示每个人左边与右边最多站的人数
	var leftMost, rightMost = make([]int, len(heights)), make([]int, len(heights))

	// 以每个人为中心求解每个人左边最多站的人数
	for center := 1; center < len(heights); center++ {
		// 根据 center 左边已经知晓的每个人的左边最多人数获得 center 左边最多人数
		for i := 0; i < center; i++ {
			if heights[center] > heights[i] && leftMost[center] < leftMost[i]+1 {
				leftMost[center] = leftMost[i] + 1
			}
		}
	}

	// 以每个人为中心求解每个人右边最多站的人数
	for center := len(heights) - 2; center >= 0; center-- {
		// 根据 center 右边已经知晓的每个人的右边最多人数获得 center 右边最多人数
		for i := len(heights) - 1; i > center; i-- {
			if heights[center] > heights[i] && rightMost[center] < rightMost[i]+1 {
				rightMost[center] = rightMost[i] + 1
			}
		}
	}

	// 获取合唱队的最多人数
	var max = 1
	for i := 0; i < len(heights); i++ {
		if max < leftMost[i]+rightMost[i]+1 {
			max = leftMost[i] + rightMost[i] + 1
		}
	}

	return len(heights) - max
}

// https://www.nowcoder.com/practice/5190a1db6f4f4ddb92fd9c365c944584?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ26 字符串排序
// 描述
//编写一个程序，将输入字符串中的字符按如下规则排序。
//规则 1 ：英文字母从 A 到 Z 排列，不区分大小写。
//如，输入： Type 输出： epTy
//规则 2 ：同一个英文字母的大小写同时存在时，按照输入顺序排列。
//如，输入： BabA 输出： aABb
//规则 3 ：非英文字母的其它字符保持原来的位置。
//如，输入： By?e 输出： Be?y
func hj26() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		chars := []rune(input.Text())
		otherChars := make([]bool, len(chars))
		letters := []rune{}
		for i, c := range chars {
			if unicode.IsLetter(c) {
				letters = append(letters, c)
				continue
			}
			otherChars[i] = true
		}
		sort.SliceStable(letters, func(i, j int) bool {
			ci, cj := letters[i], letters[j]
			return unicode.ToLower(ci) < unicode.ToLower(cj)
		})
		for i, c := range chars {
			if otherChars[i] == true {
				fmt.Printf("%c", c)
			} else {
				fmt.Printf("%c", letters[0])
				letters = letters[1:]
			}
		}
	}
}

// https://www.nowcoder.com/practice/03ba8aeeef73400ca7a37a5f3370fe68?tpId=37&tqId=21250&rp=1&ru=/exam/oj/ta&qru=/exam/oj/ta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=3&judgeStatus=undefined&tags=&title=
// HJ27 查找兄弟单词
// 输入描述：
//输入只有一行。 先输入字典中单词的个数n，再输入n个单词作为字典单词。 然后输入一个单词x 最后后输入一个整数k
//输出描述：
//第一行输出查找到x的兄弟单词的个数m 第二行输出查找到的按照字典顺序排序后的第k个兄弟单词，没有符合第k个的话则不用输出。
func hj27() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	inputs := strings.Split(input.Text(), " ")
	n, _ := strconv.Atoi(inputs[0])       // 有多少组数据
	index, _ := strconv.Atoi(inputs[n+2]) // 输出结果的第几项
	tag := inputs[n+1]                    // 以tag字符串为基准
	inputs = inputs[1 : n+1]              // 需要处理的字符串

	map1 := make(map[string]int) // 基准字符每个字母出现的次数
	res := make([]string, 0)     // 符合条件字符串
	for _, v := range tag {
		map1[string(v)]++
	}

	for _, v := range inputs { // 遍历每一项
		map2 := make(map[string]int)
		// 只有长度相同且不全等的字符串才进入循环
		if len(v) == len(tag) && v != tag {
			ta := 0
			for _, v := range v { // 当前字符生成map
				map2[string(v)]++
			}
			for _, v := range tag { // 两个map作比较
				if map1[string(v)] != map2[string(v)] {
					ta = 1
				}
			}
			if ta == 0 {
				res = append(res, v)
			}
		}
	}
	sort.Strings(res)
	fmt.Println(len(res))
	if len(res) < index {
		return
	}
	fmt.Print(res[index-1])
}

// https://www.nowcoder.com/practice/2aa32b378a024755a3f251e75cbf233a?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ29 字符串加解密
// 对输入的字符串进行加解密，并输出。
//加密方法为：
//当内容是英文字母时则用该英文字母的后一个字母替换，同时字母变换大小写,如字母a时则替换为B；字母Z时则替换为a；
//当内容是数字时则把该数字加1，如0替换1，1替换2，9替换0；
//其他字符不做变化。
//解密方法为加密的逆过程。
func hj29() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		s := input.Text()
		li := make([]byte, 0)
		for i := 0; i < len(s); i++ {
			if s[i] == 'Z' {
				li = append(li, 'a')
			} else if s[i] == 'z' {
				li = append(li, 'A')
			} else if s[i] == '9' {
				li = append(li, '0')
			} else if s[i] >= 'a' && s[i] < 'z' {
				li = append(li, s[i]-32+1)
			} else if s[i] >= 'A' && s[i] < 'Z' {
				li = append(li, s[i]+32+1)
			} else if s[i] >= '0' && s[i] < '9' {
				li = append(li, s[i]+1)
			} else {
				li = append(li, s[i])
			}
		}
		fmt.Println(string(li))

		input.Scan()
		s = input.Text()
		li = make([]byte, 0)
		for i := 0; i < len(s); i++ {
			if s[i] == 'a' {
				li = append(li, 'Z')
			} else if s[i] == 'A' {
				li = append(li, 'z')
			} else if s[i] == '0' {
				li = append(li, '9')
			} else if s[i] > 'a' && s[i] <= 'z' {
				li = append(li, s[i]-32-1)
			} else if s[i] > 'A' && s[i] <= 'Z' {
				li = append(li, s[i]+32-1)
			} else if s[i] > '0' && s[i] <= '9' {
				li = append(li, s[i]-1)
			} else {
				li = append(li, s[i])
			}

		}

		fmt.Println(string(li))
	}
}

// https://leetcode.cn/problems/permutations-ii/
// 考古碎片或者全排列问题 47. 全排列 II
// 给定一个可包含重复数字的序列 nums ，按任意顺序 返回所有不重复的全排列。
func permuteUnique(nums []int) (ans [][]int) {
	sort.Ints(nums)
	n := len(nums)
	var perm []int
	vis := make([]bool, n)
	var backtrack func(int)
	backtrack = func(idx int) {
		if idx == n {
			ans = append(ans, append([]int(nil), perm...))
			return
		}
		for i, v := range nums {
			if vis[i] || i > 0 && !vis[i-1] && v == nums[i-1] {
				continue
			}
			perm = append(perm, v)
			vis[i] = true
			backtrack(idx + 1)
			vis[i] = false
			perm = perm[:len(perm)-1]
		}
	}
	backtrack(0)
	return
}

// https://leetcode.cn/problems/nge-tou-zi-de-dian-shu-lcof/
// 剑指 Offer 60. n个骰子的点数
// 把n个骰子扔在地上，所有骰子朝上一面的点数之和为s。输入n，打印出s的所有可能的值出现的概率。
//你需要用一个浮点数数组返回答案，其中第 i 个元素代表这 n 个骰子所能掷出的点数集合中第 i 小的那个的概率。
func twoSum(n int) []float64 {
	dp := make([][]int, n+1)
	for i := 0; i <= n; i++ {
		dp[i] = make([]int, 6*n+1)
	}
	for i := 1; i <= 6; i++ {
		dp[1][i] = 1
	}

	for i := 1; i <= n; i++ {
		dp[i][i] = 1
	}

	for i := 2; i <= n; i++ {
		for j := i + 1; j <= 6*i; j++ {
			for k := 1; k <= 6; k++ {
				if j-k >= i-1 {
					dp[i][j] += dp[i-1][j-k]
				}
			}
		}
	}

	result := make([]float64, 6*n)
	for i := n; i <= 6*n; i++ {
		result[i-1] = float64(dp[n][i]) / math.Pow(6, float64(n))
	}
	return result[n-1:]
}

// https://www.nowcoder.com/practice/3cd4621963e8454594f00199f4536bb1?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ32 密码截取
// 输入描述：
//输入一个字符串（字符串的长度不超过2500）
//输出描述：
//返回有效密码串的最大长度
// 和 HJ85 最长回文子串一样
func hj32() {
	var s string
	for {
		_, err := fmt.Scan(&s)
		if err != nil {
			break
		}
		fmt.Println(longestPalindromehj32(s))
	}
}
func longestPalindromehj32(s string) int {
	dp := make([][]int, len(s))
	for i := 0; i < len(s); i++ {
		dp[i] = make([]int, len(s))
		dp[i][i] = 1
	}
	//初始化
	max := 1
	for i := len(s) - 2; i >= 0; i-- {
		for j := i + 1; j < len(s); j++ {
			if s[i] == s[j] {
				if dp[i+1][j-1] == 1 {
					dp[i][j] = 1
				} else if j-i <= 1 {
					dp[i][j] = 1
				}
				if dp[i][j] == 1 && max < j-i+1 {
					max = j - i + 1
				}
			}
		}
	}
	return max
}

// https://www.nowcoder.com/practice/66ca0e28f90c42a196afd78cc9c496ea?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ33 整数与IP地址间的转换
// 输入描述：
//输入
//1 输入IP地址
//2 输入10进制型的IP地址
//输出描述：
//输出
//1 输出转换成10进制的IP地址
//2 输出转换后的IP地址
func hj33() {
	var ip string
	fmt.Scan(&ip)
	fmt.Println(IptoIn(ip))
	var num int
	fmt.Scan(&num)
	fmt.Println(IntoIp(num))
}
func IptoIn(ip string) int {
	ipS := strings.Split(ip, ".")
	res := 0
	for i := 0; i < len(ipS); i++ {
		num, _ := strconv.Atoi(ipS[i])
		res = res*256 + num
	}
	return res
}
func IntoIp(num int) string {
	i := 1
	nums := make([]int, 0)
	var temp int
	for num > 0 {
		temp += (num % 2) * i
		num /= 2
		i *= 2
		if i == 256 {
			i = 1
			nums = append(nums, temp)
			temp = 0
		}
	}
	if temp != 0 {
		nums = append(nums, temp)
	}
	for len(nums) < 4 {
		nums = append(nums, 0)
	}
	var s string
	for _, v := range nums {
		s = fmt.Sprintf("%s.%s", strconv.Itoa(v), s)
	}
	return s[:len(s)-1]
}

// https://www.nowcoder.com/practice/e4af1fe682b54459b2a211df91a91cf3?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ36 字符串加密
// 首先，选择一个单词作为密匙，如TRAILBLAZERS。如果单词中包含有重复的字母，只保留第1个，将所得结果作为新字母表开头，并将新建立的字母表中未出现的字母按照正常字母表顺序加入新字母表
// 生成字母表
func getTable(key []byte, tableMap map[byte]byte) {
	tmpMap := make(map[byte]int)
	k := 0
	for _, ch := range key {
		if _, ok := tmpMap[ch]; !ok {
			tableMap[byte('a'+k)] = ch
			//fmt.Println(string(byte('a'+k)),string(ch))
			tmpMap[ch] = 1
			k++
		}
	}
	// 判断26个字母，没出现在temMap里面的还有哪些
	tmpMap2 := make(map[int]byte)
	k = 0
	for i := 0; i < 26; i++ {
		if _, ok := tmpMap[byte('a'+i)]; !ok {
			tmpMap2[k] = byte('a' + i)
			k++
		}
	}
	// 依次将tmpMap2 赋值给tableMap
	k = 0
	for i := 'a' + len(tmpMap); i <= 'z'; i++ {
		tableMap[byte(i)] = tmpMap2[k]
		//fmt.Println(string(i),string(tmpMap2[k]))
		k++
	}
}
func getRes(chs []byte, tableMap map[byte]byte) {
	for _, ch := range chs {
		fmt.Print(string(tableMap[ch]))
	}
	fmt.Print("\n")
}

// 输入单词，去除重复的字母，按顺序输出，26个字母没出现的再接着输出，对于a~z生成字母转换表
// 多组数据，输入仅含小写
func hj36() {
	scanner := bufio.NewScanner(os.Stdin)
	i := 0
	tableMap := make(map[byte]byte)
	for scanner.Scan() {
		i++
		str := scanner.Text()
		if i%2 == 1 { // 密钥
			getTable([]byte(str), tableMap)
		} else {
			getRes([]byte(str), tableMap)
		}
	}
}

// https://www.nowcoder.com/practice/2f6f9339d151410583459847ecc98446?tpId=37&tqId=21261&rp=1&ru=/exam/oj/ta&qru=/exam/oj/ta&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37&difficulty=3&judgeStatus=undefined&tags=&title=
// HJ38 求小球落地5次后所经历的路程和第5次反弹的高度
// 假设一个球从任意高度自由落下，每次落地后反跳回原高度的一半; 再落下, 求它在第5次落地时，共经历多少米?第5次反弹多高？
func hj38() {
	var height, all float32
	fmt.Scanf("%f", &height)
	for i := 0; i < 5; i++ {
		if i < 1 {
			all += height
		} else {
			all += height * 2
		}
		height = height / 2
	}
	fmt.Printf("%.6f \n", all)
	fmt.Printf("%.6f \n", height)
}

// https://www.nowcoder.com/practice/f9a4c19050fc477e9e27eb75f3bfd49c?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ41 称砝码
// 输入描述：
//对于每组测试数据：
//第一行：n --- 砝码的种数(范围[1,10])
//第二行：m1 m2 m3 ... mn --- 每种砝码的重量(范围[1,2000])
//第三行：x1 x2 x3 .... xn --- 每种砝码对应的数量(范围[1,10])
func hj41() {
	var n int
	fmt.Scan(&n)
	weights := make([]int, n)
	for i := 0; i < n; i++ {
		var w int
		fmt.Scan(&w)
		weights[i] = w
	}
	all := []int{} // 所有元素装到一个切片里
	for i := 0; i < n; i++ {
		var num int
		fmt.Scan(&num)
		for j := 0; j < num; j++ {
			all = append(all, weights[i])
		}
	}

	set := map[int]struct{}{0: struct{}{}} // 用map当作set用
	for _, v := range all {
		temp := []int{}
		for k, _ := range set { // map是引用类型，要先深度拷贝，不然后面的循环会循环到新增元素
			temp = append(temp, k)
		}
		for _, k := range temp {
			set[k+v] = struct{}{}
		}
	}
	fmt.Println(len(set))
}

// https://www.nowcoder.com/practice/cf24906056f4488c9ddb132f317e03bc?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ43 迷宫问题
// 定义一个二维数组 N*M ，如 5 × 5 数组下所示：
//int maze[5][5] = {
//0, 1, 0, 0, 0,
//0, 1, 1, 1, 0,
//0, 0, 0, 0, 0,
//0, 1, 1, 1, 0,
//0, 0, 0, 1, 0,
//};
//它表示一个迷宫，其中的1表示墙壁，0表示可以走的路，只能横着走或竖着走，不能斜着走，要求编程序找出从左上角到右下角的路线。入口点为[0,0],既第一格是可以走的路。
func dfs(flag [1000][1000]bool, input [][]int, ret *[][]int, i, j, m, n int) { // INFO: 深度优先搜索
	if i < 0 || j < 0 || i >= m || j >= n {
		return
	}
	if i == m-1 && j == n-1 {
		for i := range *ret {
			fmt.Printf("(%d,%d)\n", (*ret)[i][0], (*ret)[i][1])
		}
		fmt.Printf("(%d,%d)\n", m-1, n-1)
		return
	}
	if input[i][j] == 1 {
		return
	}
	if flag[i][j] {
		return
	}
	flag[i][j] = true
	*ret = append(*ret, []int{i, j})
	dfs(flag, input, ret, i-1, j, m, n)
	dfs(flag, input, ret, i+1, j, m, n)
	dfs(flag, input, ret, i, j-1, m, n)
	dfs(flag, input, ret, i, j+1, m, n)
	*ret = (*ret)[:len(*ret)-1]
	flag[i][j] = false
}
func hj43() {
	for {
		var (
			m, n int
		)
		var flag [1000][1000]bool
		_, err := fmt.Scanf("%d %d", &m, &n)
		if err == io.EOF {
			return
		}
		r := make([][]int, m)
		for i := range r {
			r[i] = make([]int, n)
		}
		for i := 0; i < m; i++ {
			for j := 0; j < n; j++ {
				fmt.Scan(&r[i][j])
			}
		}
		var ret [][]int
		dfs(flag, r, &ret, 0, 0, m, n)
	}
}

// https://www.nowcoder.com/practice/02cb8d3597cf416d9f6ae1b9ddc4fde3?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ45 名字的漂亮度
// 给出一个字符串，该字符串仅由小写字母组成，定义这个字符串的“漂亮度”是其所有字母“漂亮度”的总和。
//每个字母都有一个“漂亮度”，范围在1到26之间。没有任何两个不同字母拥有相同的“漂亮度”。字母忽略大小写。
//给出多个字符串，计算每个字符串最大可能的“漂亮度”。
func hj45() {
	n := 0
	fmt.Scan(&n)
	for i := 0; i < n; i++ {
		name := ""
		fmt.Scan(&name)
		map1 := make(map[byte]int)
		for j := 0; j < len(name); j++ {
			map1[name[j]]++
		}
		li := make([]int, 0)
		for _, v := range map1 {
			li = append(li, v)
		}
		sort.Ints(li)
		k := 26
		sum := 0
		for x := 0; x < len(li); x++ {
			num := li[len(li)-1-x]
			sum += num * k
			k--
		}
		fmt.Println(sum)
	}
}

// https://www.nowcoder.com/practice/f96cd47e812842269058d483a11ced4f?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ48 从单向链表中删除指定值的节点
func hj48() {
	n := 0
	fmt.Scan(&n)
	li := make([]int, 0)
	n1 := 0
	fmt.Scan(&n1)
	li = append(li, n1)
	for i := 0; i < n-1; i++ { // 这里 []int 顺序和链表顺序一致
		c, q := 0, 0
		fmt.Scan(&c, &q)
		l := len(li)
		for j := 0; j < len(li); j++ {
			if li[j] == q {
				if j == len(li)-1 {
					li = append(li, c)
				} else {
					li = append(li, 0)
					copy(li[j+2:], li[j+1:l])
					li[j+1] = c
				}
			}
		}
	}

	w := 0
	fmt.Scan(&w)
	for y := 0; y < len(li); y++ {
		if li[y] != w {
			fmt.Printf("%d ", li[y])
		}
	}
}

// https://www.nowcoder.com/practice/9999764a61484d819056f807d2a91f1e?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ50 四则运算
// 输入一个表达式（用字符串表示），求这个表达式的值
func hj50() {
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	data := input.Text()
	a := strings.Replace(data, "{", "(", -1)
	b := strings.Replace(a, "}", ")", -1)
	c := strings.Replace(b, "[", "(", -1)
	d := strings.Replace(c, "]", ")", -1)
	fmt.Println(cal(d))
}
func cal(s string) int {
	num := 0
	res := 0
	sign := byte('+')
	if s[0] == '-' {
		sign = byte('-')
	}
	// INFO: 使用 栈stack 来做
	var stack []int
	for i := 0; i < len(s); i++ {
		if isDigit(s[i]) {
			num = num*10 + int(s[i]-'0')
		}
		if s[i] == '(' {
			count := 1
			j := i + 1
			for count > 0 {
				if s[j] == '(' {
					count++
				} else if s[j] == ')' {
					count--
				}
				j++
			}
			num = cal(s[i+1 : j-1])
			i = j - 1
		}
		if !isDigit(s[i]) || i == len(s)-1 {
			if sign == '+' {
				stack = append(stack, num)
			} else if sign == '-' {
				stack = append(stack, -num)
			} else if sign == '*' {
				stack[len(stack)-1] *= num
			} else if sign == '/' {
				stack[len(stack)-1] /= num
			}
			sign = s[i]
			num = 0
		}
		if s[i] == ')' {
			continue
		}
	}
	for _, i := range stack {
		res += i
	}
	return res
}
func isDigit(n byte) bool {
	if n >= '0' && n <= '9' {
		return true
	} else {
		return false
	}
}

// https://www.nowcoder.com/practice/3959837097c7413a961a135d7104c314?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ52 计算字符串的编辑距离
// 字符串A: abcdefg
//字符串B: abcdef
//通过增加或是删掉字符 ”g” 的方式达到目的。这两种方案都需要一次操作。把这个操作所需要的次数定义为两个字符串的距离。
//不看答案想不出系列 动态规划 dp[i][j] = min(dp[i-1][j]+1,dp[i][j-1]+1,dp[i-1][j-1]+1 {是否加1视情况而定})
func hj52() {
	for {
		var word1, word2 string
		if _, err := fmt.Scan(&word1, &word2); err != nil {
			break
		}
		m, n := len(word1), len(word2)
		// n+1 m+1 很重要
		dp := make([][]int, m+1)
		for i := range dp {
			dp[i] = make([]int, n+1)
		}
		//边界条件初始化 很重要
		for i := 0; i < m+1; i++ {
			dp[i][0] = i // word1[i] 变成 word2[0], 删掉 word1[i], 需要 i 部操作
		}
		for j := 0; j < n+1; j++ {
			dp[0][j] = j // word1[0] 变成 word2[j], 插入 word1[j]，需要 j 部操作
		}
		// 对双重循环的理解
		for i := 1; i < m+1; i++ {
			for j := 1; j < n+1; j++ {
				if word1[i-1] == word2[j-1] {
					dp[i][j] = dp[i-1][j-1]
				} else { // Min(插入，删除，替换)
					dp[i][j] = Min(dp[i][j-1], dp[i-1][j], dp[i-1][j-1]) + 1
				}
			}
		}
		fmt.Println(dp[m][n])
	}
}
func Min(args ...int) int {
	min := args[0]
	for _, item := range args {
		if item < min {
			min = item
		}
	}
	return min
}

// https://www.nowcoder.com/practice/ba241b85371c409ea01ac0aa1a8d957b?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ55 挑7
// 输出 1到n之间 的与 7 有关数字的个数。
//一个数与7有关是指这个数是 7 的倍数，或者是包含 7 的数字（如 17 ，27 ，37 ... 70 ，71 ，72 ，73...）
func hj55() {
	bs := bufio.NewScanner(os.Stdin)
	for bs.Scan() {
		input := bs.Text()
		ret := 0
		inputInt, _ := strconv.Atoi(input)
		for i := 1; i <= inputInt; i++ {
			if i%7 == 0 {
				ret++
			} else {
				str := strconv.Itoa(i)
			loop:
				for _, stri := range str {
					switch stri {
					case '7':
						ret++
						break loop
					}
				}
			}
		}
		fmt.Println(ret)
	}
}

// https://www.nowcoder.com/practice/49e772ab08994a96980f9618892e55b6?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
//输入描述：
//输入两个字符串。保证字符串只含有'0'~'9'字符
//输出描述：
//输出求和后的结果
// HJ57 高精度整数加法
func hj57() {
	bs := bufio.NewScanner(os.Stdin)
	flag := true
	l1 := new(big.Int)
	l2 := new(big.Int)
	for bs.Scan() {
		input := bs.Text()
		if flag {
			l1.SetString(input, 10)
			flag = false
			continue
		}
		l2.SetString(input, 10)
		s := l1.Add(l1, l2)
		fmt.Println(s.String())
		flag = true
		l1 = new(big.Int)
		l2 = new(big.Int)
	}
}

// https://www.nowcoder.com/practice/e896d0f82f1246a3aa7b232ce38029d4?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ59 找出字符串中第一个只出现一次的字符
func hj59() {
	for {
		var s string
		ans := "-1"
		if _, err := fmt.Scan(&s); err != nil {
			break
		}
		str := []byte(s)
		keep := make(map[byte]int)
		for _, v := range str {
			//第一次出现
			if _, ok := keep[v]; !ok {
				keep[v] = 1
			} else {
				keep[v]++
			}
		}
		for _, v := range str {
			if value, _ := keep[v]; value == 1 {
				ans = string(v)
				break
			}
		}
		fmt.Println(ans)
	}
}

// https://www.nowcoder.com/practice/e8480ed7501640709354db1cc4ffd42a?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ63 DNA序列
func hj63() {
	var str string
	var n int
	fmt.Scan(&str, &n)
	left, right := 0, 0
	start, length := 0, math.MaxInt32
	var maxRatio float64
	c, g := 0, 0
	for {
		if str[right] == 'C' {
			c++
		}
		if str[right] == 'G' {
			g++
		}
		right++
		if right-left == n {
			break
		}
	}

	for {
		if c*g > 0 {
			ratio := float64(c+g) / float64(n)
			if ratio == 1 {
				start, length = left, right-left
				break
			}
			if ratio > maxRatio {
				start, length = left, right-left
				maxRatio = ratio
			}
		}
		if right == len(str) {
			break
		}
		if str[left] == 'C' {
			c--
		} else if str[left] == 'G' {
			g--
		}
		if str[right] == 'C' {
			c++
		} else if str[right] == 'G' {
			g++
		}
		left++
		right++
	}

	fmt.Print(str[start : start+length])
}

// https://www.nowcoder.com/practice/eaf5b886bd6645dd9cfb5406f3753e15?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ64 MP3光标位置
func hj64() {
	for {
		num := 0
		str := ""
		_, e := fmt.Scanf("%d", &num)
		fmt.Scanf("%s", &str)
		if e != nil || num == 0 {
			break
		}
		Nslice := make([]int, num)
		Tslice := []int{}
		for i := 0; i < num; i++ {
			Nslice[i] = i + 1
			if i < 4 {
				Tslice = append(Tslice, i+1)
			}
		}
		key := 0
		kt := 0
		screenLen := len(Tslice)
		for _, v := range str {
			if v == 'U' {
				if key == 0 {
					key = num - 1
					//Tslice = Nslice[num-screenLen:]
					//go的切片是按照引用传递
					copy(Tslice, Nslice[num-screenLen:])
					kt = screenLen - 1
				} else {
					key = key - 1
					if kt == 0 {
						for j := 0; j < screenLen; j++ {
							Tslice[j] = Tslice[j] - 1
						}
					} else {
						kt = kt - 1
					}
				}
			} else {
				if key == num-1 {
					key = 0
					//Tslice = Nslice[:screenLen]
					//go的切片是按照引用传递
					copy(Tslice, Nslice[:screenLen])
					kt = 0
				} else {
					key = key + 1
					if kt == screenLen-1 {
						for j := 0; j < screenLen; j++ {
							Tslice[j] = Tslice[j] + 1
						}
					} else {
						kt = kt + 1
					}
				}
			}
		}
		for _, v := range Tslice {
			fmt.Printf("%d ", v)
		}
		fmt.Println()
		fmt.Println(Nslice[key])
	}
}

// https://www.nowcoder.com/practice/181a1a71c7574266ad07f9739f791506?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ65 查找两个字符串a,b中的最长公共子串
func hj65() {
	for {
		str1, str2 := "", ""
		_, e := fmt.Scanf("%s", &str1)
		if e != nil || str1 == "" {
			break
		}
		fmt.Scanf("%s", &str2)
		if len(str1) > len(str2) {
			str1, str2 = str2, str1
		}
		fmt.Println(sonLongest(str1, str2))
	}
}
func sonLongest(str1, str2 string) string {
	L := len(str1)
	for i := 0; i < L; i++ {
		for j := 0; j <= i; j++ {
			newstr := string(str1[j : L-i+j])
			if strings.Index(str2, newstr) != -1 {
				return newstr
			}
		}
	}
	return ""
}

// https://www.nowcoder.com/practice/ca6ac6ef9538419abf6f883f7d6f6ee5?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ66 配置文件恢复
func hj66() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		s := input.Text()
		solutionhj66(s)
	}
}
func solutionhj66(s string) {
	bs := strings.Split(s, " ")
	if len(bs) == 1 {
		if strings.HasPrefix("reset", bs[0]) {
			fmt.Println("reset what")
		} else {
			fmt.Println("unknown command")
		}
	} else {
		switch {
		case strings.HasPrefix("reset", bs[0]) || strings.HasPrefix("reboot", bs[0]):
			if strings.HasPrefix("reset", bs[0]) && !strings.HasPrefix("reboot", bs[0]) {
				if strings.HasPrefix("board", bs[1]) {
					fmt.Println("board fault")
				} else {
					fmt.Println("unknown command")
				}
			} else if !strings.HasPrefix("reset", bs[0]) && strings.HasPrefix("reboot", bs[0]) {
				if strings.HasPrefix("backplane", bs[1]) {
					fmt.Println("impossible")
				} else {
					fmt.Println("unknown command")
				}
			} else {
				if strings.HasPrefix("board", bs[1]) && strings.HasPrefix("backplane", bs[1]) {
					fmt.Println("unknown command")
				} else if strings.HasPrefix("board", bs[1]) {
					fmt.Println("board fault")
				} else {
					fmt.Println("impossible")
				}
			}

		case strings.HasPrefix("board", bs[0]) || strings.HasPrefix("backplane", bs[0]):
			if strings.HasPrefix("board", bs[0]) && !strings.HasPrefix("backplane", bs[0]) {
				if strings.HasPrefix("add", bs[1]) && strings.HasPrefix("delete", bs[1]) {
					fmt.Println("unknown command")

				} else if strings.HasPrefix("add", bs[1]) {
					fmt.Println("where to add")
				} else if strings.HasPrefix("delete", bs[1]) {
					fmt.Println("no board at all")
				} else {
					fmt.Println("unknown command")
				}
			} else if !strings.HasPrefix("board", bs[0]) && strings.HasPrefix("backplane", bs[0]) {
				if strings.HasPrefix("abort", bs[1]) {
					fmt.Println("install first")
				} else {
					fmt.Println("unknown command")
				}
			} else {
				if strings.HasPrefix("add", bs[1]) && strings.HasPrefix("delete", bs[1]) ||
					(strings.HasPrefix("add", bs[1]) && strings.HasPrefix("abort", bs[1])) ||
					(strings.HasPrefix("abort", bs[1]) && strings.HasPrefix("delete", bs[1])) {
					fmt.Println("unknown command")
				} else if strings.HasPrefix("add", bs[1]) {
					fmt.Println("where to add")
				} else if strings.HasPrefix("delete", bs[1]) {
					fmt.Println("no board at all")
				} else if strings.HasPrefix("abort", bs[1]) {
					fmt.Println("install first")
				} else {
					fmt.Println("impossible")
				}
			}
		default:
			fmt.Println("unknown command")
		}
	}
}

// https://www.nowcoder.com/practice/fbc417f314f745b1978fc751a54ac8cb?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ67 24点游戏算法
// 给出4个1-10的数字，通过加减乘除运算，得到数字为24就算胜利,除法指实数除法运算,运算符仅允许出现在两个数字之间,本题对数字选取顺序无要求，但每个数字仅允许使用一次，且需考虑括号运算
//此题允许数字重复，如3 3 4 4为合法输入，此输入一共有两个3，但是每个数字只允许使用一次，则运算过程中两个3都被选取并进行对应的计算操作。
var visited []bool
var li []float64

func hj67() {
	a, b, c, d := 0.0, 0.0, 0.0, 0.0
	for true {
		if _, err := fmt.Scan(&a, &b, &c, &d); err == nil {
			li = []float64{a, b, c, d}
			visited = []bool{false, false, false, false}
			flag := counthj67(0, 0)

			fmt.Println(flag)
		} else {
			break
		}
	}
}
func counthj67(num, sum float64) bool {
	if num == 4 && sum == 24.0 {
		return true
	}
	for i := 0; i < 4; i++ {
		if !visited[i] {
			visited[i] = true
			if counthj67(num+1, sum+li[i]) || counthj67(num+1, sum/li[i]) || counthj67(num+1, sum*li[i]) || counthj67(num+1, sum-li[i]) {
				return true
			}
			visited[i] = false
		}
	}
	return false
}

// https://www.nowcoder.com/practice/ebe941260f8c4210aa8c17e99cbc663b?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ69 矩阵乘法
// 如果A是个x行y列的矩阵，B是个y行z列的矩阵，把A和B相乘，其结果将是另一个x行z列的矩阵C。这个矩阵的每个元素是由下面的公式决定的
// 输入描述：
//第一行包含一个正整数x，代表第一个矩阵的行数
//第二行包含一个正整数y，代表第一个矩阵的列数和第二个矩阵的行数
//第三行包含一个正整数z，代表第二个矩阵的列数
//之后x行，每行y个整数，代表第一个矩阵的值
//之后y行，每行z个整数，代表第二个矩阵的值
// 输出描述：
//对于每组输入数据，输出x行，每行z个整数，代表两个矩阵相乘的结果
func hj69() {
	for true {
		var a, b, c int
		if _, err := fmt.Scan(&a, &b, &c); err != nil {
			break
		}
		//初始化矩阵
		n := make([][]int, a)
		for i := 0; i < a; i++ {
			n[i] = make([]int, b, b)
		}
		m := make([][]int, b)
		for i := 0; i < b; i++ {
			m[i] = make([]int, c, c)
		}
		ans := make([][]int, a)
		for i := 0; i < a; i++ {
			ans[i] = make([]int, c, c)
		}
		for i := 0; i < a; i++ {
			for j := 0; j < b; j++ {
				fmt.Scan(&n[i][j])
			}
		}
		for i := 0; i < b; i++ {
			for j := 0; j < c; j++ {
				fmt.Scan(&m[i][j])
			}
		}
		//矩阵计算, i表示行 j表示列 k表示n所在行的列数
		//明确矩阵计算方法 行列交错相乘
		for i := 0; i < a; i++ {
			for j := 0; j < c; j++ {
				for k := 0; k < b; k++ {
					ans[i][j] += n[i][k] * m[k][j]
				}
			}
		}
		for i := 0; i < a; i++ {
			for j := 0; j < c; j++ {
				fmt.Printf("%d ", ans[i][j])
			}
			fmt.Print("\n")
		}
	}
}

// https://www.nowcoder.com/practice/15e41630514445719a942e004edc0a5b?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ70 矩阵乘法计算量估算
//矩阵乘法的运算量与矩阵乘法的顺序强相关。
//例如：
//A是一个50×10的矩阵，B是10×20的矩阵，C是20×5的矩阵
//计算A*B*C有两种顺序：((AB)C)或者(A(BC))，前者需要计算15000次乘法，后者只需要3500次。
//编写程序计算不同的计算顺序需要进行的乘法次数。
func hj70() {
	var n int
	fmt.Scan(&n)
	arr := make([][]int, n)
	for i := range arr {
		row := make([]int, 2)
		fmt.Scan(&row[0])
		fmt.Scan(&row[1])
		arr[i] = row
	}
	var str string
	fmt.Scan(&str)
	chars := []rune(str)
	var stack []rune
	count := 0
	newMatrixs := [][]int{}
	for _, c := range chars {
		if c == ')' {
			var x, y []int
			i := stack[len(stack)-2]
			if i <= 'Z' {
				x = arr[i-'A']
			} else {
				x = newMatrixs[i-'a']
			}
			j := stack[len(stack)-1]
			if j <= 'Z' {
				y = arr[j-'A']
			} else {
				y = newMatrixs[j-'a']
			}
			stack = stack[:len(stack)-3]
			// multi
			cell, s := multi(x, y)
			count += s
			newMatrixs = append(newMatrixs, cell)
			stack = append(stack, rune(len(newMatrixs)-1+'a'))
		} else {
			stack = append(stack, c)
		}
	}
	fmt.Print(count)
}
func multi(x, y []int) ([]int, int) {
	return []int{x[0], y[1]}, x[0] * y[1] * x[1]
}

// https://www.nowcoder.com/practice/43072d50a6eb44d2a6c816a283b02036?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ71 字符串通配符
// 问题描述：在计算机中，通配符一种特殊语法，广泛应用于文件搜索、数据库、正则表达式等领域。现要求各位实现字符串通配符的算法。
//要求：
//实现如下2个通配符：
//*：匹配0个或以上的字符（注：能被*和?匹配的字符仅由英文字母和数字0到9组成，下同）
//？：匹配1个字符
//注意：匹配时不区分大小写。
//考虑暴力递归过于复杂，可考虑动态规划计dp[i][j]为以s1[i],s2[j]结尾的字符串能否匹配,状态满足无后效性前提，
// 可尝试进行递推动态转移方程
// 则 dp[i][j] =
// 若 s2[j] 为给定字符类型（数字和字母） 分五种情况讨论
// 若 s1[i] = s2[j] 或者 若s1[i] = '?'，dp[i][j] = dp[i-1][j-1]
// 若s1[i] = '*' dp[i][j] = (dp[i-1][j] ||dp[i][j-1] || dp[i][j]) 若s1[i] != s2[j] s1[i]不为*或？ false
// 若s2[j] 为非给定字符类型 若s1[i] = s2[j] true 其他情况false
func hj71() {
	for true {
		var str1, str2 string
		if _, err := fmt.Scan(&str1, &str2); err != nil {
			break
		}
		// 匹配赛不分大小写 全部大写
		s1, s2 := strings.ToUpper(str1), strings.ToUpper(str2)
		length1, length2 := len(s1), len(s2)
		dp := make([][]bool, length1+1)
		for i := 0; i < length1+1; i++ {
			dp[i] = make([]bool, length2+1, length2+1)
		}
		//初始化 若s1长度为零 s2长度为大于零 为false; 若s1长度大于零 s2长度为零 只有s1全为* 才为true
		dp[0][0] = true
		n := 0
		for i := 0; i < length1; i++ {
			if s1[i] == '*' {
				n++
			} else {
				break
			}
		}
		if n == length1-1 {
			for i := 1; i < length1+1; i++ {
				dp[i][0] = true
			}
		}
		//动态转移过程
		for i := 0; i < length1; i++ {
			for j := 0; j < length2; j++ {
				//为给定字符类型
				if match(s2[j]) {
					if s1[i] == s2[j] || s1[i] == '?' {
						dp[i+1][j+1] = dp[i][j]
					} else if s1[i] == '*' {
						dp[i+1][j+1] = (dp[i][j+1] || dp[i+1][j] || dp[i][j])
					}
					if s1[i] != s2[j] && s1[i] != '?' && s1[i] != '*' {
						dp[i+1][j+1] = false
					}
				} else { //非给定类型
					if s1[i] == s2[j] {
						dp[i+1][j+1] = dp[i][j]
					} else {
						dp[i+1][j+1] = false
					}
				}
			}
		}
		fmt.Println(dp[length1][length2])
	}
}

//判断是否为给定字符类型
func match(r byte) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	return false
}

// https://www.nowcoder.com/practice/668603dc307e4ef4bb07bcd0615ea677?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ74 参数解析

// https://www.nowcoder.com/practice/98dc82c094e043ccb7e0570e5342dd1b?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// 给定两个只包含小写字母的字符串，计算两个字符串的最大公共子串的长度。
func MaxLength(str1, str2 string) int {
	m, n := len(str1), len(str2)

	//方法一：初始化一个二维切片
	//var dp [][]int
	//for i := 0; i < m+1; i++ {
	//  ar := make([]int, n+1)
	//  dp = append(dp, ar)
	//}

	//方法二：初始化一个二维切片
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	var maxLen int
	for i := 0; i <= m; i++ {
		dp[i][0] = 0
	}
	for i := 0; i <= n; i++ {
		dp[0][i] = 0
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if str1[i-1] == str2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = 0
			}
			if dp[i][j] > maxLen {
				maxLen = dp[i][j]
			}
		}
	}
	return maxLen
}
func hj75() {
	var s1, s2 string
	fmt.Scanln(&s1)
	fmt.Scanln(&s2)
	fmt.Println(MaxLength(s1, s2))
}

// https://www.nowcoder.com/practice/97ba57c35e9f4749826dc3befaeae109?tpId=37&tags=&title=&difficulty=3&judgeStatus=0&rp=1&sourceUrl=%2Fexam%2Foj%2Fta%3Fdifficulty%3D3%26page%3D1%26pageSize%3D50%26search%3D%26tpId%3D37%26type%3D37
// HJ77 火车进站
