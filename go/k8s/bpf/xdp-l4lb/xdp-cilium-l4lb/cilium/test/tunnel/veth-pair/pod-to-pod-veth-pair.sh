
cleanup()
{
  ip netns del pod1_ns
  ip netns del pod2_ns
  rm -rf /sys/fs/bpf/tc/globals/ding_lxc
}


# 验证通过
config_device()
{
  ip netns add pod1_ns
  ip netns add pod2_ns
  ip link add pod1_veth type veth peer name eth0 netns pod1_ns
  ip link add pod2_veth type veth peer name eth0 netns pod2_ns
  ip link set pod1_veth up
  ip link set pod2_veth up
  ip link set dev pod1_veth address ee:ee:ee:ee:ee:ee
  ip link set dev pod2_veth address ee:ee:ee:ee:ee:ee

  ip netns exec pod1_ns ip link set eth0 up
  ip netns exec pod2_ns ip link set eth0 up
  ip netns exec pod1_ns ip link set lo up
  ip netns exec pod2_ns ip link set lo up
  ip netns exec pod1_ns ip addr add 100.0.1.1/32 dev eth0
  ip netns exec pod2_ns ip addr add 100.0.1.2/32 dev eth0

  ip netns exec pod1_ns ip route add 169.254.1.1/32 dev eth0
  ip netns exec pod1_ns ip route add default via 169.254.1.1 dev eth0
  ip netns exec pod2_ns ip route add 169.254.1.1/32 dev eth0
  ip netns exec pod2_ns ip route add default via 169.254.1.1 dev eth0
  ip netns exec pod1_ns ip neigh replace 169.254.1.1 dev eth0 lladdr ee:ee:ee:ee:ee:ee
  #ip netns exec pod1_ns ip neigh add 169.254.1.1 dev eth0 lladdr ee:ee:ee:ee:ee:ee
  ip netns exec pod2_ns ip neigh replace 169.254.1.1 dev eth0 lladdr ee:ee:ee:ee:ee:ee

  ip route add 100.0.1.1/32 dev pod1_veth
#  ip route del 100.0.1.1/32 dev pod1_veth
  ip route add 100.0.1.2/32 dev pod2_veth
#  ip route del 100.0.1.2/32 dev pod2_veth

  # 必须同时打开 forwarding 和 proxy_arp，开启 pod1_veth/pod2_veth 网卡的 proxy_arp 应答
  echo 1 > /proc/sys/net/ipv4/conf/pod1_veth/proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod1_veth/forwarding
  # 必须同时打开 forwarding 和 proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod2_veth/proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod2_veth/forwarding

  ip netns exec pod1_ns ping -c 3 100.0.1.2
  ip netns exec pod2_ns ping -c 3 100.0.1.1
}

# 验证通过
config_device2()
{
  ip netns add pod1_ns
  ip netns add pod2_ns
  ip link add pod1_veth type veth peer name eth0 netns pod1_ns
  ip link add pod2_veth type veth peer name eth0 netns pod2_ns
  ip link set pod1_veth up
  ip link set pod2_veth up

  ip netns exec pod1_ns ip link set eth0 up
  ip netns exec pod2_ns ip link set eth0 up
  ip netns exec pod1_ns ip link set lo up
  ip netns exec pod2_ns ip link set lo up
  ip netns exec pod1_ns ip addr add 100.0.1.1/32 dev eth0
  ip netns exec pod2_ns ip addr add 100.0.1.2/32 dev eth0
  ip netns exec pod1_ns ip route add default dev eth0
  ip netns exec pod1_ns ip route add 100.0.1.0/24 dev eth0
  ip netns exec pod2_ns ip route add default dev eth0
  ip netns exec pod2_ns ip route add 100.0.1.0/24 dev eth0

  # 必须同时打开 forwarding 和 proxy_arp，开启 pod1_veth/pod2_veth 网卡的 proxy_arp 应答
  echo 1 > /proc/sys/net/ipv4/conf/pod1_veth/proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod1_veth/forwarding
  # 必须同时打开 forwarding 和 proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod2_veth/proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod2_veth/forwarding

  ip route add 100.0.1.1/32 dev pod1_veth
#  ip route del 100.0.1.1/32 dev pod1_veth
  ip route add 100.0.1.2/32 dev pod2_veth
#  ip route del 100.0.1.2/32 dev pod2_veth

  ip netns exec pod1_ns ping -c 3 100.0.1.2
  ip netns exec pod2_ns ping -c 3 100.0.1.1
}

