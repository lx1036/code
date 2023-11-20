
#ifndef XDP_CILIUM_L4LB_IPV4_H
#define XDP_CILIUM_L4LB_IPV4_H

#include <linux/ip.h>

#include "dbg.h"
#include "metrics.h"


// 这个函数意思是计算 ipv4 header 字节大小
// https://en.wikipedia.org/wiki/Internet_Protocol_version_4#IHL
static __always_inline int ipv4_hdrlen(const struct iphdr *ip4) {
    return ip4->ihl * 4; // 4 表示 32bits，4字节. 最小是 5*4=20字节，最大 15*4=60字节, 5<=ip4->ihl<=15
}

#endif //XDP_CILIUM_L4LB_IPV4_H
