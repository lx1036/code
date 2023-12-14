
/**
 * syn cookie: https://cs.pynote.net/net/tcp/202205052/
 */


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

#include <bpf/bpf_endian.h>
// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 3);
    __type(key, __u32);
    __type(value, __u32);
} results SEC(".maps");

static inline __s64 gen_syncookie(void *data_end, struct bpf_sock *sk, void *iph, __u32 iph_size, struct tcphdr *tcph) {
    __u32 thlen = tcph->doff * 4;
    // 必须是 syn 包，不是 ack 或者 synack 包
    if (tcph->syn && !tcph->ack) {
        // packet should only have an MSS option
        if (thlen != 24)
            return 0;

        if ((void *)(tcph + thlen) > data_end)
            return 0;

        // syn 包添加一个 cookie?
        return bpf_tcp_gen_syncookie(sk, iph, iph_size, tcph, thlen);
    }

    return 0;
}

static inline void check_syncookie(void *ctx, void *data, void *data_end) {
    struct ethhdr *eth;
    struct iphdr *ipv4h;
    struct tcphdr *tcph;
    struct bpf_sock_tuple tup;
    struct bpf_sock *sk;

    int ret;
    __u32 key = 0;
    __u32 key_gen = 1;
    __u32 key_mss = 2;
    __s64 seq_mss;

    // 检查二层 eth
    eth = data;
    if ((void *)(eth + 1) > data_end)
        return;

    switch (bpf_ntohs(eth->h_proto)) {
    case ETH_P_IP:
        ipv4h = data + sizeof(struct ethhdr);
        // 检查三层 ip
        if ((void *)(ipv4h + 1) > data_end)
            return;

        if (ipv4h->ihl != 5) { // ???
            return;
        }

        tcph = data + sizeof(struct ethhdr) + sizeof(struct iphdr);
        if ((void *)(tcph + 1) > data_end)
            return;

        // ip:port
        tup.ipv4.saddr = ipv4h->saddr;
        tup.ipv4.daddr = ipv4h->daddr;
        tup.ipv4.sport = tcph->source;
        tup.ipv4.dport = tcph->dest;

        // look for TCP socket matching tuple, 根据 srcIP:srcPort/dstIP:dstPort 寻找对应的 tcp socket
        sk = bpf_skc_lookup_tcp(ctx, &tup, sizeof(tup.ipv4), BPF_F_CURRENT_NETNS, 0);
        if (!sk)
            return;

        if (sk->state != BPF_TCP_LISTEN)
            goto release;

        seq_mss = gen_syncookie(data_end, sk, ipv4h, sizeof(*ipv4h), tcph);
        // check iphdr 和 tcphdr 是否包含有效的 syn cookie ack
        ret = (int)bpf_tcp_check_syncookie(sk, ipv4h, sizeof(*ipv4h), tcph, sizeof(*tcph));
        break;

    case ETH_P_IPV6:
        return;
    default:
        return;
    }

    if (seq_mss > 0) {
        __u32 cookie = (__u32)seq_mss;
        __u32 mss = seq_mss >> 32; // 高32位: MSS(16 bits)+unused(16 bits)

        bpf_map_update_elem(&results, &key_gen, &cookie, 0);
        bpf_map_update_elem(&results, &key_mss, &mss, 0);
    }

    // 如果当前 SynAck 包含 cookie
    if (ret == 0) {
        // ack_seq - 1 就是 cookie, 测试时在 ack 包里，这里 ack number=2986155030
        // client->server 发 ack 包里时 client 就已经是 tcp establish 状态
        // 33956 -> 8000, 在 ack 包里
        __u32 cookie = bpf_ntohl(tcph->ack_seq) - 1;
        bpf_map_update_elem(&results, &key, &cookie, 0);
    }

release:
    // bpf_skc_lookup_tcp()
    bpf_sk_release(sk);
}


SEC("xdp/check_syncookie")
int check_syncookie_xdp(struct xdp_md *ctx) {
    check_syncookie(ctx, (void *)(long)ctx->data, (void *)(long)ctx->data_end);
    return XDP_PASS;
}

// /root/linux-5.10.142/tools/lib/bpf/libbpf.c
// https://github.com/cilium/ebpf/blob/v0.8.1/elf_reader.go#L1069
// classifier 就是 tc, SchedCLS progType
SEC("classifier/check_syncookie")
int check_syncookie_clsact(struct __sk_buff *skb) {
    check_syncookie(skb, (void *)(long)skb->data, (void *)(long)skb->data_end);
    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";
