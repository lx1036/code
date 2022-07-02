#!/bin/bash



LIB=$1
RUNDIR=$2
IP4_HOST=$3
IP6_HOST=$4
MODE=$5
# Only set if MODE = "direct", "ipvlan", "flannel"
NATIVE_DEVS=$6
XDP_DEV=$7
XDP_MODE=$8
MTU=$9
IPSEC=${10}
ENCRYPT_DEV=${11}
HOSTLB=${12}
HOSTLB_UDP=${13}
HOSTLB_PEER=${14}
CGROUP_ROOT=${15}
BPFFS_ROOT=${16}
NODE_PORT=${17}
NODE_PORT_BIND=${18}
MCPU=${19}
NODE_PORT_IPV4_ADDRS=${20}
NODE_PORT_IPV6_ADDRS=${21}
NR_CPUS=${22}


set -e
set -x
set -o pipefail

if [[ ! $(command -v cilium-map-migrate) ]]; then
	echo "Can't be initialized because 'cilium-map-migrate' is not in the path."
	exit 1
fi



# INFO: 生产 k8s cilium 没有用 flannel/ipvlan 模式，用的是 direct 模式

# Base device setup
case "${MODE}" in
	"flannel")
		HOST_DEV1="${NATIVE_DEVS}"
		HOST_DEV2="${NATIVE_DEVS}"

		setup_dev "${NATIVE_DEVS}"
		;;
	"ipvlan")
		HOST_DEV1="cilium_host"
		HOST_DEV2="${HOST_DEV1}"

		setup_ipvlan_slave $NATIVE_DEVS $HOST_DEV1

		ip link set $HOST_DEV1 mtu $MTU
		;;
	*)
		HOST_DEV1="cilium_host"
		HOST_DEV2="cilium_net"

		setup_veth_pair $HOST_DEV1 $HOST_DEV2
    # 关闭 arp 和设置 mtu
		ip link set $HOST_DEV1 arp off
		ip link set $HOST_DEV2 arp off

		ip link set $HOST_DEV1 mtu $MTU
		ip link set $HOST_DEV2 mtu $MTU
        ;;
esac

