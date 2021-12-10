


## VFS BPF
vfsstat: https://github.com/iovisor/bcc/blob/master/tools/vfsstat_example.txt
http://manpages.ubuntu.com/manpages/bionic/man8/vfsstat-bpfcc.8.html
vfscount: https://github.com/iovisor/bcc/blob/master/tools/vfscount_example.txt
http://manpages.ubuntu.com/manpages/bionic/man8/vfscount-bpfcc.8.html
使用库 https://github.com/iovisor/gobpf 来写 bpf

lmp: https://github.com/linuxkerneltravel/lmp/blob/master/plugins/fs/vfsstat.py


## 常见问题
(1) 在 centos 机器上安装 bcc
```shell
yum update && yum install -y bcc-tools bcc bcc-devel

# 安装好 bcc 和 bcc-tools 后，可以运行相关 tools 工具
echo 'export PATH="$PATH:/usr/share/bcc/tools/"' >> /etc/profile
source /etc/profile
vfsstat # https://github.com/iovisor/bcc/blob/master/tools/vfsstat_example.txt
```
