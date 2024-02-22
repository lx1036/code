package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_sockopt.c -- -I.

// go generate .
// CGO_ENABLED=0 go run .
func main() {

}
