





```shell
cd /sys/fs/cgroup/
ls cpu/kubepods/
```





# cpuset controller 
**[cpuset](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#cpuset)**

cpuset controller可以控制实现绑核，让pods独占cpu，免得cpu切换降低性能，这对cpu密集型pod很实用。
```shell
cat /sys/fs/cgroup/cpuset/cpuset.cpu # 查看当前cgroup中可以被task使用的cpus
# 0-31

```











## 参考文献
containerd: https://github.com/containerd/cgroups
runc: https://github.com/opencontainers/runc/blob/master/libcontainer/cgroups/cgroups.go


## troubleshooting
(1) 检查 cgroup v2 是否安装
https://www.cnblogs.com/rongfengliang/p/10930455.html


(2) cgroup driver: cgroupfs 和 systemd 的区别？
从[Cgroupfs限制CPU、内存参考操作方法]及[Systemd限制CPU、内存参考操作方法]来看，相对来说Systemd更加简单，
而且目前已经被主流Linux发行版所支持（Red Hat系列、Debian系列等），而且经过几个版本的迭代已经很成熟了，
所以不管是Docker本身还是在K8S中建议使用Systemd来进行资源控制与管理。

