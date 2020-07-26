
apt update && apt-get install -y apt-transport-https socat
curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
EOF
apt update

apt install -y kubelet kubeadm kubectl
#apt-mark hold kubelet kubeadm kubectl
