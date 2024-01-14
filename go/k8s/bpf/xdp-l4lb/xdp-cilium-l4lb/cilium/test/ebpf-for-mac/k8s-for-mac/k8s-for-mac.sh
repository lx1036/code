#!/usr/bin/env bash

# 研究 flannel ipip/vxlan for ebpf
minikube start --cni=flannel --driver=docker --image-mirror-country=cn --image-repository="registry.cn-hangzhou.aliyuncs.com/google_containers" --kubernetes-version=v1.28.3

# minukube node 里安装 tcpdump
sudo apt update -y && sudo apt install -y tcpdump