# 验证通过
config_device3() {
  ip netns add pod1_ns
  ip netns add pod2_ns
  ip link add pod1_veth type veth peer name eth0 netns pod1_ns
  ip link add pod2_veth type veth peer name eth0 netns pod2_ns
  ip link set pod1_veth up
  ip link set pod2_veth up

  ip netns exec pod1_ns ip link set eth0 up
  ip netns exec pod2_ns ip link set eth0 up
  ip netns exec pod1_ns ip link set lo up
  ip netns exec pod2_ns ip link set lo up
  ip netns exec pod1_ns ip addr add 100.0.1.1/32 dev eth0
  ip netns exec pod2_ns ip addr add 100.0.1.2/32 dev eth0
  ip netns exec pod1_ns ip route add default dev eth0
  ip netns exec pod2_ns ip route add default dev eth0
  ip netns exec pod1_ns ip route add 100.0.1.0/24 dev eth0
  ip netns exec pod2_ns ip route add 100.0.1.0/24 dev eth0

  # 必须同时打开 forwarding 和 proxy_arp，开启 pod1_veth/pod2_veth 网卡的 proxy_arp 应答
  echo 1 > /proc/sys/net/ipv4/conf/pod1_veth/proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod1_veth/forwarding
  # 必须同时打开 forwarding 和 proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod2_veth/proxy_arp
  echo 1 > /proc/sys/net/ipv4/conf/pod2_veth/forwarding

  ip route add 100.0.1.1/32 dev pod1_veth
#  ip route del 100.0.1.1/32 dev pod1_veth
  ip route add 100.0.1.2/32 dev pod2_veth
#  ip route del 100.0.1.2/32 dev pod2_veth

  ######### 相比于 config_device2，这里通过 ebpf 程序来减少包跳转网卡 ##############
  # 包的流程:
  # eth0(pod1_ns) -> tcpdump 抓包, pod1_veth(tc ingress) -> bpf_redirect_peer -> eth0(pod2_ns) -> 回包
  # -> tcpdump 抓包, pod2_veth(tc ingress) -> bpf_redirect_peer -> eth0(pod1_ns)
  # 相比于没有 ebpf, 减少了 pod1_veth 和 pod2_veth netfilter 内核协议栈


  tc qdisc add dev pod1_veth clsact
  tc filter add dev pod1_veth ingress bpf da obj veth_ingress.o sec tc_from_container
 # tc filter replace dev pod1_veth ingress bpf da obj veth_ingress.o sec tc_from_container
  tc filter show dev pod1_veth ingress
  # tc filter del dev pod1_veth ingress
  tc qdisc add dev pod2_veth clsact
  tc filter add dev pod2_veth ingress bpf da obj veth_ingress.o sec tc_from_container
 # tc filter replace dev pod2_veth ingress bpf da obj veth_ingress.o sec tc_from_container
  tc filter show dev pod2_veth ingress
  # tc filter del dev pod2_veth ingress
  ls /sys/fs/bpf/tc/globals/
  # rm -rf /sys/fs/bpf/tc/424e22c8e74276a6484f398886d426f441d9b849/ding_lxc
  # bpftool map dump pinned /sys/fs/bpf/tc/globals/ding_lxc -j | jq

  ip netns exec pod1_ns ping -c 3 100.0.1.2
  ip netns exec pod2_ns ping -c 3 100.0.1.1
}

# cleanup

#config_device
#config_device2

# config_device3
