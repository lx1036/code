package api

import (
	"fmt"
	"golang.org/x/net/xsrftoken"
	"testing"
)

func TestGenerateCsrfKey(test *testing.T) {
	key := GenerateCsrfKey()
	token := xsrftoken.Generate(key, "none", "login")
	fmt.Println(token) // 7nZhW03OLGhQcrho4vvd9jWcTwU:1584157893346
}
