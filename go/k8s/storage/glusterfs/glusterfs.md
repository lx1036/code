
# Install
install: https://docs.gluster.org/en/latest/Quick-Start-Guide/Quickstart/

CentOS:
(1) 准备5个CentOS 7.6 64位虚拟机，系统盘20G，数据盘50G。3个组一个glusterfs集群，另两个作为中转机

(2)每个节点安装
```shell script
# 在3个节点上依次执行
yum search centos-release-gluster
# 选centos-release-gluster7.noarch这个版本
yum update && yum install centos-release-gluster7.noarch -y
systemctl start glusterd && systemctl status glusterd

# 在3个节点上依次执行
# 互联各个节点，<ip-address>为peer节点的ip
iptables -I INPUT -p all -s <ip-address> -j ACCEPT

# 检测集群安装是否成功，<ip-address>为peer节点的ip
gluster peer probe <ip-address>
gluster peer status
```

(3)创建一个volume
```shell script
# 在3个节点上依次执行
# 设置volume在glusterfs集群上的路径
mkdir -p /data/brick1/gv0
# 创建一个gv0 volume
gluster volume create gv0 replica 3 192.168.0.227:/data/brick1/gv0  192.168.0.134:/data/brick1/gv0 192.168.0.148:/data/brick1/gv0
# 启动gv0 volume
gluster volume start gv0
# 检查gv0状态
gluster volume info
```

(4)验证
分别登录两个中转机：
```shell script
yum search glusterfs
# 选择glusterfs-fuse.x86_64 : Fuse client
yum update && yum install glusterfs-fuse.x86_64 -y

# client1 里执行，这里192.168.0.227为glusterfs 3个任意一个：
mkdir -p /mnt/data
mount -t glusterfs 192.168.0.227:/gv0 /mnt/data

# client2 里执行

mkdir -p /mnt/data2
mount -t glusterfs 192.168.0.227:/gv0 /mnt/data2
```

分别在client1/client2创建两个文件index1.txt和index2.txt:
```shell script
# client1里执行
echo "hello" >> /mnt/data/index1.txt

# client2里执行
echo "world" >> /mnt/data2/index2.txt
```
这时client1和client2里会分别有这两个文件。
同时，登录3台glusterfs集群任何一台机器，在 /data/brick1/gv0 目录内也会存在这两个文件。

# Docs
docs: http://docs.gluster.org/



