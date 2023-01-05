
# sr-iov
sr-iov: Single Root I/O Virtualization，是 Intel 在 2007 年提出的一种基于硬件的虚拟化解决方案。比如，一个物理网卡可以虚拟化出多个
虚拟网卡给各个虚机使用。

sr-iov 使用概念 PF(physical functions) 和 VF(virtual functions) 管理 sr-iov 设备全局功能。
启用SR-IOV后，主机将在一个物理NIC上创建单个PF和多个VF。 VF的数量取决于配置和驱动程序支持。


# 参考文献
**[SR-IOV 基本概念](https://zdyxry.github.io/2020/03/12/SR-IOV-%E5%9F%BA%E6%9C%AC%E6%A6%82%E5%BF%B5/)**
