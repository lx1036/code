#!/bin/bash

# /root/linux-5.10.142/tools/testing/selftests/bpf/test_tc_tunnel.sh


readonly port=8000
readonly ns_prefix="ns-$$-"
readonly ns1="${ns_prefix}1"
readonly ns2="${ns_prefix}2"
readonly ns1_v4=192.168.1.1
readonly ns2_v4=192.168.1.2

# Must match port used by bpf program
readonly udpport=5555
# MPLSoverUDP
readonly mplsudpport=6635
readonly mplsproto=137
readonly infile="$(mktemp)"
readonly outfile="$(mktemp)"
readonly addr1="${ns1_v4}"
readonly addr2="${ns2_v4}"
readonly ipproto=4
readonly netcat_opt=-${ipproto}
readonly foumod=fou
readonly foutype=ipip
readonly fouproto=4
readonly fouproto_mpls=${mplsproto}
readonly gretaptype=gretap

readonly tuntype=$2
readonly mac=$3
readonly datalen=$4

echo "encap ${addr1} to ${addr2}, type ${tuntype}, mac ${mac} len ${datalen}"

setup() {
	ip netns add "${ns1}"
	ip netns add "${ns2}"

	ip link add dev veth1 mtu 1500 netns "${ns1}" type veth peer name veth2 mtu 1500 netns "${ns2}"

	ip netns exec "${ns1}" ethtool -K veth1 tso off

	ip -netns "${ns1}" link set veth1 up
	ip -netns "${ns2}" link set veth2 up

	ip -netns "${ns1}" -4 addr add "${ns1_v4}/24" dev veth1
	ip -netns "${ns2}" -4 addr add "${ns2_v4}/24" dev veth2

	# clamp route to reserve room for tunnel headers
	ip -netns "${ns1}" -4 route flush table main
	ip -netns "${ns1}" -4 route add "${ns2_v4}" mtu 1458 dev veth1

	sleep 1
}

cleanup() {
	ip netns del "${ns2}"
	ip netns del "${ns1}"

	if [[ -f "${outfile}" ]]; then
		rm "${outfile}"
	fi
	if [[ -f "${infile}" ]]; then
		rm "${infile}"
	fi

	if [[ -n $server_pid ]]; then
		kill $server_pid 2> /dev/null
	fi
}

server_listen() {
	ip netns exec "${ns2}" nc "${netcat_opt}" -l -p "${port}" > "${outfile}" &
	server_pid=$!
	sleep 0.2
}

client_connect() {
	ip netns exec "${ns1}" timeout 2 nc "${netcat_opt}" -w 1 "${addr2}" "${port}" < "${infile}"
	echo $?
}

verify_data() {
	wait "${server_pid}"
#	server_pid=
	# sha1sum returns two fields [sha1] [filepath]
	# convert to bash array and access first elem
	insum=($(sha1sum ${infile}))
	outsum=($(sha1sum ${outfile}))
	if [[ "${insum[0]}" != "${outsum[0]}" ]]; then
		echo "data mismatch"
		exit 1
	fi
}

setup

# 1. basic communication works
echo "test basic connectivity"
server_listen
client_connect
verify_data

# 2. clientside, insert bpf program to encap all TCP to port ${port}
# client can no longer connect
ip netns exec "${ns1}" tc qdisc add dev veth1 clsact
ip netns exec "${ns1}" tc filter add dev veth1 egress \
	bpf direct-action object-file ./test_tc_tunnel.o \
	section "encap_${tuntype}_${mac}"
echo "test bpf encap without decap (expect failure)"
server_listen
! client_connect


if [[ "$tuntype" =~ "udp" ]]; then
	# Set up fou tunnel.
	ttype="${foutype}"
	targs="encap fou encap-sport auto encap-dport $udpport"
	# fou may be a module; allow this to fail.
	modprobe "${foumod}" ||true
	if [[ "$mac" == "mpls" ]]; then
		dport=${mplsudpport}
		dproto=${fouproto_mpls}
		tmode="mode any ttl 255"
	else
		dport=${udpport}
		dproto=${fouproto}
	fi
	ip netns exec "${ns2}" ip fou add port $dport ipproto ${dproto}
	targs="encap fou encap-sport auto encap-dport $dport"
elif [[ "$tuntype" =~ "gre" && "$mac" == "eth" ]]; then
	ttype=$gretaptype
else
	ttype=$tuntype
	targs=""
fi

# tunnel address family differs from inner for SIT
if [[ "${tuntype}" == "sit" ]]; then
	link_addr1="${ns1_v4}"
	link_addr2="${ns2_v4}"
else
	link_addr1="${addr1}"
	link_addr2="${addr2}"
fi


