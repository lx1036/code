

# Cilium IPAM
IPAM: https://docs.cilium.io/en/stable/concepts/networking/ipam/

Cilium Operator 创建 CiliumNode，并通过 IPAM 来从 Cluster CIDR 中分配 Pod CIDR，给 Daemon Agent 使用。
Cilium IPAM Mode: cluster-pool(默认)、crd(自定义)、aws eni、azure、alibaba cloud
cluster-pool: https://docs.cilium.io/en/stable/concepts/networking/ipam/cluster-pool/
crd: https://docs.cilium.io/en/stable/gettingstarted/ipam-cluster-pool/
