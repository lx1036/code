



hostnamectl --static set-hostname  node-1
systemctl disable firewalld.service
systemctl stop firewalld.service


yum install -y docker
systemctl start docker

echo '{"debug":true,"registry-mirrors":["https://wvtedxym.mirror.aliyuncs.com"],"experimental":true}' > /etc/docker/daemon.json
docker pull lx1036/ubuntu:1.0.1
docker images


systemctl stop docker
ip link set dev docker0 down
brctl delbr docker0
brctl addbr bridge0
ip addr add 192.168.10.1/24 dev bridge0
ip link set dev bridge0 up
ip addr show bridge0
vim /etc/sysconfig/docker
systemctl restart docker
ifconfig
ip link set dev docker0 down && brctl delbr docker0 && systemctl restart docker

# etcd
rsync etcd-v3.4.10-linux-amd64.tar.gz root@106.75.60.229:/root/
rsync etcd-v3.4.10-linux-amd64.tar.gz root@106.75.73.132:/root/
rsync etcd-v3.4.10-linux-amd64.tar.gz root@106.75.105.253:/root/


wget https://github.com/etcd-io/etcd/releases/download/v3.4.10/etcd-v3.4.10-linux-amd64.tar.gz
tar zxf etcd-v3.4.10-linux-amd64.tar.gz
cd etcd-v3.4.10-linux-amd64/ || exit




./etcd --name node1 --initial-advertise-peer-urls http://192.168.64.40:2380 --listen-peer-urls http://0.0.0.0:2380 --listen-client-urls http://0.0.0.0:2379,http://127.0.0.1:4001 --advertise-client-urls http://0.0.0.0:2379 --initial-cluster-token etcd-cluster --initial-cluster node1=http://192.168.64.40:2380,node2=http://192.168.64.41:2380,node3=http://192.168.64.42:2380 --initial-cluster-state new

./etcd --name node2 --initial-advertise-peer-urls http://192.168.64.41:2380 --listen-peer-urls http://0.0.0.0:2380 --listen-client-urls http://0.0.0.0:2379,http://127.0.0.1:4001 --advertise-client-urls http://0.0.0.0:2379 --initial-cluster-token etcd-cluster --initial-cluster node1=http://192.168.64.40:2380,node2=http://192.168.64.41:2380,node3=http://192.168.64.42:2380 --initial-cluster-state new

./etcd --name node3 --initial-advertise-peer-urls http://192.168.64.42:2380 --listen-peer-urls http://0.0.0.0:2380 --listen-client-urls http://0.0.0.0:2379,http://127.0.0.1:4001 --advertise-client-urls http://0.0.0.0:2379 --initial-cluster-token etcd-cluster --initial-cluster node1=http://192.168.64.40:2380,node2=http://192.168.64.41:2380,node3=http://192.168.64.42:2380 --initial-cluster-state new

./etcdctl  member list

ip link add veth0 type veth peer name eth1
ip netns add ns0
ip link set eth1 netns ns0
ip netns exec ns0 ip a add 10.20.1.2/24 dev eth1
ip netns exec ns0 ip link set eth1 up
ip netns exec ns0 ip route add 169.254.1.1 dev eth1 scope link
ip netns exec ns0 ip route add default via 169.254.1.1 dev eth1
ip link set veth0 up
ip route add 10.20.1.2 dev veth0 scope link
ip route add 10.20.1.3 via 192.168.1.177 dev ens192
echo 1 > /proc/sys/net/ipv4/conf/veth0/proxy_arp



docker run --net=host --privileged \
  --name=calico-node -d --restart=always -e ETCD_DISCOVERY_SRV= -e NODENAME=node-1 \
  -e CALICO_NETWORKING_BACKEND=bird -e IP=192.168.64.40 \
  -e ETCD_ENDPOINTS=http://10.9.27.23:2379,http://10.9.48.153:2379,http://10.9.33.22:2379 \
  -v /var/log/calico:/var/log/calico -v /var/run/calico:/var/run/calico \
  -v /var/lib/calico:/var/lib/calico -v /lib/modules:/lib/modules -v /run:/run \
  quay.io/calico/node:latest
