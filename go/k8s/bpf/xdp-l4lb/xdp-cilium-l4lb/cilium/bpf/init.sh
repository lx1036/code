
#!/bin/bash



LIB=$1
RUNDIR=$2
IP4_HOST=$3
IP6_HOST=$4
MODE=$5
TUNNEL_MODE=$6
# Only set if MODE = "direct", "ipvlan"
NATIVE_DEVS=$7
HOST_DEV1=$8
HOST_DEV2=$9
MTU=${10}
HOSTLB=${11}
HOSTLB_UDP=${12}
HOSTLB_PEER=${13}
CGROUP_ROOT=${14}
BPFFS_ROOT=${15}
NODE_PORT=${16}
NODE_PORT_BIND=${17}
MCPU=${18}
NR_CPUS=${19}
ENDPOINT_ROUTES=${20}
PROXY_RULE=${21}

ID_HOST=1
ID_WORLD=2

# If the value below is changed, be sure to update bugtool/cmd/configuration.go
# as well when dumping the routing table in bugtool. See GH-5828.
PROXY_RT_TABLE=2005
TO_PROXY_RT_TABLE=2004



set -e
set -x
set -o pipefail





if [ "${TUNNEL_MODE}" != "<nil>" ]; then


  ENCAP_IDX=$(cat /sys/class/net/${ENCAP_DEV}/ifindex)
  sed -i '/^#.*ENCAP_IFINDEX.*$/d' $RUNDIR/globals/node_config.h
  echo "#define ENCAP_IFINDEX $ENCAP_IDX" >> $RUNDIR/globals/node_config.h



else
  # Remove eventual existing encapsulation device from previous run
  ip link del cilium_vxlan 2> /dev/null || true
  ip link del cilium_geneve 2> /dev/null || true
fi
