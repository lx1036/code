#!/bin/bash

# /root/linux-5.10.142/tools/testing/selftests/bpf/test_xdping.sh

# xdping tests
#   Here we setup and teardown configuration required to run
#   xdping, exercising its options.
#
#   Setup is similar to test_tunnel tests but without the tunnel.
#
# Topology:
# ---------
#     root namespace   |     tc_ns0 namespace
#                      |
#      ----------      |     ----------
#|  veth1(10.1.1.200)  | --------- |  veth0(10.1.1.100)  |
#      ----------    peer    ----------
#
# Device Configuration
# --------------------
# Root namespace with BPF
# Device names and addresses:
#	veth1 IP: 10.1.1.200
#	xdp added to veth1, xdpings originate from here.
#
# Namespace tc_ns0 with BPF
# Device names and addresses:
#       veth0 IPv4: 10.1.1.100
#	For some tests xdping run in server mode here.


readonly TARGET_IP="10.1.1.100"
readonly TARGET_NS="xdp_ns0"
readonly LOCAL_IP="10.1.1.200"

setup()
{
	ip netns add $TARGET_NS
	ip link add veth0 type veth peer name veth1
	ip link set veth0 netns $TARGET_NS
	ip netns exec $TARGET_NS ip addr add ${TARGET_IP}/24 dev veth0
	ip addr add ${LOCAL_IP}/24 dev veth1
	ip netns exec $TARGET_NS ip link set veth0 up
	ip link set veth1 up
}

cleanup()
{
	set +e
	ip netns delete $TARGET_NS 2>/dev/null
	ip link del veth1 2>/dev/null
}

set -e

setup

exit 0

# 报错，这个 .o 文件不能直接使用 ip link 挂载
ip netns exec xdp_ns0 ip link set veth0 xdp object bpf_bpfel.o section xdp/server

# 直接使用 ip link 命令来挂载 xdp 程序
ip netns exec xdp_ns0 ip link set veth0 xdp off
clang -O2 -Wall -target bpf -c xdp_ping.c -o xdp_ping.o
# clang -S -I. -O2 -emit-llvm -c xdp_ping.c -o - | llc -march=bpf -filetype=obj -o xdp_ping2.o
ip netns exec xdp_ns0 ip link set dev veth0 xdp object xdp_ping.o section xdp_server # 容器里可以挂载
ip link set veth1 xdp off
ip link set dev veth1 xdp object xdp_ping.o section xdp_client # 但是容器里不能挂载会报错

ping -c 3 -I veth1 10.1.1.100 # 可达
