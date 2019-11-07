package main

import (
	"fmt"
	"github.com/codingsince1985/checksum"
)

func main() {
	file := "/Users/lx1036/Downloads/QQ_V6.5.5.dmg"
	md5, _ := checksum.MD5sum(file)
	fmt.Println(md5)
	sha256, _ := checksum.SHA256sum(file)
	fmt.Println(sha256)
}
