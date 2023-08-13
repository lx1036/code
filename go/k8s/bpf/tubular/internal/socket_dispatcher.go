
package internal


//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc "$CLANG" -strip "$STRIP" -makebase "$MAKEDIR" dispatcher ../ebpf/socket_dispatch.c -- -mcpu=v2 -nostdinc -Wall -Werror -I../ebpf/include


type Dispatcher struct {
}