
参考文献: https://developer.aliyun.com/article/763983

# kubeadm 安装 K8s
准备：
* CentOS 7
* 每台机器 2 GB 或更多的 RAM (如果少于这个数字将会影响您应用的运行内存)
* 2 CPU 核或更多
* 集群中的所有机器的网络彼此均能相互连接(公网和内网都可以)
* 禁用交换分区。为了保证 kubelet 正常工作，必须 禁用交换分区。

1. 安装 docker kubelet kubeadm
```shell script
yum update && yum install -y yum-utils lvm2 device-mapper-persistent-data nfs-utils xfsprogs wget

# 关闭swap，防止虚拟内存存在，把一部分内存放到硬盘上，k8s要求的，主要是为了防止如读取pod内存时数据不准
swapoff -a

# 由于 iptables 被绕过而导致流量无法正确路由的问题。
# 应该确保 在 sysctl 配置中的 net.bridge.bridge-nf-call-iptables 被设置为 1
cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system

# 安装docker使用阿里云源，docker镜像拉取使用源https://wvtedxym.mirror.aliyuncs.com，可以使用自己的阿里云源
# 修改docker cgroup driver为systemd，如果不修改则在后续添加worker节点时会遇到"detected cgroupfs as ths Docker driver.xx"的报错信息
wget -O /etc/yum.repos.d/CentOS-Base.repo http://mirrors.aliyun.com/repo/Centos-7.repo
sed -i -e '/mirrors.cloud.aliyuncs.com/d' '/mirrors.aliyuncs.com/d' /etc/yum.repos.d/CentOS-Base.repo
yum makecache
yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
yum update && yum install -y docker-ce docker-ce-cli containerd.io
systemctl enable docker
systemctl start docker
# 设置docker cgroup driver为systemd，k8s需要
cat > /etc/docker/daemon.json <<EOF
{
	"exec-opts": [
		"native.cgroupdriver=systemd"
	],
	"registry-mirrors":[
		"https://wvtedxym.mirror.aliyuncs.com"
	]
}
EOF
systemctl daemon-reload
systemctl restart docker
systemctl status docker
```

```shell script
# k8s 阿里云源
cat > /etc/yum.repos.d/kubernetes.repo <<EOF
[kubernetes]
name=Kubernetes
baseurl=http://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=http://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg http://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
yum update && yum -y install kubelet kubeadm kubectl

# 二进制启动docker kubelet
systemctl daemon-reload
systemctl enable kubelet
systemctl start kubelet

# kubelet 现在每隔几秒就会重启，因为它陷入了一个等待 kubeadm 指令的死循环
# systemctl status kubelet
```

3. master节点部署
```shell script
# 在master节点执行，pod网路段选择192.168.0.0/16
kubeadm init \
#    --apiserver-advertise-address 0.0.0.0 \
#    --apiserver-bind-port 6443 \
    --cert-dir /etc/kubernetes/pki \
    --image-repository registry.cn-hangzhou.aliyuncs.com/google_containers \
    --kubernetes-version 1.18.2 \
    --pod-network-cidr 192.168.0.0/16 \
#    --control-plane-endpoint ${master_ip} \
    --upload-certs

# 配置kubectl
rm -rf /root/.kube && mkdir /root/.kube
cp -i /etc/kubernetes/admin.conf .kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
kubectl get nodes
kubectl get pod -A
```

(可选)如果需要增加master节点，可在新的一台服务器上执行如下类似命令，作为k8s里的master节点(当然，还需要先执行(1)和(2)两个基本操作)：
```shell script
kubeadm join ip --token ${token} --discovery-token-ca-cert-hash ${ca-hash} --control-plane --certificate-key ${ca-key}
```

```shell script
# kubeadm join 命令可重新生成：
kubeadm token create --print-join-command
```

4. 以daemonset形式安装calico
```shell script
# 由于以上pod网路段选择192.168.0.0/16，所以不需要修改calico.yaml里的pod网络段配置
# 否则如果选择别的网络段如10.10.0.0/16，则需要改一下calico.yaml里的pod网络段，如：
# `sed -i "s#192\.168\.0\.0/16#10\.10\.0\.0/16#" calico.yaml`
kubectl apply -f https://docs.projectcalico.org/v3.8/manifests/calico.yaml

# 验证，过一会master节点应该是ready状态
kubectl get nodes
# 看看是不是所有pod都已经running了
kubectl get pods -A --watch
```

5. worker节点部署
5.1 按照(1)和(2)安装基本软件。
5.2 把该worker节点加入k8s集群
```shell script
# 执行第三步生成的命令，一般类似如下：
kubeadm join ip --token ${token} --discovery-token-ca-cert-hash ${ca-hash}

kubectl label node ${node1} ${node2} node-role.kubernetes.io/node=""
```

6. 测试

```shell script
kubectl apply -f https://github.com/lx1036/code/blob/master/go/k8s/network/nginx/minikube-nginx.yml
```
