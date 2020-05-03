

vip=117.50.107.43
rs1=192.168.1.142
rs2=192.168.1.37

sudo ipvsadm -C
sudo ipvsadm -A -t $vip:80 -s rr
sudo ipvsadm -a -t $vip:80 -r $rs1:80 -g
sudo ipvsadm -a -t $vip:80 -r $rs2:80 -g
