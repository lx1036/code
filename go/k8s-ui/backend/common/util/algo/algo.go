package algo

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
)

func Sha256(source string) string {
	algo := sha256.New()
	_, _ = io.WriteString(algo, fmt.Sprintf("%s", source))
	sign := fmt.Sprintf("%x", algo.Sum(nil))

	return sign
}

func Md5(source string) string {
	algo := md5.New()
	_, _ = io.WriteString(algo, fmt.Sprintf("%s", source))
	sign := fmt.Sprintf("%x", algo.Sum(nil))

	return sign
}

var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyz")

func Random(number int) string {
	b := make([]rune, number)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
