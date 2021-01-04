

**[calico,CNI的一种实现](https://www.yuque.com/baxiaoshi/tyado3/lvfa0b)**

**[containernetworking/cni](https://github.com/containernetworking/cni)**
**[projectcalico/cni-plugin](https://github.com/projectcalico/cni-plugin)**












# Kubernetes学习笔记之Calico IPAM Plugin源码解析

## Overview
IP地址分配的性能有哪些问题要考虑呢？在大规模集群的场景下，Calico IP地址的分配速率是否受到集群规模的限制?
IP地址和Block size怎么配置才能保持高速的IP地址分配?另外，Calico的IP地址在Node节点异常时，IP地址如何回收?什么时候有可能产生IP地址冲突？
为了解答这些疑问，需要熟悉CalicoIP地址分配的执行流程。







```shell
sudo journalctl --since="2020-12-27 01:04:00" -u kubelet
```



## 参考文献
**[Use a specific IP address with a pod](https://docs.projectcalico.org/networking/use-specific-ip)**
**[Calico IPAM源码解析](https://mp.weixin.qq.com/s/lyfeZh6VWWjXuLY8fl3ciw)**
