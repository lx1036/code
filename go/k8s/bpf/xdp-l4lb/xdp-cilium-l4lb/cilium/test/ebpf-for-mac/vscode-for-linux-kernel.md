



# vscode 阅读 linux 源码
https://luckymrwang.github.io/2022/09/09/VSCode%E9%98%85%E8%AF%BBLinux%E6%BA%90%E7%A0%81/

https://zhuanlan.zhihu.com/p/353592340

可以解决 C 函数跳转，以及头文件 .h 报错问题，阅读 linux 源码的确顺畅很多！！！

(1)搞一台 ubuntu ecs，不要 centos 装不了 global
```shell
wget https://mirrors.edge.kernel.org/pub/linux/kernel/v5.x/linux-5.10.142.tar.gz

#或者用阿里云镜像:
wget https://mirrors.aliyun.com/linux-kernel/v5.x/linux-5.10.142.tar.gz
tar -xzf linux-5.10.142.tar.gz
sudo apt-get update && apt-get install -y global
```

(2)vscode 安装 global 插件
VSCode 上有现成的插件可以直接使用，我们在 VSCode 这个 SSH 会话里安装 C/C++ GNU Global 插件，然后在内核代码项目中新建 .vscode/settings.json
vscode 配置 global 插件
```json
{
"gnuGlobal.globalExecutable": "/usr/bin/global",
"gnuGlobal.gtagsExecutable": "/usr/bin/gtags",
"gnuGlobal.objDirPrefix": "/root/linux-5.10.142/.global"
}
```

(3)生成 tag
在 VSCode 工作区中按 F1 执行 Show GNU Global Version，如果配置正确，右下角会显示 global (GNU GLOBAL) <Global_Version>。
执行 Rebuild Gtags Database，等待完成后就可以愉快地阅读 Linux 源码了！

# ubuntu/centos 安装 go
ubuntu 和 cilium/ebpf 编译 bpf 程序，需要安装：
```shell
# ubuntu/centos 都可以，版本是 1.18.9, 1.17.9
wget -c https://dl.google.com/go/go1.17.9.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
export PATH=$PATH:/usr/local/go/bin
source ~/.profile
go version
go env -w GOPROXY="https://mirrors.aliyun.com/goproxy/,direct"
apt install -y clang nginx
# 安装 bpftool
apt install -y linux-tools-5.15.0-56-generic
# 如果是 centos
yum install -y go

# ubuntu 和 cilium/ebpf 编译 bpf 程序，需要安装：
apt install -y clang-13 llvm-13
```

