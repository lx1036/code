#!/bin/bash

# /root/linux-5.10.142/tools/testing/selftests/bpf/test_tc_redirect.sh


# This test sets up 3 netns (src <-> fwd <-> dst). There is no direct veth link
# between src and dst. The netns fwd has veth links to each src and dst. The
# client is in src and server in dst. The test installs a TC BPF program to each
# host facing veth in fwd which calls into i) bpf_redirect_neigh() to perform the
# neigh addr population and redirect or ii) bpf_redirect_peer() for namespace
# switch from ingress side; it also installs a checker prog on the egress side
# to drop unexpected traffic.


netns_cleanup()
{
  ip netns del ns-src
  ip netns del ns-fwd
  ip netns del ns-dst
}

# 验证通过
# src <-> fwd <-> dst, 通过添加 route 和 neigh 实现连通性
# 没有 ebpf，所有网卡 netfilter 都要走，没有网络加速
netns_setup()
{
  ip netns add ns-src
  ip netns add ns-fwd
  ip netns add ns-dst

  ip link add veth_src type veth peer name veth_src_fwd
  ip link add veth_dst type veth peer name veth_dst_fwd
  ip link set veth_src netns ns-src
  ip link set veth_src_fwd netns ns-fwd
  ip link set veth_dst netns ns-dst
  ip link set veth_dst_fwd netns ns-fwd

  ip netns exec ns-src ip addr add 173.16.1.100/32 dev veth_src
  ip netns exec ns-dst ip addr add 173.16.2.100/32 dev veth_dst
  ip netns exec ns-fwd ip addr add 169.254.0.1/32 dev veth_src_fwd
  ip netns exec ns-fwd ip addr add 169.254.0.2/32 dev veth_dst_fwd
  ip netns exec ns-src ip link set dev veth_src up
  ip netns exec ns-fwd ip link set dev veth_src_fwd up
  ip netns exec ns-dst ip link set dev veth_dst up
  ip netns exec ns-fwd ip link set dev veth_dst_fwd up

  # The fwd netns automatically get a v6 LL address / routes, but also
  # needs v4 one in order to start ARP probing. IP4_NET route is added
  # to the endpoints so that the ARP processing will reply.
  ip netns exec ns-src ip route add 173.16.2.100/32 dev veth_src scope global
  ip netns exec ns-src ip route add 169.254.0.0/16 dev veth_src scope global
  ip netns exec ns-fwd ip route add 173.16.1.100/32 dev veth_src_fwd scope global
  ip netns exec ns-fwd ip route add 173.16.2.100/32 dev veth_dst_fwd scope global
  ip netns exec ns-dst ip route add 173.16.1.100/32 dev veth_dst scope global
  ip netns exec ns-dst ip route add 169.254.0.0/16 dev veth_dst scope global

  fmac_src=$(ip netns exec ns-fwd cat /sys/class/net/veth_src_fwd/address)
  fmac_dst=$(ip netns exec ns-fwd cat /sys/class/net/veth_dst_fwd/address)
  ip netns exec ns-src ip neigh add 173.16.2.100 dev veth_src lladdr $fmac_src
  ip netns exec ns-dst ip neigh add 173.16.1.100 dev veth_dst lladdr $fmac_dst

  ip netns exec ns-src ping -c 1 173.16.2.100
  ip netns exec ns-dst ping -c 1 173.16.1.100
}

hex_mem_str()
{
  perl -e 'print join(" ", unpack("(H2)8", pack("L", @ARGV)))' $1
}

# 下发 ebpf 程序来网络加速
netns_setup_bpf()
{
  local obj=$1
  # ${2:-0} 的含义是：如果位置参数 $2 已经被赋予了一个非空值，则使用该值；否则，使用默认值 0。这里的 $2 表示命令行参数中的第二个参数。
  local use_forwarding=${2:-0}

  ip netns exec ns-fwd tc qdisc add dev veth_src_fwd clsact
  ip netns exec ns-fwd tc filter add dev veth_src_fwd ingress bpf da obj $obj sec src_ingress
  ip netns exec ns-fwd tc filter add dev veth_src_fwd egress  bpf da obj $obj sec chk_egress

  ip netns exec ns-fwd tc qdisc add dev veth_dst_fwd clsact
  ip netns exec ns-fwd tc filter add dev veth_dst_fwd ingress bpf da obj $obj sec dst_ingress
  ip netns exec ns-fwd tc filter add dev veth_dst_fwd egress  bpf da obj $obj sec chk_egress

  if [ "$use_forwarding" -eq "1" ]; then
    # bpf_fib_lookup() checks if forwarding is enabled
    ip netns exec ns-fwd sysctl -w net.ipv4.ip_forward=1
    return 0
  fi

  veth_src=$(ip netns exec ns-fwd cat /sys/class/net/veth_src_fwd/ifindex)
  veth_dst=$(ip netns exec ns-fwd cat /sys/class/net/veth_dst_fwd/ifindex)
  progs=$(ip netns exec ns-fwd bpftool net --json | jq -r '.[] | .tc | map(.id) | .[]')
  for prog in $progs; do
    map=$(bpftool prog show id $prog --json | jq -r '.map_ids | .? | .[]')
    if [ ! -z "$map" ]; then
      bpftool map update id $map key hex $(hex_mem_str 0) value hex $(hex_mem_str $veth_src)
      bpftool map update id $map key hex $(hex_mem_str 1) value hex $(hex_mem_str $veth_dst)
      bpftool map dump id $map -j | jq
    fi
  done
}

netns_test_connectivity()
{
  echo -e "TCPv4 connectivity test"
  ip netns exec ns-dst bash -c "python3 -m http.server 8080 &"
  ip netns exec ns-src bash -c "curl 173.16.2.100:8080"

  echo -e "ICMPv4 connectivity test"
  ip netns exec ns-src ping -c 3 173.16.2.100
  ip netns exec ns-dst ping -c 3 173.16.1.100
}

# 验证通过
netns_cleanup
netns_setup
netns_setup_bpf test_tc_peer.o
netns_test_connectivity

## 以下两个 bpf 程序，没有看到网络加速效果，很奇怪!!!

# 验证通过
netns_cleanup
netns_setup
netns_setup_bpf test_tc_neigh.o
netns_test_connectivity

# 验证通过
netns_cleanup
netns_setup
netns_setup_bpf test_tc_neigh_fib.o 1
netns_test_connectivity
