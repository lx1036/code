package main

import (
	"fmt"
	"time"
)

const (
	debugLevel uint32 = iota
	infoLevel
)

func main() {
	fmt.Println(debugLevel, infoLevel,time.Now().Unix())
}
