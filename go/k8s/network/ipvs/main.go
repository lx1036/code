package main

import (
	"fmt"
	"github.com/moby/ipvs"
	"log"
)

func main() {
	handle, err := ipvs.New("")
	if err != nil {
		log.Fatalf("ipvs.New: %s", err)
	}
	svcs, err := handle.GetServices()
	if err != nil {
		log.Fatalf("handle.GetServices: %s", err)
	}

	for _, svc := range svcs {
		fmt.Println(svc)
	}
}
