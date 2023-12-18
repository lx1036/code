package tcp

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestHexDecode(test *testing.T) {
	b, err := hex.DecodeString("0xde")
	if err != nil {
		test.Fatal(err) // invalid byte: U+0078 'x'
	}
	u := uint8(b[0])
	fmt.Println(u)
}

func TestUint8(test *testing.T) {
	i := 0xde
	u := uint8(i)
	fmt.Println(u)
}
