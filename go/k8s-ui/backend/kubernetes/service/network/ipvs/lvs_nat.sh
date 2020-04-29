echo 1 > /proc/sys/net/ipv4/ip_forward

vip=117.50.107.43
rs1=192.168.1.142
rs2=192.168.1.37

sudo ipvsadm -C
sudo ipvsadm -A -t $vip:80 -s rr
sudo ipvsadm -a -t $vip:80 -r $rs1:80 -m
sudo ipvsadm -a -t $vip:80 -r $rs2:80 -m

# ipvsadm -ln

# 在每一个rs上执行：修改 rs 的默认网关为 vip
route add default gw $vip dev eth0

# 查看当前linux主机的默认网关
ip route show
