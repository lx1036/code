package demo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestArgs(test *testing.T) {
	fmt.Println(os.Args)
	binaryName := filepath.Base(os.Args[0])
	fmt.Println(binaryName)
}
