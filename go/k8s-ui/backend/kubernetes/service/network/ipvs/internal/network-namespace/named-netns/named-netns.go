package main

import (
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"runtime"
	"syscall"
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
		os.Exit(0)
	}

	if err := ExecNetworkNamespace(args[0], args[1:]); err != nil {
		panic(err)
	}
}

func ExecNetworkNamespace(name string, args []string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()


}

//
func CreateNetworkNamespace(name string)  {
	if err := syscall.Unshare(syscall.CLONE_NEWNET); err != nil {
		return -1, err
	}

}

func SetNs(file *os.File) error {
	_, _, err := syscall.Syscall(unix.SYS_SETNS, file.Fd(), syscall.CLONE_NEWNET, 0)
	if err != 0 {
		return err
	}

	return nil
}
