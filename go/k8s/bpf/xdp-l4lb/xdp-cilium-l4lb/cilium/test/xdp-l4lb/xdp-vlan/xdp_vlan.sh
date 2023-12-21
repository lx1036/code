#!/bin/bash

# /root/linux-5.10.142/tools/testing/selftests/bpf/test_xdp_vlan.sh

# Exit on failure
set -e

which ip > /dev/null
which tc > /dev/null
which ethtool > /dev/null

ip netns add ns1
ip netns add ns2
ip link add veth1 type veth peer name veth2
ip link set veth1 netns ns1
ip link set veth2 netns ns2
# NOTICE: XDP require VLAN header inside packet payload
#  - Thus, disable VLAN offloading driver features
#  - For veth REMEMBER TX side VLAN-offload
#
# Disable rx-vlan-offload (mostly needed on ns1)
ip netns exec ns1 ethtool -K veth1 rxvlan off
ip netns exec ns2 ethtool -K veth2 rxvlan off
#
# Disable tx-vlan-offload (mostly needed on ns2)
ip netns exec ns2 ethtool -K veth2 txvlan off
ip netns exec ns1 ethtool -K veth1 txvlan off

export IPADDR1=100.64.41.1
export IPADDR2=100.64.41.2
# In ns1/veth1 add IP-addr on plain net_device
ip netns exec ns1 ip addr add ${IPADDR1}/24 dev veth1
ip netns exec ns1 ip link set veth1 up

# In ns2/veth2 create VLAN device, veth2.4011
export VLAN=4011
export DEVNS2=veth2
ip netns exec ns2 ip link add link $DEVNS2 name $DEVNS2.$VLAN type vlan id $VLAN
ip netns exec ns2 ip addr add ${IPADDR2}/24 dev $DEVNS2.$VLAN
ip netns exec ns2 ip link set $DEVNS2 up
ip netns exec ns2 ip link set $DEVNS2.$VLAN up

ip netns exec ns1 ip link set lo up
ip netns exec ns2 ip link set lo up

# At this point, the hosts cannot reach each-other,
# because ns2 are using VLAN tags on the packets.
ip netns exec ns2 sh -c 'ping -W 1 -c 1 100.64.41.1 || echo "Success: First ping must fail"'

# Now we can use the test_xdp_vlan.c program to pop/push these VLAN tags
# ----------------------------------------------------------------------
# In ns1: ingress use XDP to remove VLAN tags
# xdp,xdpgeneric,xdpdrv
XDP_MODE=xdp

export DEVNS1=veth1
export FILE=test_xdp_vlan.o

# First test: Remove VLAN by setting VLAN ID 0, using "xdp_vlan_change"
export XDP_PROG=xdp_vlan_change
ip netns exec ns1 ip link set $DEVNS1 $XDP_MODE object $FILE section $XDP_PROG

# In ns1: egress use TC to add back VLAN tag 4011
#  (del cmd)
#  tc qdisc del dev $DEVNS1 clsact 2> /dev/null
#
ip netns exec ns1 tc qdisc add dev $DEVNS1 clsact
ip netns exec ns1 tc filter add dev $DEVNS1 egress \
  prio 1 handle 1 bpf da obj $FILE sec tc_vlan_push

# Now the namespaces can reach each-other, test with ping:
ip netns exec ns2 ping -i 0.2 -W 2 -c 2 $IPADDR1
ip netns exec ns1 ping -i 0.2 -W 2 -c 2 $IPADDR2

# Second test: Replace xdp prog, that fully remove vlan header
#
# Catch kernel bug for generic-XDP, that does didn't allow us to
# remove a VLAN header, because skb->protocol still contain VLAN
# ETH_P_8021Q indication, and this cause overwriting of our changes.
#
export XDP_PROG=xdp_vlan_remove_outer2
ip netns exec ns1 ip link set $DEVNS1 $XDP_MODE off
ip netns exec ns1 ip link set $DEVNS1 $XDP_MODE object $FILE section $XDP_PROG

# Now the namespaces should still be able reach each-other, test with ping:
ip netns exec ns2 ping -i 0.2 -W 2 -c 2 $IPADDR1
ip netns exec ns1 ping -i 0.2 -W 2 -c 2 $IPADDR2


cleanup()
{
  ip netns del ns1 2> /dev/null
  ip netns del ns2 2> /dev/null
}
