package string

import (
	"fmt"
	"strconv"
	"testing"
)

func TestReverse(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello, world", "dlrow ,olleH"},
		{"Hello, 世界", "界世 ,olleH"},
		{"", ""},
	}
	for _, c := range cases {
		got := Reverse(c.in)
		if got != c.want {
			t.Errorf("Reverse(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSubstring(t *testing.T) {
	query := "我是胡八一"
	//query := "hello"
	runeQuery := []rune(query)
	length := len(runeQuery)
	fmt.Println(length)
	fmt.Println(string(runeQuery[0:4]) + strconv.Itoa(length) + string(runeQuery[length-4:]))
}
