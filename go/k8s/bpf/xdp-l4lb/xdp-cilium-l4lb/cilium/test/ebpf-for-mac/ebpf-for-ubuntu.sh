
# ubuntu 22.04，内核版本 5.15 > 5.7

# Use Alibaba Cloud mirror for ubuntu
sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list
apt update -y
apt install -y wget lsb-release software-properties-common \
bpftrace net-tools iproute2 kmod vim bison build-essential cmake flex git libedit-dev \
libcap-dev zlib1g-dev libelf-dev libfl-dev python3.11 python3-pip python3.11-dev clang libclang-dev \
net-tools iproute2 iptables ipset curl wget sysstat git tcpdump vim ethtool jq \
gcc-multilib libbpf-dev clang linux-tools-`uname -r` jq bpfcc-tools iputils-ping llvm


