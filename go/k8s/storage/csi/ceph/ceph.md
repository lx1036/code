
# Ceph
**[ceph 官网](https://ceph.io/)**
**[ceph 介绍](https://www.qikqiak.com/k8strain/storage/ceph/)**


# Auth
**[cephfs开启client端认证](http://www.yangguanjun.com/2017/07/01/cephfs-client-authentication/)**









# Install

```
2.2 磁盘准备

需要在三台主机创建磁盘,并挂载到主机的/var/local/osd{0,1,2}

[root@master ~]# mkfs.xfs /dev/vdc

[root@master ~]# mkdir -p /var/local/osd0

[root@master ~]# mount /dev/vdc /var/local/osd0/

 

 

[root@node01 ~]# mkfs.xfs /dev/vdc

[root@node01 ~]# mkdir -p /var/local/osd1

[root@node01 ~]# mount /dev/vdc /var/local/osd1/

 

[root@node02 ~]# mkfs.xfs /dev/vdc 

[root@node02 ~]# mkdir -p /var/local/osd2

[root@node02 ~]# mount /dev/vdc /var/local/osd2/

 

将磁盘添加进入fstab中，确保开机自动挂载

 

2.3 配置各主机hosts文件

127.0.0.1 localhost localhost.localdomain localhost4 localhost4.localdomain4

::1 localhost localhost.localdomain localhost6 localhost6.localdomain6

172.16.60.2 master

172.16.60.3 node01

172.16.60.4 node02

 

2.4 管理节点ssh免密钥登录node1/node2

[root@master ~]# ssh-keygen -t rsa

[root@master ~]# ssh-copy-id -i /root/.ssh/id_rsa.pub root@node01

[root@master ~]# ssh-copy-id -i /root/.ssh/id_rsa.pub root@node02

 

2.5 master节点安装ceph-deploy工具

\# 各节点均更新ceph的yum源

vim /etc/yum.repos.d/ceph.repo 

 

[ceph]

name=ceph

baseurl=http://mirrors.aliyun.com/ceph/rpm-jewel/el7/x86_64/

gpgcheck=0

priority =1

[ceph-noarch]

name=cephnoarch

baseurl=http://mirrors.aliyun.com/ceph/rpm-jewel/el7/noarch/

gpgcheck=0

priority =1

[ceph-source]

name=Ceph source packages

baseurl=http://mirrors.aliyun.com/ceph/rpm-jewel/el7/SRPMS

gpgcheck=0

priority=1

 

\# 安装ceph-deploy工具

yum clean all && yum makecache

yum -y install ceph-deploy

 

2.6 创建monitor服务

创建monitor服务,指定master节点的hostname

[root@master ~]# mkdir /etc/ceph && cd /etc/ceph

[root@master ceph]# ceph-deploy new master

[root@master ceph]# ll

total 12

-rw-r--r-- 1 root root 195 Sep 3 10:56 ceph.conf

-rw-r--r-- 1 root root 2915 Sep 3 10:56 ceph-deploy-ceph.log

-rw------- 1 root root 73 Sep 3 10:56 ceph.mon.keyring

 

 

[root@master ceph]# cat ceph.conf 

[global]

fsid = 5b9eb8d2-1c12-4f6d-ae9c-85078795794b

mon_initial_members = master

mon_host = 172.16.60.2

auth_cluster_required = cephx

auth_service_required = cephx

auth_client_required = cephx

osd_pool_default_size = 2

 

配置文件的默认副本数从3改成2，这样只有两个osd也能达到active+clean状态，把下面这行加入到[global]段（可选配置）

 

2.7 所有节点安装ceph

\# 各节点安装软件包

yum -y install yum-plugin-priorities epel-release

\# master节点利用ceph-deply 部署ceph

 

[root@master ceph]# ceph-deploy install master node01 node02

 

[root@master ceph]# ceph --version

ceph version 10.2.11 (e4b061b47f07f583c92a050d9e84b1813a35671e)

 

2.8 部署相关服务

\# 安装ceph monitor

[root@master ceph]# ceph-deploy mon create master

 

\# 收集节点的keyring文件

[root@master ceph]# ceph-deploy gatherkeys master

 

\# 创建osd

[root@master ceph]# ceph-deploy osd prepare master:/var/local/osd0 node01:/var/local/osd1 node02:/var/local/osd2

 

\# 权限修改

[root@master ceph]# chmod 777 -R /var/local/osd{0..2}

[root@master ceph]# chmod 777 -R /var/local/osd{0..2}/*

 

\# 激活osd

[root@master ceph]# ceph-deploy osd activate master:/var/local/osd0 node01:/var/local/osd1 node02:/var/local/osd2

 

\# 查看状态

[root@master ceph]# ceph-deploy osd list master node01 node02

 

2.9 统一配置

用ceph-deploy把配置文件和admin密钥拷贝到所有节点，这样每次执行Ceph命令行时就无需指定monitor地址和ceph.client.admin.keyring了

[root@master ceph]# ceph-deploy admin master node01 node02

 

\# 各节点修改ceph.client.admin.keyring权限：

[root@master ceph]# chmod +r /etc/ceph/ceph.client.admin.keyring

 

 

\# 查看状态

[root@master ceph]# ceph health

HEALTH_OK

[root@master ceph]# ceph -s

cluster 5b9eb8d2-1c12-4f6d-ae9c-85078795794b

health HEALTH_OK

monmap e1: 1 mons at {master=172.16.60.2:6789/0}

election epoch 3, quorum 0 master

osdmap e15: 3 osds: 3 up, 3 in

flags sortbitwise,require_jewel_osds

pgmap v27: 64 pgs, 1 pools, 0 bytes data, 0 objects

15681 MB used, 1483 GB / 1499 GB avail

64 active+clean

 

2.10 部署MDS服务

我们在node01/node02上安装部署MDS服务

[root@master ceph]# ceph-deploy mds create node01 node02

 

\# 查看状态

[root@master ceph]# ceph mds stat

e3:, 2 up:standby

[root@master ~]# ceph mon stat

e1: 1 mons at {master=172.16.60.2:6789/0}, election epoch 4, quorum 0 master

 

\# 查看服务

[root@master ceph]# systemctl list-unit-files |grep ceph

ceph-create-keys@.service static 

ceph-disk@.service static 

ceph-mds@.service disabled

ceph-mon@.service enabled 

ceph-osd@.service enabled 

ceph-radosgw@.service disabled

ceph-mds.target enabled 

ceph-mon.target enabled 

ceph-osd.target enabled 

ceph-radosgw.target enabled 

ceph.target enabled 

 

至此，基本上完成了ceph存储集群的搭建。

三 创建ceph文件系统

3.1 创建文件系统

关于创建存储池

确定 pg_num 取值是强制性的，因为不能自动计算。下面是几个常用的值：

少于 5 个 OSD 时可把 pg_num 设置为 128

OSD 数量在 5 到 10 个时，可把 pg_num 设置为 512

OSD 数量在 10 到 50 个时，可把 pg_num 设置为 4096

OSD 数量大于 50 时，你得理解权衡方法、以及如何自己计算 pg_num 取值

自己计算 pg_num 取值时可借助 pgcalc 工具

　　随着 OSD 数量的增加，正确的 pg_num 取值变得更加重要，因为它显著地影响着集群的行为、以及出错时的数据持久性（即灾难性事件导致数据丢失的概率）。

[root@master ceph]# ceph osd pool create cephfs_data <pg_num> 

[root@master ceph]# ceph osd pool create cephfs_metadata <pg_num>

 

[root@master ~]# ceph osd pool ls 

rbd

[root@master ~]# ceph osd pool create kube 128

pool 'kube' created

[root@master ~]# ceph osd pool ls 

rbd

kube

 

 

创建文件系统

ceph fs new <fs_name> <metadata> <data>

例如：

$ ceph fs new cephfs cephfs_metadata cephfs_data

$ ceph fs ls

name: cephfs, metadata pool: cephfs_metadata, data pools: [cephfs_data ]

文件系统创建完毕后， MDS 服务器就能达到 *active* 状态了，比如在一个单 MDS 系统中：



 $ ceph mds stat

e5: 1/1/1 up {0=a=up:active}

建好文件系统且 MDS 活跃后，你就可以挂载此文件系统了：

 

 

 

\# 查看证书

[root@master ~]# ceph auth list

installed auth entries:

 

mds.node01

key: AQB56m1dE42rOBAA0yRhsmQb3QMEaTsQ71jHdg==

caps: [mds] allow

caps: [mon] allow profile mds

caps: [osd] allow rwx

mds.node02

key: AQB66m1dWuhWKhAAtbiZN7amGcjUh6Rj/HNFkg==

caps: [mds] allow

caps: [mon] allow profile mds

caps: [osd] allow rwx

osd.0

key: AQA46W1daFx3IxAAE1esQW+t1fWJDfEQd+167w==

caps: [mon] allow profile osd

caps: [osd] allow *

osd.1

key: AQBA6W1daJG9IxAAQwETgrVc3awkEZejDSaaow==

caps: [mon] allow profile osd

caps: [osd] allow *

osd.2

key: AQBI6W1dot4/GxAAle3Ii3/D38RmwNC4yTCoPg==

caps: [mon] allow profile osd

caps: [osd] allow *

client.admin

key: AQBu4W1d90dZKxAAH/kta03cP5znnCcWeOngzQ==

caps: [mds] allow *

caps: [mon] allow *

caps: [osd] allow *

client.bootstrap-mds

key: AQBv4W1djJ1uHhAACzBcXjVoZFgLg3lN+KEv8Q==

caps: [mon] allow profile bootstrap-mds

client.bootstrap-mgr

key: AQCS4W1dna9COBAAiWPu7uk3ItJxisVIwn2duA==

caps: [mon] allow profile bootstrap-mgr

client.bootstrap-osd

key: AQBu4W1dxappOhAA5FanGhQhAOUlizqa5uMG3A==

caps: [mon] allow profile bootstrap-osd

client.bootstrap-rgw

key: AQBv4W1dpwvsDhAAyp58v08XttJWzLoHWVHZow==

caps: [mon] allow profile bootstrap-rgw

 

3.2 创建客户端密钥（可选）

# 创建keyring

[root@master ~]# ceph auth get-or-create client.kube mon 'allow r' osd 'allow rwx pool=kube' -o /etc/ceph/ceph.client.kube.keyring

[root@master ~]# ceph auth list

 

# 将密钥拷贝到node1和node2

[root@master ceph]# scp ceph.client.kube.keyring root@node01:/etc/ceph/

 

2. 创建挂载点

[root@client ceph]# mkdir /mnt/cephfs

 

3. 创建密钥文件

3.1 在admin节点查看密钥内容

[root@admin ceph]# cat ceph.client.admin.keyring

[client.admin]

key = AQBLR7xdckUIFhAA2G05Hidq5aoSse0nxGNdJQ==

 

3.2 在client节点创建密钥文件

[root@client ceph]# vim /etc/ceph/admin.secret #复制上面文件中的内容,注意不是全部内容

AQBLR7xdckUIFhAA2G05Hidq5aoSse0nxGNdJQ==

 

4. 挂载

[root@client ceph]# mount.ceph 192.168.2.163:6789:/ /ceph/ -o name=admin,secretfile=/etc/ceph/ceph.client.admin.keyring

四 卸载

清理机器上的ceph相关配置：

停止所有进程： stop ceph-all

卸载所有ceph程序：ceph-deploy uninstall [{ceph-node}]

删除ceph相关的安装包：ceph-deploy purge {ceph-node} [{ceph-data}]

删除ceph相关的配置：ceph-deploy purgedata {ceph-node} [{ceph-data}]

删除key：ceph-deploy forgetkeys

 

卸载ceph-deploy管理：yum -y remove ceph-deploy

 
```

 

