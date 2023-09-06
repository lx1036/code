

# runc 命令安装

```shell
#ubuntu
apt install -y libseccomp-dev pkg-config
git clone https://github.com/opencontainers/runc
cd runc
make && make install
ls /usr/local/sbin/runc
runc --help
```


# runc 基本使用
使用说明：https://github.com/opencontainers/runc

```shell

mkdir -p mycontainer/rootfs && cd mycontainer
docker export $(docker create busybox) | tar -C rootfs -xvf - # 把 busybox:latest 镜像文件放入 rootfs 目录下

mkdir nginx
cd nginx
mkdir rootfs
docker export $(docker create nginx:1.17.8) | tar -C rootfs -xvf -
runc spec
# config.json 修改
"args": [
  "nginx", "-g", "daemon off;", "-p", "/tmp"
],
runc run nginx1
```

```shell
mkdir hello
cd hello
docker pull hello-world
docker export $(docker create hello-world) > hello-world.tar
mkdir rootfs
tar -C rootfs -xf hello-world.tar
runc spec
sed -i 's;"sh";"/hello";' config.json
runc run container1
runc list

# 以下没有验证
sed -i 's;"/hello";"sleep 3600";' config.json
cat /proc/$(pidof container1)/status|grep Cap
runc exec container1 ls
```
