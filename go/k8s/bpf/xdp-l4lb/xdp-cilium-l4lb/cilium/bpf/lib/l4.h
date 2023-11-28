


#ifndef XDP_CILIUM_L4LB_L4_H
#define XDP_CILIUM_L4LB_L4_H


#include <linux/tcp.h>
#include <linux/udp.h>
#include "common.h"
#include "dbg.h"
#include "csum.h"

#define TCP_DPORT_OFF (offsetof(struct tcphdr, dest))
#define TCP_SPORT_OFF (offsetof(struct tcphdr, source))
#define UDP_DPORT_OFF (offsetof(struct udphdr, dest))
#define UDP_SPORT_OFF (offsetof(struct udphdr, source))


// 从 xdp_md 字节数组里取 port 字段
static __always_inline int l4_load_port(struct xdp_md *ctx, int off, __be16 *port) {
    return ctx_load_bytes(ctx, off, port, sizeof(__be16));
}

#endif //XDP_CILIUM_L4LB_L4_H
