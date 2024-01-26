#!/usr/bin/env bash

# install minikube on linux
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube


# 研究 flannel ipip/vxlan for ebpf
minikube start --cni=flannel --driver=docker --image-mirror-country=cn --image-repository="registry.cn-hangzhou.aliyuncs.com/google_containers" --kubernetes-version=v1.28.3
minikube start --cni=cilium --driver=docker --image-mirror-country=cn --image-repository="registry.cn-hangzhou.aliyuncs.com/google_containers" --kubernetes-version=v1.28.3
# ecs
minikube start --cni=cilium --driver=docker --kubernetes-version=v1.28.3 --force --listen-address=0.0.0.0
kubectl create deploy my-nginx --image=nginx:1.24.0 --replicas=3

# minukube node 里安装 tcpdump
sudo apt update -y && sudo apt install -y tcpdump




