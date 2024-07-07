package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"k8s-lx1036/k8s/scheduler/volcano/volcano/cmd/scheduler/app/options"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/cmd/scheduler/app"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"k8s.io/apimachinery/pkg/util/wait"
	cliflag "k8s.io/component-base/cli/flag"
)

var logFlushFreq = pflag.Duration("log-flush-frequency", 5*time.Second, "Maximum number of seconds between log flushes")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	klog.InitFlags(nil)

	s := options.NewServerOption()
	s.AddFlags(pflag.CommandLine)
	s.RegisterOptions()
	if err := s.CheckOptionOrDie(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if s.CaCertFile != "" && s.CertFile != "" && s.KeyFile != "" {
		if err := s.ParseCAFiles(nil); err != nil {
			klog.Fatalf("Failed to parse CA file: %v", err)
		}
	}

	// print flags
	cliflag.InitFlags()

	go wait.Until(klog.Flush, *logFlushFreq, wait.NeverStop)
	defer klog.Flush()

	if err := app.Run(s); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
