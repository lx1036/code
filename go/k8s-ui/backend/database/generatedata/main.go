package main

import (
	"fmt"
	"k8s-lx1036/k8s-ui/backend/database/initial"
)

func main() {
	for _, data := range initial.InitialData {
		fmt.Println(data)
	}
}
