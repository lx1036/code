#!/bin/bash

# /root/linux-5.10.142/tools/testing/selftests/bpf/test_tunnel.sh


cleanup()
{
	ip netns delete vxlan_ns0 2> /dev/null
	ip link del vxlan_tunnel1 2> /dev/null
}


# Set static ARP entry here because iptables set-mark works
# on L3 packet, as a result not applying to ARP packets,
# causing errors at get_tunnel_{key/opt}.

config_device()
{
  ip netns add vxlan_ns0
  ip link add vxlan_veth0 type veth peer name vxlan_veth1
  ip link set vxlan_veth0 netns vxlan_ns0
  ip netns exec vxlan_ns0 ip addr add 173.16.1.100/24 dev vxlan_veth0
  ip netns exec vxlan_ns0 ip link set dev vxlan_veth0 up
  ip link set dev vxlan_veth1 up mtu 1500
  ip addr add dev vxlan_veth1 173.16.1.200/24

  # vxlan_ns0 ns
  ip netns exec vxlan_ns0 ip link add dev vxlan_tunnel0 type vxlan id 2 dstport 4789 gbp remote 173.16.1.200
  ip netns exec vxlan_ns0 ip link set dev vxlan_tunnel0 address 52:54:00:d9:01:00 up
  ip netns exec vxlan_ns0 ip addr add dev vxlan_tunnel0 10.1.1.100/24
  ip netns exec vxlan_ns0 arp -s 10.1.1.200 52:54:00:d9:02:00
  # ip netns exec vxlan_ns0 ip neigh add 10.1.1.200 dev vxlan_tunnel0 lladdr 52:54:00:d9:02:00
  ip netns exec vxlan_ns0 iptables -A OUTPUT -j MARK --set-mark 0x800FF

  # root ns
  ip link add dev vxlan_tunnel1 type vxlan external gbp dstport 4789
  ip link set dev vxlan_tunnel1 address 52:54:00:d9:02:00 up
  ip addr add dev vxlan_tunnel1 10.1.1.200/24
  arp -s 10.1.1.100 52:54:00:d9:01:00
  # ip neigh add 10.1.1.100 dev vxlan_tunnel1 lladdr 52:54:00:d9:01:00

  tc qdisc add dev vxlan_tunnel1 clsact
  tc filter add dev vxlan_tunnel1 egress bpf da obj test_tunnel_kern.o sec vxlan_set_tunnel
  tc filter add dev vxlan_tunnel1 ingress bpf da obj test_tunnel_kern.o sec vxlan_get_tunnel

  ping -c 1 10.1.1.100
  ip netns exec vxlan_ns0 ping -c 1 10.1.1.200
}

#cleanup
config_device



# ping -c 1 10.1.1.100
# vxlan_tunnel1->
