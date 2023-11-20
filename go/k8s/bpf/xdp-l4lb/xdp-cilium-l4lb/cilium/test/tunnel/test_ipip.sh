


#!/bin/bash


cleanup()
{
	ip netns delete ipip_ns0 2> /dev/null
	ip link del veth1 2> /dev/null
	ip link del ipip11 2> /dev/null

	ip link del ipip6tnl11 2> /dev/null
	ip link del ip6ip6tnl11 2> /dev/null

	ip link del gretap11 2> /dev/null

	ip link del ip6gre11 2> /dev/null
	ip link del ip6gretap11 2> /dev/null

	ip link del vxlan11 2> /dev/null
	ip link del ip6vxlan11 2> /dev/null

	ip link del geneve11 2> /dev/null
	ip link del ip6geneve11 2> /dev/null

	ip link del erspan11 2> /dev/null
	ip link del ip6erspan11 2> /dev/null

	ip xfrm policy delete dir out src 10.1.1.200/32 dst 10.1.1.100/32 2> /dev/null
	ip xfrm policy delete dir in src 10.1.1.100/32 dst 10.1.1.200/32 2> /dev/null
	ip xfrm state delete src 173.16.1.100 dst 173.16.1.200 proto esp spi 0x1 2> /dev/null
	ip xfrm state delete src 173.16.1.200 dst 173.16.1.100 proto esp spi 0x2 2> /dev/null
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
}

add_ipip_tunnel()
{
	# ipip_ns0 namespace
	ip netns exec ipip_ns0 \
		ip link add dev $DEV_NS type $TYPE \
		local 173.16.1.100 remote 173.16.1.200
	ip netns exec ipip_ns0 ip link set dev $DEV_NS up
	ip netns exec ipip_ns0 ip addr add dev $DEV_NS 10.1.1.100/24

	# root namespace
    # root namespace 只能有一个 ipip external 网卡
	ip link add dev $DEV type $TYPE external
	ip link set dev $DEV up
	ip addr add dev $DEV 10.1.1.200/24
}

attach_bpf()
{
	DEV=$1
	SET=$2
	GET=$3
	tc qdisc add dev $DEV clsact
	tc filter add dev $DEV egress bpf da obj test_tunnel_kern.o sec $SET
    # 
	tc filter add dev $DEV ingress bpf da obj test_tunnel_kern.o sec $GET
}

test_ipip()
{
	TYPE=ipip
	DEV_NS=ipip_tunnel0
	DEV=ipip_tunnel1
	ret=0

	config_device
	add_ipip_tunnel

	attach_bpf $DEV ipip_set_tunnel ipip_get_tunnel
	
    # ping -c10 10.1.1.100
	# ip netns exec ipip_ns0 ping -c10 10.1.1.200
	
    # cleanup
}

test_ipip
