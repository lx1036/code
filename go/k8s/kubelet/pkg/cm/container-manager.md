

# container-manager

kubelet涉及container-manager相关参数：
* --kubelet-cgroups(KubeletCgroups): Optional absolute name of cgroups to create and run the Kubelet in.
* --system-reserved-cgroup(SystemReservedCgroup): 系统cgroup name，默认 ''
* --kube-reserved-cgroup(KubeReservedCgroup): k8s组件cgroup name，默认 ''
* --cgroup-root(CgroupRoot): 默认""，pods的root cgroup
* --cgroup-driver(CgroupDriver): kubelet用来管理cgroups的driver, 有cgroupfs/systemd, 默认是cgroupfs 
* --cgroups-per-qos(CgroupsPerQOS): Burstable/BestEffort pods 都在其对应的cgroup目录下: /sys/fs/cgroup/cpuset/kubepods/besteffort
和 /sys/fs/cgroup/cpuset/kubepods/burstable，该参数开启 pod level cgroups


# cgroups(control groups)
CGroup是一种资源隔离方案，CGroup提供了一些子系统：
cpu: 可以限制使用CPU权限。
memory: 可以限制内存使用量
blockio: 限制I/O速率
cpuset: 基于CPU核心进行限制(限制可以使用哪些核)
cpuacct: 记录进程组使用的资源数量(CPU时间)
device: 允许或拒绝CGroup中的任务访问设备
frezzer: 可以控制进程挂起或者恢复(挂起后可以释放资源)
ns: 名称空间子系统
net_cls: 可用于标记网络数据包，它不直接控制网络读写。
net_proi: 可用于设置网络设备的优先级


**[Cgroups中的CPU资源控制](https://mp.weixin.qq.com/s/O65oX2urY_zaADG22eg_Kw)**

## cpu subsystem
cpu子系统用于控制cgroup中所有进程可以使用的cpu时间片。
cpu subsystem主要涉及5接口: cpu.cfs_period_us，cpu.cfs_quota_us，cpu.shares，cpu.rt_period_us，cpu.rt_runtime_us.

* cpu.cfs_period_us: cfs_period_us表示一个cpu带宽，单位为微秒。系统总CPU带宽：cpu核心数 * cfs_period_us
* cpu.cfs_quota_us: cfs_quota_us表示Cgroup可以使用的cpu的带宽，单位为微秒
* cpu.shares: cpu.shares以相对比例限制cgroup的cpu
* cpu.rt_runtime_us: 以微秒（µs，这里以“us”代表）为单位指定在某个时间段中 cgroup 中的任务对 CPU 资源的最长连续访问时间
* cpu.rt_period_us: 以微秒（µs，这里以“us”代表）为单位指定在某个时间段中 cgroup 对 CPU 资源访问重新分配的频率

## cpuacct subsystem
cpuacct子系统（CPU accounting）会自动生成报告来显示cgroup中任务所使用的CPU资源。报告有两大类：cpuacct.stat和cpuacct.usage。

* cpuacct.stat: cpuacct.stat记录cgroup的所有任务（包括其子孙层级中的所有任务）使用的用户和系统CPU时间
```shell
cat cpuacct.stat
# user 78 #用户模式中任务使用的CPU时间
# system 447 #系统模式(内核)中任务使用的CPU时间
```

* cpuacct.usage: cpuacct.usage记录这个cgroup中所有任务（包括其子孙层级中的所有任务）消耗的总CPU时间（纳秒）
* cpuacct.usage_percpu: cpuacct.usage_percpu记录这个cgroup中所有任务（包括其子孙层级中的所有任务）在每个CPU中消耗的CPU时间（以纳秒为单位)

## cpuset subsystem
cpuset主要是为了numa使用的，numa技术将CPU划分成不同的node，每个node由多个CPU组成，
并且有独立的本地内存、I/O等资源(硬件上保证)。可以使用numactl查看当前系统的node信息。

* cpuset.cpus: cpuset.cpus指定允许这个 cgroup 中任务访问的 CPU。
  这是一个用逗号分开的列表，格式为 ASCII，使用小横线（"-"）代表范围。如下，代表 CPU 0、1、2 和 16
> 这个才是最重要的一个指标
```shell
cd /sys/fs/cgroup/cpuset/kubepods/pod24caaeb4-610e-4477-bbc7-a8110b25a513/05cad2ecf9ece3761c1433d03331a7f49b3353f30a9d2d15dcd7999244902d49
cat cpuset.cpus
# 3,15
```
* cpuset.mems: cpuset.mems指定允许这个 cgroup 中任务可访问的内存节点。
  这是一个用逗号分开的列表，格式为 ASCII，使用小横线（"-"）代表范围。如下代表内存节点 0、1、2 和 16。
```shell
cat cpuset.mems
#0-1
```

使用numactl工具查看机器numa节点分布：
```shell
yum install -y numactl
numactl -H
#available: 2 nodes (0-1)
#node 0 cpus: 0 1 2 3 4 5 12 13 14 15 16 17
#node 0 size: 32002 MB
#node 0 free: 23407 MB
#node 1 cpus: 6 7 8 9 10 11 18 19 20 21 22 23
#node 1 size: 32253 MB
#node 1 free: 27065 MB
#node distances:
#node   0   1
#  0:  10  21
#  1:  21  10
```


# kubelet 使用 cgroups
**[kubernetes kubelet组件中cgroup的层层"戒备"](https://www.cnblogs.com/gaorong/p/11716907.html)**

## container level cgroups



## pod level cgroups



## qos level cgroups



## node level cgroups


