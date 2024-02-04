#!/bin/bash


apt-get update -y
apt-get install -y gcc-multilib libbpf-dev clang linux-tools-`uname -r` jq

# 这里安装 libbpf-dev 包后，代码里可以直接 include linux 头文件
clang -O2 -Wall -target bpf -c acl.c -o acl.o


ip link add dev eth0-acl type dummy
ip link set dev eth0-acl up
ip addr add 10.20.30.40/24 dev eth0-acl
ip addr add 10.20.30.41/24 dev eth0-acl
