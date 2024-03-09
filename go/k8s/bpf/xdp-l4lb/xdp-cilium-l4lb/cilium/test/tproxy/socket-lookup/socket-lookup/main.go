package main

import (
    "fmt"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "os"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go dispatcher test_sk_lookup.c -- -I.

// go generate .
// CGO_ENABLED=0 go run .

/**
redirect: tcp://127.0.0.1:8080 > tcp://127.0.0.1:80

1. sk-lookup bind foo tcp 127.0.0.1 8080
2. sk-lookup register-pid 12345 foo tcp 127.0.0.1 80
*/

func init() {
    logrus.SetReportCaller(true)
}

var rootCmd = &cobra.Command{
    Use:  "sk-lookup",
    Long: "sk-lookup for lookup listening(TCP)/unconnected(UDP) socket",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func main() {
    Execute()
}
