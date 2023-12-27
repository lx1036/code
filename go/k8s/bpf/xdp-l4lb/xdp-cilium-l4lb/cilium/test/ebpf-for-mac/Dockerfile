


# 基于最新的 5.15.49，测试没有问题
FROM docker/for-desktop-kernel:5.15.49-13422a825f833d125942948cf8a8688cef721ead AS ksrc

FROM ubuntu:20.04 AS bpftrace
COPY --from=ksrc /kernel-dev.tar /
RUN tar xf kernel-dev.tar && rm kernel-dev.tar
# Use Alibaba Cloud mirror for ubuntu
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list
# Install LLVM 10.0.1
RUN apt-get update && apt install -y wget lsb-release software-properties-common && wget https://apt.llvm.org/llvm.sh && chmod +x llvm.sh && ./llvm.sh 10
ENV PATH "$PATH:/usr/lib/llvm-10/bin"

# Build/Install bpftrace
RUN apt-get install -y bpftrace net-tools iproute2

# Build/Install bcc
WORKDIR /root
RUN DEBIAN_FRONTEND="noninteractive" apt install -y kmod vim bison build-essential cmake flex git libedit-dev \
  libcap-dev zlib1g-dev libelf-dev libfl-dev python3.8 python3-pip python3.8-dev clang libclang-dev && \
  ln -s $(which python3) /usr/bin/python

#CMD mount -t debugfs debugfs /sys/kernel/debug && /bin/bash

# golang
RUN apt update -y && apt install -y net-tools iproute2 iptables \
    ipset curl wget sysstat git tcpdump \
    inetutils-ping vim ethtool jq

RUN cd /tmp && wget https://go.dev/dl/go1.20.5.linux-amd64.tar.gz && tar zxf go1.20.5.linux-amd64.tar.gz && mv go /usr/local/
#RUN cp /usr/local/go/bin/go /usr/bin/go
ENV GOROOT /usr/local/go
ENV PATH $GOROOT/bin:$PATH
ENV GOPATH /root/go
RUN go env -w GOPROXY="https://goproxy.cn,https://proxy.golang.org,direct"
RUN go version
RUN mkdir -p /root/go
RUN go install github.com/axw/gocov/gocov@latest && \
    go install github.com/AlekSi/gocov-xml@latest && \
    go install gotest.tools/gotestsum@latest && \
    pip3 install diff_cover

# ebpf
RUN apt-get update -y && apt-get install -y gcc-multilib libbpf-dev clang \
    linux-tools-`uname -r` jq bpfcc-tools iputils-ping

# docker build -t lx1036/ebpf-for-mac:2.0 -f ./Dockerfile .
# 或者直接 pull 已经 build 好的
# docker pull lx1036/ebpf-for-mac:2.0

# mac 本地运行，需要 volume debugfs
# docker volume create --driver local --opt type=debugfs --opt device=debugfs debugfs
# docker run -it --name ebpf-for-mac2 --privileged -v debugfs:/sys/kernel/debug:ro -v /lib/modules:/lib/modules:ro -v /etc/localtime:/etc/localtime:ro --pid=host -v /Users/liuxiang/Code/code:/mnt/code -v /Users/liuxiang/go/pkg/mod:/root/go/pkg/mod lx1036/ebpf-for-mac:2.1 /bin/bas

