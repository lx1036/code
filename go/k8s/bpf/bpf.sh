

# Ubuntu

# 安装 bpf 相关依赖

# linux-tools-$(uname -r) linux headers 安装在目录 /usr/include/linux
apt update -y
apt install -y clang llvm libelf-dev libpcap-dev build-essential linux-tools-$(uname -r) linux-tools-common linux-tools-generic

# https://docs.cilium.io/en/stable/bpf/toolchain/#development-environment
sudo apt-get install -y make gcc libssl-dev bc libelf-dev libcap-dev \
  clang gcc-multilib llvm libncurses5-dev git pkg-config libmnl-dev bison flex \
  graphviz
