



#!/bin/bash


apt-get update -y
apt-get install -y gcc-multilib libbpf-dev clang linux-tools-`uname -r` jq

# 这里安装 libbpf-dev 包后，代码里可以直接 include linux 头文件
clang -O2 -Wall -target bpf -c tcprtt_sockops.c -o tcprtt_sockops.o

