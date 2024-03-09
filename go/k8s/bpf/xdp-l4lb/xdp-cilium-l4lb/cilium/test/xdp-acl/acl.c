

#include <stdbool.h>
#include <stdio.h>

// /root/linux-5.10.142/tools/include/uapi/linux/bpf.h
#include <linux/bpf.h>
// /root/linux-5.10.142/tools/include/uapi/linux/types.h
#include <linux/types.h>
// /root/linux-5.10.142/include/uapi/linux/pkt_cls.h
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
// /root/linux-5.10.142/include/uapi/linux/ip.h
#include <linux/ip.h>
// /root/linux-5.10.142/include/uapi/linux/tcp.h
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/in.h>

#include <bpf/bpf_endian.h>
// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>

/**
 * 注：该 map 定义格式需要 cilium/ebpf 库来 load/attach，而不是 ip 命令
 *
 * bpf_ntohs: 在C语言中，ntohs()（网络字节序到主机字节序）函数通常用于将网络传输中的16位整数从网络字节序转换为主机字节序。
  网络字节序是大端字节序（Big-Endian），即最高有效字节位于最低地址，而不同的计算机架构可能使用不同的主机字节序。
  当你从网络接收数据，例如通过套接字（socket）编程读取TCP或UDP报头中的端口号时，需要使用ntohs()来正确解析这些数值。
  因为这些端口号在协议规范中是以网络字节序定义的。
 */

#ifndef likely
#define likely(X) __builtin_expect(!!(X), 1) // !!, true == 1
#endif

#ifndef unlikely
#define unlikely(X) __builtin_expect(!!(X), 0) // !, false == 0
#endif

#ifndef aligned
#define aligned(X) __attribute__((aligned(X)))
#endif

struct {
    __uint(type, BPF_MAP_TYPE_PROG_ARRAY);
    __type(key, __u32);
    __type(value, __u32);
    __uint(max_entries, 1);
} progs SEC(".maps");

static volatile const __u32 XDPACL_DEBUG = 0;

#define bpf_debug_printk(fmt, ...)          \
    do {                                    \
        if (XDPACL_DEBUG)                   \
            bpf_printk(fmt, ##__VA_ARGS__); \
    } while (0)

#define PIN_GLOBAL_NS 2
#define TARGET_IP_NUM 4
#define MAX_ENDPOINT (65535*2)

struct server_ips {
    __u32 target_ips[TARGET_IP_NUM];
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);
    __type(value, struct server_ips);
    __uint(max_entries, 1);
//    __uint(pinning, PIN_GLOBAL_NS);
//    __uint(map_flags, BPF_F_NO_PREALLOC);
} servers SEC(".maps");

struct ports {
    __u16 source; // src port
    __u16 dest; // dst port
};

struct endpoint {
    __u16 dport; // dst port
    __u8 protocol;
};

struct action {
    __u8 action;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);

    // keySize and valueSize need to be sizeof(struct{u32 + u8}) + 1 + padding = 8, 必须都是 8 字节
    // BPF_F_NO_PREALLOC needs to be set
//    __uint(type, BPF_MAP_TYPE_LPM_TRIE);

    __type(key, struct endpoint);
    __type(value, struct action);
    __uint(max_entries, MAX_ENDPOINT);
//    __uint(pinning, PIN_GLOBAL_NS);
//    __uint(map_flags, BPF_F_NO_PREALLOC);
} endpoints SEC(".maps");

