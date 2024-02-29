

/**
 * https://lwn.net/Articles/542629/
 *
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_select_reuseport_kern.c
 */

#include <errno.h>
#include <stdbool.h>
#include <stddef.h>
#include <sys/socket.h>

#include <linux/bpf.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/if_ether.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

enum result {
    DROP_ERR_INNER_MAP,
    DROP_ERR_SKB_DATA,
    DROP_ERR_SK_SELECT_REUSEPORT,
    DROP_MISC,
    PASS,
    PASS_ERR_SK_SELECT_REUSEPORT,
    NR_RESULTS,
};

struct data_check {
    __u32 ip_protocol;
    __u32 skb_addrs[8]; // 为何是 8???
    __u16 skb_ports[2];
    __u16 eth_protocol;
    __u8 bind_inany;
    __u8 equal_check_end[0];

    __u32 len;
    __u32 hash;
};

struct cmd {
    __u32 reuseport_index;
    __u32 pass_on_failure;
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY_OF_MAPS); // 测试 array_of_maps
//    __uint(type, BPF_MAP_TYPE_HASH_OF_MAPS); // hash_of_maps map type 也必须有 inner_map
    __uint(max_entries, 1);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(__u32));
} outer_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, int);
} index_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, NR_RESULTS);
    __type(key, __u32);
    __type(value, __u32);
} result_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct data_check);
} data_check_map SEC(".maps");

/**
 * 一直报 bpf 程序问题，没有找到原因???
 */


SEC("sk_reuseport")
int select_by_skb_data(struct sk_reuseport_md *ctx) {
    void *data = (void *) (long) ctx->data;
    void *data_end = (void *) (long) ctx->data_end;
    struct data_check data_check = {};
    enum result result;
    struct cmd *cmd, cmd_copy; // cmd 是四层 payload
    __u32 *result_cnt;

    data_check.eth_protocol = ctx->eth_protocol;
    data_check.ip_protocol = ctx->ip_protocol;
    data_check.len = ctx->len;
    data_check.hash = ctx->hash;
    data_check.bind_inany = ctx->bind_inany;
    if (data_check.eth_protocol == bpf_htons(ETH_P_IP)) { // ipv4
        // 直接从 ctx 里 load bytes, 读取 saddr ???
        if (bpf_skb_load_bytes_relative(ctx, offsetof(struct iphdr, saddr), data_check.skb_addrs, 8, BPF_HDR_START_NET)) {
            result = DROP_MISC;
            goto done;
        }
    } else { // ipv6
        return SK_PASS;
    }

    if (data_check.ip_protocol == IPPROTO_TCP) {
        struct tcphdr *tcph = (struct tcphdr *) data;
//        if ((void *) (tcph + 1) > data_end) {
//            result = DROP_MISC;
//            goto done;
//        }

        data_check.skb_ports[0] = tcph->source;
        data_check.skb_ports[1] = tcph->dest;
        if (tcph->fin) {
            /* The connection is being torn down at the end of a
            * test. It can't contain a cmd, so return early.
            */
            return SK_PASS;
        }

        if ((tcph->doff << 2) + sizeof(*cmd) > data_check.len) { // ???
            result = DROP_ERR_SKB_DATA;
            goto done;
        }
        if (bpf_skb_load_bytes(ctx, tcph->doff << 2, &cmd_copy, sizeof(cmd_copy))) { // ???
            result = DROP_MISC;
            goto done;
        }
        cmd = &cmd_copy;
    } else if (data_check.ip_protocol == IPPROTO_UDP) {
//        struct udphdr *udph = (struct udphdr *) data;
//        if ((void *) (udph + 1) > data_end) {
//            result = DROP_MISC;
//            goto done;
//        }
//
//        data_check.skb_ports[0] = udph->source;
//        data_check.skb_ports[1] = udph->dest;
//
//        if (sizeof(struct udphdr) + sizeof(*cmd) > data_check.len) {
//            result = DROP_ERR_SKB_DATA;
//            goto done;
//        }
//
//        if (data + sizeof(struct udphdr) + sizeof(*cmd) > data_end) {
//            if (bpf_skb_load_bytes(ctx, sizeof(struct udphdr), &cmd_copy, sizeof(cmd_copy))) {
//                result = DROP_MISC;
//                goto done;
//            }
//            cmd = &cmd_copy;
//        } else {
//            cmd = data + sizeof(struct udphdr); // data 偏移 udp header 后，就是 cmd payload
//        }
    } else {
        result = DROP_MISC;
        goto done;
    }

    __u32 index_zero = 0;
    __u32 *reuseport_array_map = bpf_map_lookup_elem(&outer_map, &index_zero);
    if (!reuseport_array_map) {
        result = DROP_ERR_INNER_MAP;
        goto done;
    }

    __u32 index = cmd->reuseport_index;
    int *index_ovr = bpf_map_lookup_elem(&index_map, &index_zero);
    if (!index_ovr) {
        result = DROP_MISC;
        goto done;
    }
    if (*index_ovr != -1) {
        index = *index_ovr;
        *index_ovr = -1;
    }

    int err = bpf_sk_select_reuseport(ctx, reuseport_array_map, &index, 0);
    if (!err) { // err 是 0/负数，这里可以 !err
        result = PASS;
        goto done;
    }

    if (cmd->pass_on_failure) {
        result = PASS_ERR_SK_SELECT_REUSEPORT;
        goto done;
    } else {
        result = DROP_ERR_SK_SELECT_REUSEPORT;
        goto done;
    }

    done:
    result_cnt = bpf_map_lookup_elem(&result_map, &result);
    if (!result_cnt) {
        return SK_DROP;
    }

    bpf_map_update_elem(&data_check_map, &index_zero, &data_check, BPF_ANY);
    (*result_cnt)++; // __u32

    return result < PASS ? SK_DROP : SK_PASS;
}


char _license[] SEC("license") = "GPL";
