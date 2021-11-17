

tc qdisc del dev eth0 root
tc qdisc add dev eth0 root handle 1: htb default 1
tc class add dev eth0 parent 1: classid 1:1 htb rate 1Gbit
tc class add dev eth0 parent 1: classid 1:2 htb rate 500Mbit # 总共占用 100Mbit 宽带
tc class add dev eth0 parent 1:2 classid 1:3 htb rate 250Mbit ceil 500Mbit prio 3
tc class add dev eth0 parent 1:2 classid 1:5 htb rate 130Mbit ceil 500Mbit prio 5
tc class add dev eth0 parent 1:2 classid 1:7 htb rate 120Mbit ceil 500Mbit prio 7
tc filter add dev eth0 parent 1: protocol ip handle 1: cgroup

#mkdir /sys/fs/cgroup/net_cls/high
#mkdir /sys/fs/cgroup/net_cls/low
echo 0x10003 > /sys/fs/cgroup/net_cls/high/net_cls.classid
echo 0x10007 > /sys/fs/cgroup/net_cls/low/net_cls.classid

# 查看 eth0 网卡宽带
ethtool eth0

# echo $$ > /sys/fs/cgroup/net_cls/low/cgroup.procs
# echo $$ > /sys/fs/cgroup/net_cls/high/cgroup.procs
# iperf3 -c 10.208.40.96 -p 5000 --bandwidth 10G -t 1000
# iperf3 -c 10.208.40.96 -p 5001 --bandwidth 10G -t 1000
