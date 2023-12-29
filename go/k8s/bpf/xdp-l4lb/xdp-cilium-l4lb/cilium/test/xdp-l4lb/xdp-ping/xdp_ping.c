
// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/xdping_kern.c

/**
 * 调试运行成功
 * 抓包:
 * tcpdump -i veth1 -nneevv -A icmp -w client.pcap
 * ip netns exec xdp_ns0 tcpdump -i veth0 -nneevv -A icmp -w server.pcap (抓不到包)
 *
 * (1)xdp_server 可以使用 ip link 命令挂载
 * (2)xdp_client 只能 go 来挂载
 */

#include <stddef.h>
#include <string.h>
#include <linux/bpf.h>
#include <linux/icmp.h>
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/if_vlan.h>
#include <linux/ip.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define ICMP_ECHO_LEN   64 // ???, 抓包看下 icmp 包长度
#define	XDPING_MAX_COUNT	10
#define	XDPING_DEFAULT_COUNT	4

struct pinginfo {
    __u64	start;
    __be16	seq;
    __u16	count;
    __u32	pad;
    __u64	times[XDPING_MAX_COUNT];
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 256);
    __type(key, __u32);
    __type(value, struct pinginfo);
} ping_map SEC(".maps");

// icmp check 逻辑可以复用!!!
static __always_inline int icmp_check(struct xdp_md *ctx, int type) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    struct ethhdr *eth = data;
    struct icmphdr *icmph;
    struct iphdr *iph;

    if (data + sizeof(*eth) + sizeof(*iph) + ICMP_ECHO_LEN > data_end)
        return XDP_PASS;

    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return XDP_PASS;

    iph = data + sizeof(*eth);
    if (iph->protocol != IPPROTO_ICMP) // 必须是 icmp 协议包
        return XDP_PASS;

    if (bpf_ntohs(iph->tot_len) - sizeof(*iph) != ICMP_ECHO_LEN) // ???
        return XDP_PASS;

    icmph = data + sizeof(*eth) + sizeof(*iph);
    if (icmph->type != type)
        return XDP_PASS;

    return XDP_TX;
}

// 可以复用
static __always_inline void swap_src_dst_mac(void *data) {
    unsigned short *p = data; // unsigned short??
    unsigned short dst[3];

    dst[0] = p[0];
    dst[1] = p[1];
    dst[2] = p[2];
    p[0] = p[3];
    p[1] = p[4];
    p[2] = p[5];
    p[3] = dst[0];
    p[4] = dst[1];
    p[5] = dst[2];
}

/** ipv4 checksum, 可以复用*/
static __always_inline __u16 csum_fold_helper(__wsum sum) {
    sum = (sum & 0xffff) + (sum >> 16);
    return ~((sum & 0xffff) + (sum >> 16));
}
static __always_inline __u16 ipv4_csum(void *data_start, int data_size) {
    __wsum sum;

    sum = bpf_csum_diff(0, 0, data_start, data_size, 0);
    return csum_fold_helper(sum);
}

// tail -n 100 /sys/kernel/debug/tracing/trace

SEC("xdp_drop")
int xdp_drop_prog(struct xdp_md *ctx) {
    return XDP_DROP;
}

SEC("xdp_client")
int xdping_client(struct xdp_md *ctx) {
//    void *data_end = (void *) (long) ctx->data_end;
    void *data = (void *) (long) ctx->data;
//    struct pinginfo *pinginfo = NULL;
    struct ethhdr *eth = data;
    struct icmphdr *icmph;
    struct iphdr *iph;
    __u64 recvtime;
    __u32 raddr;
    __be16 seq;
    int ret;
    __u8 i;

    // icmp echo reply 报文
    ret = icmp_check(ctx, ICMP_ECHOREPLY);
    if (ret != XDP_TX)
        return ret;

    iph = data + sizeof(*eth);
    icmph = data + sizeof(*eth) + sizeof(*iph);
    raddr = iph->saddr;
    char fmt4[] = "iph->saddr:%x, %x";
    bpf_trace_printk(fmt4, sizeof(fmt4), raddr, bpf_htonl(iph->saddr)); // 6401010a, a010164, 10.1.1.100

    /* Record time reply received. */
    char fmt1[] = "pinginfo->seq:%d, icmph->un.echo.sequence:%d\n"; // %x 用于输出十六进制整数值
    recvtime = bpf_ktime_get_ns();
//    seq = icmph->un.echo.sequence;
//    seq = bpf_htons(bpf_ntohs(icmph->un.echo.sequence));
    struct pinginfo *pinginfo = bpf_map_lookup_elem(&ping_map, &raddr); // raddr=10.1.1.100
    if(!pinginfo) {
//        bpf_trace_printk(fmt1, sizeof(fmt1), pinginfo->seq, seq); // 这里调试发现会报错
    }

    char fmt5[] = "icmph->un.echo.sequence: %d";
    bpf_trace_printk(fmt5, sizeof(fmt5), icmph->un.echo.sequence);

    if (!pinginfo) {
//    if (!pinginfo || pinginfo->seq != icmph->un.echo.sequence) {
        return XDP_PASS;
    }

    char fmt2[] = "pinginfo->start: %d";
    bpf_trace_printk(fmt2, sizeof(fmt2), pinginfo->start);
    if (pinginfo->start) {
#pragma clang loop unroll(full)
        for (i = 0; i < XDPING_MAX_COUNT; i++) {
            if (pinginfo->times[i] == 0)
                break;
        }

        /* verifier is fussy here... */
        if (i < XDPING_MAX_COUNT) {
            pinginfo->times[i] = recvtime - pinginfo->start;
            pinginfo->start = 0;
            i++;
        }
        /* No more space for values? */
        if (i == pinginfo->count || i == XDPING_MAX_COUNT)
            return XDP_PASS;
    }

    /* Now convert reply back into echo request. */
    swap_src_dst_mac(data);
    iph->saddr = iph->daddr;
    iph->daddr = raddr;
    icmph->type = ICMP_ECHO;
    seq = bpf_htons(bpf_ntohs(icmph->un.echo.sequence) + 1); // 256(0x0100) -> 512(0x0200) -> 768(0x0300), 这里的 1 是加一个 0x0100(LE)
    icmph->un.echo.sequence = seq;
    icmph->checksum = 0;
    icmph->checksum = ipv4_csum(icmph, ICMP_ECHO_LEN); // 对 icmp 包做 checksum
    pinginfo->seq = seq;
    pinginfo->start = bpf_ktime_get_ns();

    char fmt3[] = "pinginfo->seq:%d";
    bpf_trace_printk(fmt3, sizeof(fmt3), pinginfo->seq);

    return XDP_TX;
}

SEC("xdp_server")
int xdping_server(struct xdp_md *ctx) {
//    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    struct ethhdr *eth = data;
    struct icmphdr *icmph; // icmp 和 tcp 都是四层协议
    struct iphdr *iph;
    __be32 raddr;
    int ret;

    // icmp echo 报文
    ret = icmp_check(ctx, ICMP_ECHO);
    if (ret != XDP_TX)
        return ret;

    iph = data + sizeof(*eth);
    icmph = data + sizeof(*eth) + sizeof(*iph);
    raddr = iph->saddr;

    /* Now convert request into echo reply. */
    swap_src_dst_mac(data);
    iph->saddr = iph->daddr;
    iph->daddr = raddr;
    icmph->type = ICMP_ECHOREPLY;
    icmph->checksum = 0;
    icmph->checksum = ipv4_csum(icmph, ICMP_ECHO_LEN);

    return XDP_TX;
}

char _license[] SEC("license") = "GPL";