static __always_inline int xdp_acl_ipv4_port(struct xdp_md *ctx) {
    void *data_end = (void *) (long) ctx->data_end;
    void *data = (void *) (long) ctx->data;

    struct iphdr *ipv4h = (data + sizeof(struct ethhdr));
    if ((void *) (ipv4h + 1) > data_end) {
        bpf_printk("fail to lookup from servers map. (void *) (ipv4h + 1) > data_end"); // 这里不能打印日志
        return XDP_DROP;
    }
    if (ipv4h->ihl != 5) { // 必然是 5，感觉这种检查意义不大
        bpf_printk("fail to lookup from servers map. ipv4h->ihl != 5");
        return XDP_PASS;
    }

    // 这个逻辑是为了调试，因为 ecs eth0 网卡一直都有流量，这里限定另一台 ecs saddr 发 tcp 包
//    if (ipv4h->saddr != bpf_htonl(0xac100a02)) { /* 172.16.10.2 */
//        return XDP_PASS;
//    }

    // 1.因为 eth0 可能绑定多个 ip 地址，dstIP 必须是指定的 ip，必须做检查过滤
    __u32 key = 0;
    struct server_ips *server = bpf_map_lookup_elem(&servers, &key);
    if (!server) {
        bpf_printk("fail to lookup from servers map.");
        return XDP_PASS;
    }

    bool found = false;
#pragma unroll
    for (int i = 0; i < TARGET_IP_NUM; ++i) {
        if (ipv4h->daddr == server->target_ips[i]) {
            // bpf_printk("target_ip found");
            found = true;
            break; // break 起作用的
        }

//        bpf_printk("target_ip not found %d", i);
    }

    // bpf_printk("target_ip 0x%x", server->target_ips[0]);

    if (!found) {
        // 使用 u32toIP() 报错
//        bpf_printk("dst ip: 0x%x is not target ip, skip it. target_ip 0x%x", ipv4h->daddr, server->target_ips[0]);
        return XDP_PASS;
    }

//    bpf_debug_printk("dst ip: 0x%x is a target ip, acl it.", ipv4h->daddr); // XDPACL_DEBUG 参数起作用的, %pI4 不行

    if (ipv4h->protocol != IPPROTO_TCP && ipv4h->protocol != IPPROTO_UDP) {
        bpf_printk("protocol: %x is not tcp or udp, skip it.", ipv4h->protocol);
        return XDP_PASS;
    }

    // 这样可以不用区分 tcphdr 和 udphdr
    struct ports *port;
    port = data + sizeof(struct ethhdr) + sizeof(struct iphdr);
    if ((void *) (port + 1) > data_end) {
        bpf_printk("fail to fetch tcp/udp ports, skip it.");
        return XDP_PASS;
    }

    // 这个逻辑是为了调试，因为 lo 网卡一直都有 tcp 包，把 ->9090 包过滤出来
    if (bpf_ntohs(port->dest) != 9090) {
        return XDP_PASS;
    }

    // unsigned short, __u16 -> unsigned int, __u32
    struct endpoint endpoint = {}; // 对象初始化
    // network to host shorts, 不加 bpf_ntohs() dport 为 0x8223, 而是加上 hex(9090)=0x2382, 需要注意!!!
    endpoint.dport = bpf_ntohs(port->dest);
    endpoint.protocol = ipv4h->protocol;
    // bpftool map dump name endpoints | jq
    struct action *action;
    action = bpf_map_lookup_elem(&endpoints, &endpoint);
    if (!action) {
        bpf_printk("dport: %x, protocol: %x", endpoint.dport, endpoint.protocol);
        // fail to lookup endpoints map, dport: 9090, protocol: 6, action:0
        bpf_printk("fail to lookup endpoints map, dport: %d, protocol: %x, action:%x", bpf_ntohs(port->dest),
                   ipv4h->protocol, action);
        return XDP_PASS;
    }

    if (action->action == 0) { // deny
        bpf_printk("action of protocol:%x port:%d is deny, drop it.", ipv4h->protocol, bpf_ntohs(port->dest));
        return XDP_DROP;
    }

    bpf_printk("dport %d action is %x", bpf_ntohs(port->dest), action->action);
    return XDP_PASS;
}

/**
 * 根据 protocol/svcPort 来判断 action(XDP_PASS/XDP_DROP)
 *
 * 但是没有验证通过!!! 这里报错一直查找不到对应的 endpoint???
 * 答案已经找到：bpf_ntohs(port->dest) 需要加上 bpf_ntohs(), 验证通过!!!
 *
 * TODO: 根据 bitmap 来寻找 action rule, 减少内存
 */

SEC("xdp_acl")
int xdp_acl_func_imm(struct xdp_md *ctx) {
    return xdp_acl_ipv4_port(ctx);
}

SEC("xdp_acl")
int xdp_acl_func(struct xdp_md *ctx) {
    void *data_end = (void *) (long) ctx->data_end;
    void *data = (void *) (long) ctx->data;

    struct ethhdr *eth;
    eth = data;
    if ((void *) (eth + 1) > data_end)
        return XDP_PASS;

    if (eth->h_proto == bpf_htons(ETH_P_IP)) { // only ipv4
        bpf_tail_call_static(ctx, &progs, 0); // /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
