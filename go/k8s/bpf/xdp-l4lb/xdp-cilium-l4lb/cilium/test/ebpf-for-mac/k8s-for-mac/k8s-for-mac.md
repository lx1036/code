



# 1. 抓包 worker-eth0-udp

* 在 worker 上抓包 worker-eth0-udp.pcap
* 在 master 上 curl nginx pod ip

```shell
minikube ssh m02
sudo tcpdump -i eth0 -nneevv -A udp -w worker-eth0-udp.pcap

curl 10.244.1.3
```


# 2. 抓包 worker-cni0-tcp

* 在 worker 上抓包 worker-cni0-tcp.pcap
* 在 master 上 curl nginx pod ip

```shell
minikube ssh m02
sudo tcpdump -i cni0 -nneevv -A tcp -w worker-cni0-tcp.pcap

curl 10.244.1.3

minikube cp minikube-m02:/home/docker/worker-cni0-tcp.pcap worker-cni0-tcp.pcap
```



