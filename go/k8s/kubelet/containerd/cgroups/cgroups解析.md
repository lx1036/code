





```shell
cd /sys/fs/cgroup/
ls cpu/kubepods/
```

```shell
# 查看挂载在 /sys/fs/cgroup 下的 cgroup subsystem controllers
mount -t cgroup
#cgroup on /sys/fs/cgroup/systemd type cgroup (rw,nosuid,nodev,noexec,relatime,xattr,release_agent=/usr/lib/systemd/systemd-cgroups-agent,name=systemd)
#cgroup on /sys/fs/cgroup/cpu,cpuacct type cgroup (rw,nosuid,nodev,noexec,relatime,cpu,cpuacct)
#cgroup on /sys/fs/cgroup/net_cls type cgroup (rw,nosuid,nodev,noexec,relatime,net_cls)
#cgroup on /sys/fs/cgroup/perf_event type cgroup (rw,nosuid,nodev,noexec,relatime,perf_event)
#cgroup on /sys/fs/cgroup/pids type cgroup (rw,nosuid,nodev,noexec,relatime,pids)
#cgroup on /sys/fs/cgroup/memory type cgroup (rw,nosuid,nodev,noexec,relatime,memory)
#cgroup on /sys/fs/cgroup/cpuset type cgroup (rw,nosuid,nodev,noexec,relatime,cpuset)
#cgroup on /sys/fs/cgroup/blkio type cgroup (rw,nosuid,nodev,noexec,relatime,blkio)
#cgroup on /sys/fs/cgroup/devices type cgroup (rw,nosuid,nodev,noexec,relatime,devices)
#cgroup on /sys/fs/cgroup/hugetlb type cgroup (rw,nosuid,nodev,noexec,relatime,hugetlb)
#cgroup on /sys/fs/cgroup/freezer type cgroup (rw,nosuid,nodev,noexec,relatime,freezer)


# 查看内核支持的 cgroup controllers
cat /proc/cgroups
##subsys_name	hierarchy	num_cgroups	enabled
#cpuset	7	36	1
#cpu	2	135	1
#cpuacct	2	135	1
#blkio	8	135	1
#memory	6	202	1
#devices	9	135	1
#freezer	11	32	1
#net_cls	3	32	1
#perf_event	4	32	1
#hugetlb	10	32	1
#pids	5	135	1


# 查看某个进程的cgroup情况
cat /sys/fs/cgroup/cpuset/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012/cgroup.procs
#28097
#28111
cat /proc/28111/cgroup # 第一列编号是 cgroup controller 的编号，比如 cpuset 编号为 7，与上面 `cat /proc/cgroups` 对应
#11:freezer:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#10:hugetlb:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#9:devices:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#8:blkio:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#7:cpuset:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#6:memory:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#5:pids:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#4:perf_event:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#3:net_cls:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#2:cpu,cpuacct:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#1:name=systemd:/kubepods/pod235148fe-393f-47a8-a17d-bd55bc1a836b/be27303c38cd3ee17ca4ee18c0772a42f1743b5682c30d617560878304342012
#0::/


# 可以通过挂载，把 cgroup 挂载到其他目录，看到的内容和 /sys/fs/cgroup/cpuset 内容一样
mkdir -p /tmp/cgroup/cpuset
mount -t cgroup -o cpuset none /tmp/cgroup/cpuset
ll /tmp/cgroup/cpuset # 看到的内容和 /sys/fs/cgroup/cpuset 内容一样

mount -t cgroup -o all cgroup /tmp/cgroup # 可以挂载所有的cgroup controllers，可以省略 `-o all`
umount /tmp/cgroup/cpuset # 需要卸载所有子目录


```





# cpuset controller 
**[cpuset](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html#cpuset)**

cpuset controller可以控制实现绑核，让pods独占cpu，免得cpu切换降低性能，这对cpu密集型pod很实用。
```shell
cat /sys/fs/cgroup/cpuset/cpuset.cpu # 查看当前cgroup中可以被task使用的cpus
# 0-31

```

## cgroup cli

```shell
# 安装 cgroups v2 cli
# https://github.com/containerd/cgroups/blob/master/cmd/cgctl/main.go
go get github.com/containerd/cgroups/cmd/cgctl

```

**如何判断宿主机 cgroups 是否使用 cgroup v2?**
答案：查看文件 /sys/fs/cgroup/cgroup.controllers 是否存在。内核版本至少 5.2 以上。
cgroup v2 强烈建议使用 systemd 作为 cgroup driver，而不是以前的 cgroupfs 。
目前我们的 k8s 1.19 用的是 4.19 kernel，用的还是 cgroups v1，driver 用的 cgroupfs。且 docker cgroup driver 也是用的 cgroupfs。 


## 参考文献

**[containerd cgroups go客户端库](https://github.com/containerd/cgroups)**










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

