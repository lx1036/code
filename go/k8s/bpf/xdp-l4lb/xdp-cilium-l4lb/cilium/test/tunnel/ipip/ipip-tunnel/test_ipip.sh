#!/bin/bash

# /root/linux-5.10.142/tools/testing/selftests/bpf/test_tunnel.sh

# clang -O2 -Wall -target bpf -c test_tunnel_kern.c -o test_tunnel_kern.o

cleanup()
{
	ip netns delete ipip_ns0 2> /dev/null
	ip netns delete ipip_ns1 2> /dev/null
#	ip link del ipip_tunnel1 2> /dev/null
}

config_device()
{
	ip netns add ipip_ns0
	ip link add ipip_veth0 type veth peer name ipip_veth1
	ip link set ipip_veth0 netns ipip_ns0
	ip netns exec ipip_ns0 ip link set dev ipip_veth0 up
	ip netns exec ipip_ns0 ip addr add 173.16.1.100/24 dev ipip_veth0
	ip link set dev ipip_veth1 up mtu 1500
	ip addr add dev ipip_veth1 173.16.1.200/24

	ip netns exec ipip_ns0 ip link add dev ipip_tunnel0 type ipip local 173.16.1.100 remote 173.16.1.200 # 注意这个ip
	ip netns exec ipip_ns0 ip link set dev ipip_tunnel0 up
  ip netns exec ipip_ns0 ip addr add 10.1.1.100/24 dev ipip_tunnel0
  # root namespace 只能有一个 ipip external 网卡
  ip link add dev ipip_tunnel1 type ipip external
  ip link set dev ipip_tunnel1 up
  ip addr add dev ipip_tunnel1 10.1.1.200/24

  tc qdisc add dev ipip_tunnel1 clsact
  tc filter add dev ipip_tunnel1 egress bpf da obj test_tunnel_kern.o sec ipip_set_tunnel
  tc filter add dev ipip_tunnel1 ingress bpf da obj test_tunnel_kern.o sec ipip_get_tunnel

  ping -c 1 10.1.1.100
  ip netns exec ipip_ns0 ping -c 1 10.1.1.200
}

config_device2()
{
  ip netns add ipip_ns0
  ip netns add ipip_ns1
  ip link add ipip_veth0 type veth peer name ipip_veth1
  ip link set ipip_veth0 netns ipip_ns0
  ip link set ipip_veth1 netns ipip_ns1

  ip netns exec ipip_ns0 ip link set dev ipip_veth0 up
  ip netns exec ipip_ns1 ip link set dev ipip_veth1 up
  ip netns exec ipip_ns0 ip addr add 173.16.1.200/24 dev ipip_veth0
  ip netns exec ipip_ns0 ip link set dev ipip_veth0 mtu 1500
  ip netns exec ipip_ns1 ip addr add 173.16.1.100/24 dev ipip_veth1

  ip netns exec ipip_ns1 ip link add dev ipip_tunnel1 type ipip local 173.16.1.100 remote 173.16.1.200 # 注意这个ip
  ip netns exec ipip_ns1 ip link set dev ipip_tunnel1 up
  ip netns exec ipip_ns1 ip addr add 10.1.1.100/24 dev ipip_tunnel1

  # ipip_ns0 namespace 只能有一个 ipip external 网卡
  ip netns exec ipip_ns0 ip link add dev ipip_tunnel0 type ipip external
  ip netns exec ipip_ns0 ip link set dev ipip_tunnel0 up
  ip netns exec ipip_ns0 ip addr add dev ipip_tunnel0 10.1.1.200/24

  ip netns exec ipip_ns0 tc qdisc add dev ipip_tunnel0 clsact
  ip netns exec ipip_ns0 tc filter add dev ipip_tunnel0 egress bpf da obj test_tunnel_kern.o sec ipip_set_tunnel
  ip netns exec ipip_ns0 tc filter add dev ipip_tunnel0 ingress bpf da obj test_tunnel_kern.o sec ipip_get_tunnel

  ip netns exec ipip_ns0 ping -c 1 10.1.1.100
  ip netns exec ipip_ns1 ping -c 1 10.1.1.200
}

# 验证通过
cleanup
config_device2
