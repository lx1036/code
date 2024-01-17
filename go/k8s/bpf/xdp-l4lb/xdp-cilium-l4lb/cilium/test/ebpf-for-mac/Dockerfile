


# 基于最新的 5.15.49，测试没有问题
FROM docker/for-desktop-kernel:5.15.49-13422a825f833d125942948cf8a8688cef721ead AS ksrc

FROM ubuntu:22.04 AS bpftrace
COPY --from=ksrc /kernel-dev.tar /
RUN tar xf kernel-dev.tar && rm kernel-dev.tar
# Use Alibaba Cloud mirror for ubuntu
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list
# Install LLVM 17.0.1
#RUN apt-get update -y && apt install -y wget lsb-release software-properties-common && wget https://apt.llvm.org/llvm.sh && chmod +x llvm.sh && ./llvm.sh 17
RUN apt-get update -y && apt install -y wget lsb-release software-properties-common llvm
#ENV PATH "$PATH:/usr/lib/llvm/bin"

# Build/Install bpftrace
RUN apt-get install -y bpftrace net-tools iproute2

# Build/Install bcc
WORKDIR /root
RUN DEBIAN_FRONTEND="noninteractive" apt install -y kmod vim bison build-essential cmake flex git libedit-dev \
  libcap-dev zlib1g-dev libelf-dev libfl-dev python3 python3-pip python3-dev clang libclang-dev && \
  ln -s $(which python3) /usr/bin/python

#CMD mount -t debugfs debugfs /sys/kernel/debug && /bin/bash

# golang
RUN apt update -y && apt install -y net-tools iproute2 iptables \
    ipset curl wget sysstat git tcpdump \
    inetutils-ping vim ethtool jq

RUN cd /tmp && wget https://go.dev/dl/go1.20.5.linux-amd64.tar.gz && tar zxf go1.20.5.linux-amd64.tar.gz && mv go /usr/local/
#RUN cp /usr/local/go/bin/go /usr/bin/go
ENV GOROOT /usr/local/go
ENV GOPATH /root/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH
RUN go env -w GOPROXY="https://goproxy.cn,https://proxy.golang.org,direct"
RUN go version
RUN mkdir -p /root/go
RUN go install github.com/axw/gocov/gocov@latest && \
    go install github.com/AlekSi/gocov-xml@latest && \
    go install gotest.tools/gotestsum@latest && \
    pip3 install diff_cover

# ebpf
# RUN apt install -y llvm linux-tools-`uname -r` # linux 机器上编译用这个, mac 上编译用 linux-tools-generic
RUN apt-get update -y && apt-get install -y gcc-multilib libbpf-dev clang \
    linux-tools-`uname -r` jq bpfcc-tools iputils-ping

# docker build -t lx1036/ebpf-for-mac:2.2 -f ./Dockerfile .
# 或者直接 pull 已经 build 好的
# docker pull lx1036/ebpf-for-mac:2.2

# mac 本地运行，需要 volume debugfs
# docker volume create --driver local --opt type=debugfs --opt device=debugfs debugfs
# docker run -it --name ebpf-for-mac2 --privileged -v debugfs:/sys/kernel/debug:ro -v /lib/modules:/lib/modules:ro -v /etc/localtime:/etc/localtime:ro --pid=host -v /Users/liuxiang/Code/code:/mnt/code -v /Users/liuxiang/go/pkg/mod:/root/go/pkg/mod lx1036/ebpf-for-mac:2.1 /bin/bas


# !!! 安装 bpftool 貌似不行，linux-tools-`uname -r` bpfcc-tools 包里都没有，很奇怪
