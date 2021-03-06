

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
cgroups 是一种资源隔离方案，cgroups 提供了一些子系统：
cpu subsystem: 限制进程的cpu使用率
cpuacct(cpu account) subsystem: 统计 cgroups 中进程的cpu使用报告
cpuset subsystem: 为 cgroups 中的进程分配单独的 cpu/memory 节点
memory subsystem: 限制进程的 memory 使用量
blkio subsystem: 限制进程的块设备io
devices subsystem: 控制进程能够访问哪些设备
net_cls subsystem: 标记 cgroups 中进程的网络数据包，然后可以使用 tc(traffic control) 模块对数据包进行控制
freezer subsystem: 挂起或者恢复 cgroups 中的进程
ns subsystem: 不同 cgroups 下面的进程使用不同的 namespace



**[Cgroups中的CPU资源控制](https://mp.weixin.qq.com/s/O65oX2urY_zaADG22eg_Kw)**

## cpu subsystem
cpu子系统用于控制cgroup中所有进程可以使用的cpu时间片。
cpu subsystem主要涉及5接口: cpu.cfs_period_us，cpu.cfs_quota_us，cpu.shares，cpu.rt_period_us，cpu.rt_runtime_us.

* cpu.cfs_period_us: cfs_period_us表示一个cpu带宽，单位为微秒。系统总CPU带宽：cpu核心数 * cfs_period_us
* cpu.cfs_quota_us: cfs_quota_us表示Cgroup可以使用的cpu的带宽，单位为微秒
* cpu.shares: cpu.shares以相对比例限制cgroup的cpu
* cpu.rt_runtime_us: 以微秒（µs，这里以“us”代表）为单位指定在某个时间段中 cgroup 中的任务对 CPU 资源的最长连续访问时间
* cpu.rt_period_us: 以微秒（µs，这里以“us”代表）为单位指定在某个时间段中 cgroup 对 CPU 资源访问重新分配的频率


## docker cgroups 参数
见文档：https://docs.docker.com/config/containers/resource_constraints/#cpu
* --cpus=<value>: 比如2 cpus机器， --cpus="1.5"，等同于 --cpu-period="100000" --cpu-quota="150000"
* --cpu-period=<value>: CPU CFS scheduler period，默认是 100 ms。指定容器对cpu的使用在多长周期内重新分配cpu时间片
* --cpu-quota=<value>: CPU CFS quato 限额，指定单个周期内有多少时间来跑这个容器。--cpu-period和--cpu-quota配合使用
* --cpu-shared: 容器对宿主cpu的使用占比
* --cpuset-cpus: 绑定容器使用指定的宿主cpu，绑核，比如一台0-23 processor，cpuset.cpus为0-19表示当前容器在0-19 processor上运行
  不会在20-23 processor上运行。



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
**[Kubelet从入门到放弃:识透CPU管理](https://mp.weixin.qq.com/s/ViuaEIE0mEaWMJPCJm5-xg)**

## container level cgroups



## pod level cgroups



## qos level cgroups



## node level cgroups


