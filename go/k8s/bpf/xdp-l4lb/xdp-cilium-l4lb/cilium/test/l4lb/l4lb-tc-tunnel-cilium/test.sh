



# 验证：安装完软件后，然后直接运行 `sh test.sh`，验证没有问题

# ipip demo 验证: https://blog.csdn.net/WuYuChen20/article/details/104572969


#!/bin/bash

# set -eux

IMG_OWNER=${1:-cilium}
IMG_TAG=${2:-v1.10.20}
HELM_CHART_DIR=${3:-./kubernetes/cilium}

###########
#  SETUP  #
###########

# bpf_xdp_veth_host is a dummy XDP program which is going to be attached to LB
# node's veth pair end in the host netns. When bpf_xdp, which is attached in
# the container netns, forwards a LB request with XDP_TX, the request needs to
# be picked in the host netns by a NAPI handler. To register the handler, we
# attach the dummy program.
apt-get update -y
apt-get install -y gcc-multilib libbpf-dev

# 这里安装 libbpf-dev 包后，代码里可以直接 include linux 头文件
clang -O2 -Wall -target bpf -c bpf_xdp_veth_host.c -o bpf_xdp_veth_host.o

# worker node 收到的包还是 IPIP 包，在 worker node 上没有创建 ipip 类型网卡来解包
# 通过在 eth0 上挂载 tc ingress bpf 程序来解析 ipip 包

# The worker (aka backend node) will receive IPIP packets from the LB node.
# To decapsulate the packets instead of creating an ipip dev which would
# complicate network setup, we will attach the following program which
# terminates the tunnel.
# The program is taken from the Linux kernel selftests.
clang -O2 -Wall -target bpf -c test_tc_tunnel.c -o test_tc_tunnel.o

# With Kind we create two nodes cluster:
#
# * "kind-control-plane" runs cilium in the LB-only mode.
# * "kind-worker" runs the nginx server.
#
# The LB cilium does not connect to the kube-apiserver. For now we use Kind
# just to create Docker-in-Docker containers.
kind create cluster --config kind-config.yaml --image=kindest/node:v1.19.16

# l4lb-veth0(host) 3.3.3.1 <-> l4lb-veth1(control-plane) 3.3.3.2

# Create additional veth pair which is going to be used to test XDP_REDIRECT.
ip l a l4lb-veth0 type veth peer l4lb-veth1
SECOND_LB_NODE_IP=3.3.3.2
ip a a "3.3.3.1/24" dev l4lb-veth0
CONTROL_PLANE_PID=$(docker inspect kind-control-plane -f '{{ .State.Pid }}')
ip l s dev l4lb-veth1 netns $CONTROL_PLANE_PID
ip l s dev l4lb-veth0 up
nsenter -t $CONTROL_PLANE_PID -n /bin/sh -c "\
    ip a a "${SECOND_LB_NODE_IP}/24" dev l4lb-veth1 && \
    ip l s dev l4lb-veth1 up"

# Install Cilium as standalone L4LB
# 新的安装方式：
# helm repo add cilium https://helm.cilium.io/
# helm install cilium cilium/cilium --version 1.14.4 --namespace kube-system
# 文件来自 https://github.com/cilium/cilium/blob/511463db8f42541cef3730138a58591dce2f3a44/install/kubernetes/cilium
helm install cilium ${HELM_CHART_DIR} \
    --wait \
    --namespace kube-system \
    --set image.repository="quay.io/${IMG_OWNER}/cilium" \
    --set image.tag="${IMG_TAG}" \
    --set image.useDigest=false \
    --set image.pullPolicy=IfNotPresent \
    --set operator.enabled=false \
    --set loadBalancer.standalone=true \
    --set loadBalancer.algorithm=maglev \
    --set loadBalancer.mode=dsr \
    --set loadBalancer.acceleration=native \
    --set loadBalancer.dsrDispatch=ipip \
    --set devices='{eth0,l4lb-veth1}' \
    --set nodePort.directRoutingDevice=eth0 \
    --set ipv6.enabled=false \
    --set affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key="kubernetes.io/hostname" \
    --set affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator=In \
    --set affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]=kind-control-plane

# xdp bpf 程序挂载在 l4lb-veth0 和 vethf919756@if33(control-plane node 里 eth0 的 veth-pair)
IFIDX=$(docker exec -i kind-control-plane \
    /bin/sh -c 'echo $(( $(ip -o l show eth0 | awk "{print $1}" | cut -d: -f1) ))')
LB_VETH_HOST=$(ip -o l | grep "if$IFIDX" | awk '{print $2}' | cut -d@ -f1)
ip l set dev $LB_VETH_HOST xdp obj bpf_xdp_veth_host.o
ip l set dev l4lb-veth0 xdp obj bpf_xdp_veth_host.o

# Disable TX and RX csum offloading, as veth does not support it. Otherwise,
# the forwarded packets by the LB to the worker node will have invalid csums.
ethtool -K $LB_VETH_HOST rx off tx off
ethtool -K l4lb-veth0 rx off tx off

# worker node 上 eth0 挂载 tc ingress IPIP 解包程序
# `tc filter show dev eth0 ingress`
# 卸载 tc 程序：`tc filter del dev eth0 ingress pref 49152`
docker exec kind-worker /bin/sh -c 'apt-get update && apt-get install -y nginx && systemctl start nginx'
WORKER_IP=$(docker exec kind-worker ip -o -4 a s eth0 | awk '{print $4}' | cut -d/ -f1)
nsenter -t $(docker inspect kind-worker -f '{{ .State.Pid }}') -n /bin/sh -c \
    'tc qdisc add dev eth0 clsact && tc filter add dev eth0 ingress bpf direct-action object-file ./test_tc_tunnel.o section decap'

collect_sysdump() {
    curl -sLO https://github.com/cilium/cilium-sysdump/releases/latest/download/cilium-sysdump.zip
    python cilium-sysdump.zip --output /tmp/cilium-sysdump-out
}

trap collect_sysdump ERR

CILIUM_POD_NAME=$(kubectl -n kube-system get pod -l k8s-app=cilium -o=jsonpath='{.items[0].metadata.name}')
echo "CILIUM_POD_NAME: $CILIUM_POD_NAME"
kubectl -n kube-system wait --for=condition=Ready pod "$CILIUM_POD_NAME" --timeout=5m

##########
#  TEST  #
##########

LB_VIP="2.2.2.2"

nsenter -t $(docker inspect kind-worker -f '{{ .State.Pid }}') -n /bin/sh -c \
    "ip a a dev eth0 ${LB_VIP}/32"

kubectl -n kube-system exec "${CILIUM_POD_NAME}" -- \
    cilium service update --id 1 --frontend "${LB_VIP}:80" --backends "${WORKER_IP}:80" --k8s-node-port

LB_NODE_IP=$(docker exec kind-control-plane ip -o -4 a s eth0 | awk '{print $4}' | cut -d/ -f1)
ip r a "${LB_VIP}/32" via "$LB_NODE_IP"

# Issue 10 requests to LB
for i in $(seq 1 10); do
    curl -o /dev/null "${LB_VIP}:80"
done

# Now steer the traffic to LB_VIP via the secondary device so that XDP_REDIRECT
# can be tested on the L4LB node
ip r d "${LB_VIP}/32"
ip r a "${LB_VIP}/32" via "$SECOND_LB_NODE_IP"

# Issue 10 requests to LB
for i in $(seq 1 10); do
    curl -o /dev/null "${LB_VIP}:80"
done
