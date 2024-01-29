#!/bin/bash

bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h
clang -O2 -Wall -target bpf -c tc_nodeport.c -I. -o tc_nodeport.o


tc qdisc add dev eth-svc clsact
tc filter add dev eth-svc ingress bpf da obj tc_nodeport.o sec tc_svc_ingress
tc filter show dev eth-svc ingress
