



#!/bin/bash


apt-get update -y
apt-get install -y gcc-multilib libbpf-dev clang linux-tools-generic jq

# 这里安装 libbpf-dev 包后，代码里可以直接 include linux 头文件
clang -O2 -Wall -target bpf -c tc_l2_redirect_kern.c -o tc_l2_redirect_kern.o

