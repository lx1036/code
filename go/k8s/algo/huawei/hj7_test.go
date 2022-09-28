package huawei

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
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
