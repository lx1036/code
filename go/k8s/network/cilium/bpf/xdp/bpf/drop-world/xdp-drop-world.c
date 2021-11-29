

#include <linux/bpf.h>
#include <compiler.h>


// 由于机器上没有最新的 iproute 和 glibc，只有 cilium pod namespace 里有，还得需要把 xdp.o 拷贝过去然后使用最新的 `ip` 命令
// docker cp ./xdp.o 64d6796758b4:/mnt/xdp.o
// nsenter -t 26416 -m -n ip -force link set dev lxc6e7eb5daff06 xdp obj /mnt/xdp.o sec xdp
// nsenter -t 26416 -m -n ip -force link set dev lxc6e7eb5daff06 xdp off


__section("xdp")
int xdp_drop_the_world(struct xdp_md *ctx) {
    // 意思是无论什么网络数据包，都drop丢弃掉
    return XDP_DROP;
}
