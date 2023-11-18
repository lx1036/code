

#!/bin/bash

# /root/linux-5.10.142/samples/bpf/tc_l2_redirect.sh


[[ -z $TC ]] && TC='tc'
[[ -z $IP ]] && IP='ip'


RP_FILTER=$(< /proc/sys/net/ipv4/conf/all/rp_filter)
IPV6_FORWARDING=$(< /proc/sys/net/ipv6/conf/all/forwarding)

REDIRECT_USER='./tc_l2_redirect'
REDIRECT_BPF='./tc_l2_redirect_kern.o'


VIP='10.10.1.102'

function cleanup {
    set +e
    
    $IP netns delete ns1 >& /dev/null
	$IP netns delete ns2 >& /dev/null
	$IP link del ve1 >& /dev/null
	$IP link del ve2 >& /dev/null
	$IP link del tun1 >& /dev/null
	$IP link del ip6t >& /dev/null

    sysctl -q -w net.ipv4.conf.all.rp_filter=$RP_FILTER
	sysctl -q -w net.ipv6.conf.all.forwarding=$IPV6_FORWARDING
	rm -f /sys/fs/bpf/tc/globals/tun_iface

    set -e
}

function config_common {
    $IP netns add ns1
	$IP netns add ns2
	$IP link add ve1 type veth peer name vens1
	$IP link add ve2 type veth peer name vens2
	$IP link set dev ve1 up
	$IP link set dev ve2 up
	$IP link set dev ve1 mtu 1500
	$IP link set dev ve2 mtu 1500
	$IP link set dev vens1 netns ns1
	$IP link set dev vens2 netns ns2

    # ns1
    $IP -n ns1 link set dev lo up
	$IP -n ns1 link set dev vens1 up
	$IP -n ns1 addr add 10.1.1.101/24 dev vens1
	# $IP -n ns1 addr add 2401:db01::65/64 dev vens1 nodad
	$IP -n ns1 route add default via 10.1.1.1 dev vens1
	# $IP -n ns1 route add default via 2401:db01::1 dev vens1

    # ns2
    $IP -n ns2 link set dev lo up
	$IP -n ns2 link set dev vens2 up
	$IP -n ns2 addr add 10.2.1.102/24 dev vens2
	# $IP -n ns2 addr add 2401:db02::66/64 dev vens2 nodad
	$IP -n ns2 addr add $VIP dev lo
	# $IP -n ns2 addr add 2401:face::66/64 dev lo nodad
    $IP -n ns2 link add tun2 type ipip local 10.2.1.102 remote 10.2.1.1
	# $IP -n ns2 link add ip6t2 type ip6tnl mode any local 2401:db02::66 remote 2401:db02::1
    $IP -n ns2 link set dev tun2 up
	# $IP -n ns2 link set dev ip6t2 up
    $IP netns exec ns2 $TC qdisc add dev vens2 clsact
	$IP netns exec ns2 $TC filter add dev vens2 ingress bpf da obj $REDIRECT_BPF sec drop_non_tun_vip

    # ipip
	# add route for return path，回包时走该路由
    $IP -n ns2 route add 10.1.1.0/24 dev tun2
	$IP netns exec ns2 sysctl -q -w net.ipv4.conf.all.rp_filter=0
	$IP netns exec ns2 sysctl -q -w net.ipv4.conf.tun2.rp_filter=0

    $IP addr add 10.1.1.1/24 dev ve1
	# $IP addr add 2401:db01::1/64 dev ve1 nodad
	$IP addr add 10.2.1.1/24 dev ve2
	# $IP addr add 2401:db02::1/64 dev ve2 nodad

    $TC qdisc add dev ve2 clsact
	$TC filter add dev ve2 ingress bpf da obj $REDIRECT_BPF sec l2_to_iptun_ingress_forward

	sysctl -q -w net.ipv4.conf.all.rp_filter=0
	sysctl -q -w net.ipv6.conf.all.forwarding=1
}


function l2_to_ipip {
    echo -n "l2_to_ipip $1: "

	local dir=$1

    config_common

	# tun1 网卡的作用是啥?

    $IP link add tun1 type ipip external
	$IP link set dev tun1 up
	sysctl -q -w net.ipv4.conf.tun1.rp_filter=0
	sysctl -q -w net.ipv4.conf.tun1.forwarding=1

    if [[ $dir == "egress" ]]; then
		$IP route add 10.10.1.0/24 via 10.2.1.102 dev ve2
		$TC filter add dev ve2 egress bpf da obj $REDIRECT_BPF sec l2_to_iptun_ingress_redirect
		sysctl -q -w net.ipv4.conf.ve1.forwarding=1
	else
		$TC qdisc add dev ve1 clsact
		$TC filter add dev ve1 ingress bpf da obj $REDIRECT_BPF sec l2_to_iptun_ingress_redirect
	fi

    # 用户态程序写入数据. 通过 bpftool map 写数据
    # $REDIRECT_USER -U /sys/fs/bpf/tc/globals/tun_iface -i $(< /sys/class/net/tun1/ifindex)
	bpftool map update pinned /sys/fs/bpf/tc/globals/tun_iface key 0 0 0 0 value $(< /sys/class/net/tun1/ifindex) 0 0 0
	bpftool map dump pinned /sys/fs/bpf/tc/globals/tun_iface -j | jq

    $IP netns exec ns1 ping -c10 $VIP #>& /dev/null

    if [[ $dir == "egress" ]]; then
		# test direct egress to ve2 (i.e. not forwarding from
		# ve1 to ve2).
		ping -c10 $VIP >& /dev/null
	fi

    # cleanup

	echo "OK"
}


# ecs 验证测试没问题!!!

# bash tc_l2_redirect.sh
cleanup
l2_to_ipip ingress
# l2_to_ipip egress
