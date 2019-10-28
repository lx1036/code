package util

import (
	"bufio"
	"fmt"
	"os"
)

func Prompt()  {
	fmt.Println("Press Return key to continue...")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
