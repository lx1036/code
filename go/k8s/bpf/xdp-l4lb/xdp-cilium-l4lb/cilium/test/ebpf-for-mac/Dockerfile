


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
RUN git clone https://github.com/iovisor/bcc.git && \
    mkdir bcc/build && \
    cd bcc/build && \
    cmake .. && \
    make && \
    make install && \
    cmake -DPYTHON_CMD=python3 .. && \
    cd src/python/ && \
    make && \
    make install && \
    sed -i "s/self._syscall_prefixes\[0\]/self._syscall_prefixes\[1\]/g" /usr/lib/python3/dist-packages/bcc/__init__.py

CMD mount -t debugfs debugfs /sys/kernel/debug && /bin/bash


#docker build -t ebpf-for-mac -f ./Dockerfile .
# 或者直接 pull 已经 build 好的
# docker pull lx1036/ebpf-for-mac
#
