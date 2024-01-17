
# ebpf for mac
参考文档: https://github.com/singe/ebpf-docker-for-mac/tree/main

ubuntu/centos docker 安装：
```shell
# https://developer.aliyun.com/article/110806 docker 安装文档

# centos 安装 docker
yum install -y yum-utils device-mapper-persistent-data lvm2
yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
yum install -y docker-ce
systemctl enable docker && systemctl daemon-reload && systemctl start docker
systemctl status docker

# 登录 dockerhub
docker login -u lx1036 # 密码是 $OPlx19911010
cat /root/.docker/config.json


# ubuntu 安装 docker
sudo apt-get update -y
sudo apt-get install apt-transport-https ca-certificates curl software-properties-common lrzsz -y
sudo curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get update -y
sudo apt-get install docker-ce -y
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": ["https://gbmfgk59.mirror.aliyuncs.com"]
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker
```


mac 本地运行 ebpf bcc 示例:
```shell
docker run -it --name ebpf-for-mac --privileged -v /lib/modules:/lib/modules:ro -v /etc/localtime:/etc/localtime:ro --pid=host lx1036/ebpf-for-mac:2.0 /bin/bash

wget https://raw.githubusercontent.com/singe/ebpf-docker-for-mac/main/hello_world.py
python3 hello_world.py
```



