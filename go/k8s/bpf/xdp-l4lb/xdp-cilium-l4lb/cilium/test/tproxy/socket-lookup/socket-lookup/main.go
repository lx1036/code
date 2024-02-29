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

var (
    logr   *logrus.Logger
    logger *logrus.Entry
)

func init() {
    logrus.SetReportCaller(true)
    logr = logrus.New()
    //logger = logr.WithField("id", )
    logr.SetFormatter(&logrus.JSONFormatter{})
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
