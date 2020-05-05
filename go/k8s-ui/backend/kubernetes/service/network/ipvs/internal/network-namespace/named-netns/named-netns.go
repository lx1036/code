package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "command[%s] in named network namespace:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	if err := ExecNetworkNamespace(args[0], args[1:]); err != nil {
		panic(err)
	}
}

func ExecNetworkNamespace(name string, args []string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()


}

func CreateNetworkNamespace()  {
	
}
