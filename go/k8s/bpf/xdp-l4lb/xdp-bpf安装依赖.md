


# 安装依赖环境

安装 xdp 环境的依赖

```
apt install -y clang llvm libelf-dev libpcap-dev build-essential linux-tools-$(uname -r) linux-headers-$(uname -r) linux-tools-common linux-tools-generic tcpdump

```



# load xdp

xdp c code -> (compile) -> elf file(xdp byte code: BTF BPF Type Format) -> load into kernel -> attach to xdp net device(e.g. eth0 interface) -> xdp packet process



