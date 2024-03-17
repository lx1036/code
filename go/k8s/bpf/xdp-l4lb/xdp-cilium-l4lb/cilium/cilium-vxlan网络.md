
# pod in cilium vxlan

加载的 bpf 程序:
```shell
# cilium 1.12.3
minikube start --cni=cilium --driver=docker --kubernetes-version=v1.28.3 --force --listen-address=0.0.0.0
minikube node add --worker=true
kubectl create deploy my-nginx --image=nginx:1.24.0 --replicas=3
kubectl expose deployment my-nginx --port=8080 --target-port=80 --type=NodePort --name=nginx-svc

docker@minikube-m02:~$ tc filter show dev lxc81bb9e5d5ed7@if9 ingress
filter protocol all pref 1 bpf chain 0 
filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_lxc.o:[from-container] direct-action not_in_hw id 2526 tag ed63eca2d31b5035 
docker@minikube-m02:~$ tc filter show dev lxc81bb9e5d5ed7@if9 egress

root@minikube-m02:/home/cilium# tc filter show dev cilium_net ingress
filter protocol all pref 1 bpf chain 0 
filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_host_cilium_net.o:[to-host] direct-action not_in_hw id 2478 tag 6b89d4d09c799b6f jited 
root@minikube-m02:/home/cilium# tc filter show dev cilium_net egress
root@minikube-m02:/home/cilium# tc filter show dev cilium_host ingress
filter protocol all pref 1 bpf chain 0 
filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_host.o:[to-host] direct-action not_in_hw id 2458 tag 6b89d4d09c799b6f jited 
root@minikube-m02:/home/cilium# tc filter show dev cilium_host egress
filter protocol all pref 1 bpf chain 0 
filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_host.o:[from-host] direct-action not_in_hw id 2468 tag 1c9907f1d5ea3240 jited 
root@minikube-m02:/home/cilium# 

docker@minikube-m02:~$ tc filter show dev cilium_vxlan ingress
filter protocol all pref 1 bpf chain 0 
filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_overlay.o:[from-overlay] direct-action not_in_hw id 2386 tag 8dadd616a2c190d7 
docker@minikube-m02:~$ tc filter show dev cilium_vxlan egress
filter protocol all pref 1 bpf chain 0 
filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_overlay.o:[to-overlay] direct-action not_in_hw id 2395 tag 4f2c6c27fb1b3d28 

docker@minikube-m02:~$ tc filter show dev eth0 ingress
filter protocol all pref 1 bpf chain 0 
filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_netdev_eth0.o:[from-netdev] direct-action not_in_hw id 2488 tag ad1249d16f715492 
docker@minikube-m02:~$ tc filter show dev eth0 egress
filter protocol all pref 1 bpf chain 0 
filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_netdev_eth0.o:[to-netdev] direct-action not_in_hw id 2498 tag ce4cbdcee6c055c2 

```




