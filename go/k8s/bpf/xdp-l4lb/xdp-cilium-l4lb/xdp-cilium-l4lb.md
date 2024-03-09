
# docker in ubuntu
```shell
# https://docs.docker.com/engine/install/ubuntu/

# Add Docker's official GPG key:
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Add the repository to Apt sources:
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update -y

sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y

```


# Cilium Standalone L4LB
与 K8S 环境无关，Cilium 可以单独作为一个 XDP L4LB 部署，和 Katran 一样，部署代码：

```shell
# ubuntu-20 上验证没问题, 版本 v1.14.3 验证没问题。但是目前以代码 v1.10.20 版本为主!!!

# https://hub.docker.com/r/cilium/cilium
# v1.10.20 从 v1.10 版本开始生产可用 Standalone L4LB
docker pull cilium/cilium:v1.10.20

# 关闭 tso gso rx-checksum tx-checksum
# ethtool -K eth0 tso off gso off rx off tx off

docker rm l4lb

# 注意，--bpf-xxx 这些参数逐渐废弃，使用其他参数代替
docker run --cap-add NET_ADMIN --cap-add SYS_MODULE --cap-add CAP_SYS_ADMIN --network host --privileged \
-v /sys/fs/bpf:/sys/fs/bpf -v /lib/modules:/lib/modules -v /var/run/cilium:/var/run/cilium \
--name l4lb cilium/cilium:v1.10.20 cilium-agent \
--bpf-lb-algorithm=maglev \
--bpf-lb-mode=dsr \
#--bpf-lb-acceleration=native \ # ecs eth0 不支持 xdpdrv mode，只能 xdpgeneric mode
--bpf-lb-acceleration="testing-only" \
--bpf-lb-dsr-dispatch=ipip \
--devices=eth0 \
--datapath-mode=lb-only \
--enable-l7-proxy=false \
--tunnel=disabled \
--install-iptables-rules=false \
--enable-bandwidth-manager=false \
--enable-local-redirect-policy=false \
--enable-hubble=false \
--enable-l7-proxy=false \
--preallocate-bpf-maps=false \
--disable-envoy-version-check=true \
--auto-direct-node-routes=false \
--enable-ipv4=true \
--enable-ipv6=true \
--bpf-lb-map-max 512000

# https://www.ebpf.top/post/cilium-standalone-L4LB-XDP-zh/
# https://github.com/cilium/cilium-l4lb-test/blob/master/cilium-lb-example.yaml
# 配置 vip/rs
docker exec -it l4lb bash
cilium service update --id 1 --frontend "10.20.30.40:7047" \
--backends "10.30.30.41:7047,10.30.30.42:7047,10.30.30.43:7047" --k8s-node-port
cilium service list
```


## XDP modes
具体查看文档：https://docs.cilium.io/en/stable/bpf/toolchain/#iproute2
xdp 支持三种模式: xdpdrv(driver 模式)、xdpoffload 和 xdpgeneric:
```shell
# 比如 xdp driver 模式，但是在 ecs 里会报错不支持 xdp driver 模式
ip -force link set dev eth0 xdpdrv obj /var/run/cilium/state/bpf_xdp.o sec from-netdev

# Makefile
build:
    # https://docs.cilium.io/en/stable/bpf/toolchain/#llvm
	clang -target bpf -O2 -c xdp.c -I. -o xdp.o
	# clang -O2 -Wall --target=bpf -c xdp-example.c -o xdp-example.o
	# readelf -a xdp-example.o

attach:
	# load and attach, xdp driver mode
	ip -force link set dev eth0 xdpdrv obj /root/xdp/xdp-cilium-l4lb/bpf/xdp.o sec xdp
	#ip -force link set dev eth0 xdp obj /root/xdp/xdp-cilium-l4lb/bpf/xdp.o sec xdp # 顺序 xdpdrv->xdpgeneric

detach:
    ip -force link set dev eth0 xdp off
    ip -force link set dev eth0 xdpgeneric off

```

xdpdrv mode: 10G/40G 网卡一般都支持
xdpgeneric mode: 一般实验使用，运行在 xdp driver 后面 

对于 tc 的编译和加载过程:
```shell
clang -O2 -Wall --target=bpf -c tc-example.c -o tc-example.o
tc qdisc add dev em1 clsact
tc filter add dev em1 ingress bpf da obj tc-example.o sec ingress
tc filter add dev em1 egress bpf da obj tc-example.o sec egress
tc filter show dev em1 ingress
#filter protocol all pref 49152 bpf
#filter protocol all pref 49152 bpf handle 0x1 tc-example.o:[ingress] direct-action id 1 tag c5f7825e5dac396f
tc filter show dev em1 egress
#filter protocol all pref 49152 bpf
#filter protocol all pref 49152 bpf handle 0x1 tc-example.o:[egress] direct-action id 2 tag b2fd5adc0f262714

```


# xdpdump
https://github.com/cloudflare/xdpcap
https://blog.cloudflare.com/xdpcap/

安装 xdpdump:
```shell
# ubuntu-20 上验证通过

sudo apt-get install libpcap-dev -y
# go 1.20.10
go install github.com/cloudflare/xdpcap/cmd/xdpcap@latest
cp /root/go/bin/xdpcap /usr/local/bin/
# 没有成功，这里的 /sys/fs/bpf/l4lb 参数不是随便指定的
sudo xdpcap /sys/fs/bpf/l4lb - "tcp and port 80" | sudo tcpdump -r -
```
