

#!/bin/bash

# /root/linux-5.10.142/samples/bpf/tc_l2_redirect.sh

# rp_filter
# rp_filter 是Linux内核中的一个参数，用于控制网络包的接收策略。它主要用于防止IP欺骗攻击，
# 通过检查接收到的数据包是否来自正确的源地址来决定是否接受该数据包
# 0：关闭反向路径过滤功能。
# 1：只检查本地子网内的数据包。如果数据包的目的地址不在本地子网内，则直接丢弃该数据包。
# 2：检查所有的数据包。如果数据包的目的地址不在本地子网内，并且没有有效的路由可以到达该地址，则直接丢弃该数据包。

# forwarding
# forwarding 是Linux内核中的一个参数，用于控制网络包的转发行为。
# 当forwarding被启用时，Linux系统将会允许在网络接口之间转发数据包，从而实现路由器的功能。
# 0：关闭数据包转发功能。
# 1：启用数据包转发功能。

# 默认值
RP_FILTER=$(< /proc/sys/net/ipv4/conf/all/rp_filter)
FORWARDING=$(< /proc/sys/net/ipv4/conf/all/forwarding)

REDIRECT_USER='./tc_l2_redirect'
REDIRECT_BPF='./tc_l2_redirect_kern.o'

VIP='10.10.1.102'

function cleanup {
  set +e
    
  ip netns delete ns1 >& /dev/null
	ip netns delete ns2 >& /dev/null
	ip link del veth1 >& /dev/null
	ip link del veth2 >& /dev/null
	ip link del tun1 >& /dev/null
  # rp_filter 和 forwarding
  sysctl -q -w net.ipv4.conf.all.rp_filter=$RP_FILTER
	rm -f /sys/fs/bpf/tc/globals/tun_iface

  set -e
}

function config_common {
  ip netns add ns1
  ip netns add ns2
  ip link add veth1 type veth peer name eth0 netns ns1
  ip link add veth2 type veth peer name eth0 netns ns2
  ip link set dev veth1 up
  ip link set dev veth2 up
  ip link set dev veth1 mtu 1500
  ip link set dev veth2 mtu 1500

  # ns1
  ip netns exec ns1 ip link set dev lo up
  ip netns exec ns1 ip link set dev eth0 up
  ip netns exec ns1 ip addr add 10.1.1.101/24 dev eth0
  ip netns exec ns1 ip route add default via 10.1.1.1 dev eth0

  # ns2
  ip netns exec ns2 ip link set dev lo up
	ip netns exec ns2 ip link set dev eth0 up
	ip netns exec ns2 ip addr add 10.1.2.101/24 dev eth0
	ip netns exec ns2 ip addr add $VIP dev lo
  ip netns exec ns2 ip link add tun2 type ipip local 10.1.2.101 remote 10.1.2.1 # 注意这里的 10.1.2.1 网关地址, 是 veth2 网卡地址
  ip netns exec ns2 ip link set dev tun2 up

	# add route for return path，回包时走该路由
  ip netns exec ns2 ip route add 10.1.1.0/24 dev tun2
  ip netns exec ns2 sysctl -q -w net.ipv4.conf.all.rp_filter=0
  ip netns exec ns2 sysctl -q -w net.ipv4.conf.tun2.rp_filter=0 # ipip/1 为 ipip.1, `sysctl net.ipv4.conf.ipip/1.rp_filter`
  ip addr add 10.1.1.1/24 dev veth1
  ip addr add 10.1.2.1/24 dev veth2

  sysctl -q -w net.ipv4.conf.all.rp_filter=0
}

function loadTCBpf() {
  # 加载 tc ingress bpf 程序
  ip netns exec ns2 tc qdisc add dev eth0 clsact
  ip netns exec ns2 tc filter add dev eth0 ingress bpf da obj tc_l2_redirect_kern.o sec drop_non_tun_vip
  ip netns exec ns2 tc filter show dev eth0 ingress
  # 加载 tc ingress bpf 程序
  tc qdisc add dev veth2 clsact
	tc filter add dev veth2 ingress bpf da obj tc_l2_redirect_kern.o sec l2_to_iptun_ingress_forward
	tc filter show dev veth2 ingress
}

function l2_to_ipip {
  echo -n "l2_to_ipip $1: "

	local dir=$1

	# tun1 网卡的作用是啥?
  ip link add tun1 type ipip external
  ip link set dev tun1 up
  sysctl -q -w net.ipv4.conf.tun1.rp_filter=0 # 貌似没起作用，每次运行都是 2，需要手动设置 0
  sysctl -q -w net.ipv4.conf.tun1.forwarding=1

  if [[ $dir == "egress" ]]; then
    ip route add 10.10.1.0/24 via 10.1.2.1 dev veth2
    tc filter add dev veth2 egress bpf da obj tc_l2_redirect_kern.o sec l2_to_iptun_ingress_redirect
    sysctl -q -w net.ipv4.conf.veth1.forwarding=1
	else
    tc qdisc add dev veth1 clsact
    tc filter add dev veth1 ingress bpf da obj tc_l2_redirect_kern.o sec l2_to_iptun_ingress_redirect
	fi

  # 用户态程序写入数据. 通过 bpftool map 写数据
  # $REDIRECT_USER -U /sys/fs/bpf/tc/globals/tun_iface -i $(< /sys/class/net/tun1/ifindex)
  bpftool map update pinned /sys/fs/bpf/tc/globals/tun_iface key 0 0 0 0 value $(< /sys/class/net/tun1/ifindex) 0 0 0
  bpftool map dump pinned /sys/fs/bpf/tc/globals/tun_iface -j | jq

  # 重新写一遍，就可以
  sysctl -q -w net.ipv4.conf.tun1.rp_filter=0
  sysctl -q -w net.ipv4.conf.tun1.forwarding=1
  ip netns exec ns1 ping -c 3 10.10.1.102

  if [[ $dir == "egress" ]]; then
    # test direct egress to veth2 (i.e. not forwarding from veth1 to veth2)
    ping -c 3 10.10.1.102
  fi

  # cleanup
	echo "OK"
}

# ecs 验证测试没问题!!!
# bash tc_l2_redirect.sh
cleanup
config_common
loadTCBpf

# 验证没问题，但是不能同时打开
l2_to_ipip ingress
# 验证没问题
#l2_to_ipip egress
