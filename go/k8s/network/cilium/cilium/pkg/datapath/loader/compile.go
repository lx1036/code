package loader

import (
	"context"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/config/option"
)

// Compile compiles a BPF program generating an object file.
func Compile(ctx context.Context, src string, out string) error {
	debug := option.Config.BPFCompilationDebug
	prog := progInfo{
		Source:     src,
		Output:     out,
		OutputType: outputObject,
	}
	dirs := directoryInfo{
		Library: option.Config.BpfDir,
		Runtime: option.Config.StateDir,
		Output:  option.Config.StateDir,
		State:   option.Config.StateDir,
	}
	return compile(ctx, &prog, &dirs, debug)
}
