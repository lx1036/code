package api

import (
	"crypto/rand"
	"fmt"
)

func GenerateCsrfKey() string {
	bytes := make([]byte, 256)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("can't generate csrf key because of %s", err.Error()))
	}

	return string(bytes)
}
