
# k8s mirror
apt-get update && apt-get install -y apt-transport-https
curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
EOF
apt-get update
apt-get install -y kubelet kubeadm kubectl

# docker-ce mirror
curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
# or
sudo apt-get update
sudo apt-get -y install apt-transport-https ca-certificates curl software-properties-common
# step 2: 安装GPG证书
curl -fsSL http://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | sudo apt-key add -
# Step 3: 写入软件源信息
sudo add-apt-repository "deb [arch=amd64] http://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable"
# Step 4: 更新并安装 Docker-CE
sudo apt-get -y update
sudo apt-get -y install docker-ce

# 关闭 swap
swapoff -a
vi /etc/fstab # 注释掉swap


# 使用阿里镜像
# https://www.latelee.org/kubernetes/k8s-deploy-1.17.0-detail.html
#!/bin/bash
# 下面的镜像应该去除"k8s.gcr.io/"的前缀，版本换成kubeadm config images list命令获取到的版本
images=(
    kube-apiserver:v1.17.0
    kube-controller-manager:v1.17.0
    kube-scheduler:v1.17.0
    kube-proxy:v1.17.0
    pause:3.1
    etcd:3.4.3-0
    coredns:1.6.5
)
# shellcheck disable=SC2068
for imageName in ${images[@]} ; do
    docker pull registry.cn-hangzhou.aliyuncs.com/google_containers/$imageName
    docker tag registry.cn-hangzhou.aliyuncs.com/google_containers/$imageName k8s.gcr.io/$imageName
    docker rmi registry.cn-hangzhou.aliyuncs.com/google_containers/$imageName
done
