



# 基本概念
(1)linux 内核开启 nvme-rdma 和 iscsi-tcp 模块
```shell
modprobe nvme-rdma iscsi-tcp
modinfo nvme-rdma
modinfo iscsi-tcp

# 安装 nvme/iscsi cli 客户端
yum install -y nvme-cli nvmetcli open-iscsi targetcli e2fsprogs xfsprogs blkid
```

(2)RDMA(Remote Direct Memory Access)
RDMA是一种新的内存访问技术，RDMA让计算机可以直接存取其他计算机的内存，而不需要经过处理器耗时的处理。




