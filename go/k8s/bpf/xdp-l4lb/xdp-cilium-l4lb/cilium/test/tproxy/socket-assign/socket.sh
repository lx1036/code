#!/bin/bash

clang -O2 -Wall -target bpf -c test_sk_assign.c -o test_sk_assign.o
tc qdisc add dev lo clsact
tc filter add dev lo ingress bpf da obj test_sk_assign.o sec classifier/sk_assign_test

tc qdisc del dev lo clsact
rm -rf /sys/fs/bpf/tc/globals/server_map
