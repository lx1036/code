
# basic
yum update && yum install -y yum-utils lvm2 device-mapper-persistent-data nfs-utils xfsprogs wget
swapoff -a

cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sysctl --system

# docker
yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
yum update && yum -y install docker-ce docker-ce-cli containerd.io
systemctl enable docker
systemctl start docker
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
systemctl enable docker
systemctl start docker
systemctl status docker

# kubelet
cat > /etc/yum.repos.d/kubernetes.repo <<EOF
[kubernetes]
name=Kubernetes
baseurl=http://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=http://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg http://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
# 删除旧版本
yum -y remove kubelet kubadm kubctl
yum update && yum install -y kubelet kubeadm kubectl
systemctl daemon-reload
systemctl enable kubelet
systemctl start kubelet
systemctl status kubelet
