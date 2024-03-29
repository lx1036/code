
// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tcp_hdr_options.c
// /root/linux-5.10.142/tools/testing/selftests/bpf/test_tcp_hdr_options.h
// /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/tcp_hdr_options.c

/**
 * https://github.com/torvalds/linux/commit/ad2f8eb0095e9036724d9cf0eb6960f1e6d52d21
 * https://lwn.net/Articles/827672/
 *
 * This patch adds tests for the new bpf tcp header option feature.

test_tcp_hdr_options.c:
- It tests header option writing and parsing in 3WHS: regular
  connection establishment, fastopen, and syncookie.
- In syncookie, the passive side's bpf prog is asking the active side
  to resend its bpf header option by specifying a RESEND bit in the
  outgoing SYNACK. handle_active_estab() and write_nodata_opt() has
  some details.
- handle_passive_estab() has comments on fastopen.
- It also has test for header writing and parsing in FIN packet.
- Most of the tests is writing an experimental option 254 with magic 0xeB9F.
- The no_exprm_estab() also tests writing a regular TCP option
  without any magic.

test_misc_tcp_options.c:
- It is an one directional test.  Active side writes option and
  passive side parses option.  The focus is to exercise
  the new helpers and API.
- Testing the new helper: bpf_load_hdr_opt() and bpf_store_hdr_opt().
- Testing the bpf_getsockopt(TCP_BPF_SYN).
- Negative tests for the above helpers.
- Testing the sock_ops->skb_data.
 */


#include <stddef.h>
#include <stdbool.h>
// #include <errno.h>

// #include <sys/types.h>
// #include <sys/socket.h>

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/tcp.h>

// /root/linux-5.10.142/tools/lib/bpf/bpf_tracing.h
// /root/linux-5.10.142/tools/lib/bpf/bpf_endian.h
// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
//#include <bpf/errno.h>

#define    EPERM         1    /* Operation not permitted */
#define    ENOENT         2    /* No such file or directory */
#define ENOMSG 80    /* No message of desired type */

// @see https://lore.kernel.org/bpf/20200730205736.3354304-1-kafai@fb.com/

#ifndef SOL_TCP
#define SOL_TCP 6
#endif

#define TCPHDR_FIN 0x01
#define TCPHDR_SYN 0x02
#define TCPHDR_RST 0x04
#define TCPHDR_PSH 0x08
#define TCPHDR_ACK 0x10
#define TCPHDR_URG 0x20
#define TCPHDR_ECE 0x40
#define TCPHDR_CWR 0x80
#define TCPHDR_SYNACK (TCPHDR_SYN | TCPHDR_ACK)
#define TCPOPT_EOL        0
#define TCPOPT_NOP        1
#define TCPOPT_WINDOW        3
#define TCPOPT_EXP        254

#define CG_OK    1
#define CG_ERR    0

#define TCP_BPF_EXPOPT_BASE_LEN 4

#define OPTION_F_RESEND        (1 << OPTION_RESEND)
#define OPTION_F_MAX_DELACK_MS    (1 << OPTION_MAX_DELACK_MS)
#define OPTION_F_RAND        (1 << OPTION_RAND)
#define OPTION_MASK        ((1 << __NR_OPTION_FLAGS) - 1)


struct bpf_test_option {
    __u8 flags;
    __u8 max_delack_ms;
    __u8 rand;
} __attribute__((packed));

static volatile const struct bpf_test_option passive_synack_out = {};
static volatile const struct bpf_test_option passive_fin_out = {};
static volatile const struct bpf_test_option active_syn_out = {};
static volatile const struct bpf_test_option active_fin_out = {};
static volatile const __u8 test_kind = TCPOPT_EXP;
static volatile const __u16 test_magic = 0xeB9F;
static volatile __u32 inherit_cb_flags = 0;

struct bpf_test_option passive_estab_in = {};
struct bpf_test_option passive_fin_in = {};
struct bpf_test_option active_estab_in = {};
struct bpf_test_option active_fin_in = {};

struct linum_err {
    unsigned int linum;
    int err;
};

// key->value
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, int);
    __type(value, struct linum_err);
    __uint(max_entries, 2);
//    __uint(pinning, LIBBPF_PIN_BY_NAME);
} lport_linum_map SEC(".maps");

/* Store in bpf_sk_storage */
struct hdr_stg {
    bool active;
    bool resend_syn; /* active side only */
    bool syncookie;  /* passive side only */
    bool fastopen;    /* passive side only */
};
struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, int);
    __type(value, struct hdr_stg);
//    __uint(pinning, LIBBPF_PIN_BY_NAME);
} hdr_stg_map SEC(".maps");

static inline void clear_hdr_cb_flags(struct bpf_sock_ops *skops) {
    bpf_sock_ops_cb_flags_set(skops, (int) (skops->bpf_sock_ops_cb_flags &
                                            ~(BPF_SOCK_OPS_PARSE_UNKNOWN_HDR_OPT_CB_FLAG |
                                              BPF_SOCK_OPS_WRITE_HDR_OPT_CB_FLAG)));
}

static inline void clear_parse_all_hdr_cb_flags(struct bpf_sock_ops *skops) {
    bpf_sock_ops_cb_flags_set(skops, (int) (skops->bpf_sock_ops_cb_flags & ~BPF_SOCK_OPS_PARSE_ALL_HDR_OPT_CB_FLAG));
}

static inline void set_parse_all_hdr_cb_flags(struct bpf_sock_ops *skops) {
    bpf_sock_ops_cb_flags_set(skops, (int) (skops->bpf_sock_ops_cb_flags | BPF_SOCK_OPS_PARSE_ALL_HDR_OPT_CB_FLAG));
}


#define RET_CG_ERR(__err) ({            \
    struct linum_err __linum_err;        \
    int __lport;                \
                        \
    __linum_err.linum = __LINE__;        \
    __linum_err.err = __err;        \
    __lport = skops->local_port;        \
    bpf_map_update_elem(&lport_linum_map, &__lport, &__linum_err, BPF_NOEXIST); \
    clear_hdr_cb_flags(skops);                    \
    clear_parse_all_hdr_cb_flags(skops);                \
    return CG_ERR;                            \
})


struct tcp_exprm_opt {
    __u8 kind;
    __u8 len;
    __u16 magic;
    union {
        __u8 data[4];
        __u32 data32;
    };
} __attribute__((packed));

struct tcp_opt {
    __u8 kind;
    __u8 len;
    union {
        __u8 data[4];
        __u32 data32;
    };
} __attribute__((packed));


#define TEST_OPTION_FLAGS(flags, option) (1 & ((flags) >> (option)))

enum {
    OPTION_RESEND,
    OPTION_MAX_DELACK_MS,
    OPTION_RAND,
    __NR_OPTION_FLAGS,
};

static int ret_cg_error(struct bpf_sock_ops *skops, int err) {
    struct linum_err linum_err;
    int lport;

    linum_err.linum = __LINE__;
    linum_err.err = err;
    lport = (int) skops->local_port;
    bpf_map_update_elem(&lport_linum_map, &lport, &linum_err, BPF_NOEXIST);
    clear_hdr_cb_flags(skops);
    clear_parse_all_hdr_cb_flags(skops);
    return CG_ERR;
}

static __u8 option_total_len(__u8 flags) {
    __u8 i, len = 1; /* +1 for flags */

    if (!flags)
        return 0;

    /* RESEND bit does not use a byte */
    for (i = OPTION_RESEND + 1; i < __NR_OPTION_FLAGS; i++)
        len += !!TEST_OPTION_FLAGS(flags, i);

    if (test_kind == TCPOPT_EXP)
        return len + TCP_BPF_EXPOPT_BASE_LEN;
    else
        return len + 2; /* +1 kind, +1 kind-len */
}

static bool skops_current_mss(const struct bpf_sock_ops *skops) {
    return skops->args[0] == BPF_WRITE_HDR_TCP_CURRENT_MSS;
}

static int current_mss_opt_len(struct bpf_sock_ops *skops) {
    /* Reserve maximum that may be needed */
    int err;

    err = bpf_reserve_hdr_opt(skops, option_total_len(OPTION_MASK), 0);
    if (err)
        ret_cg_error(skops, err);

    return CG_OK;
}

static void write_test_option(const struct bpf_test_option *test_opt, __u8 *data) {
    __u8 offset = 0;

    data[offset++] = test_opt->flags;
    if (TEST_OPTION_FLAGS(test_opt->flags, OPTION_MAX_DELACK_MS))
        data[offset++] = test_opt->max_delack_ms;

    if (TEST_OPTION_FLAGS(test_opt->flags, OPTION_RAND))
        data[offset++] = test_opt->rand;
}

static int parse_test_option(struct bpf_test_option *opt, const __u8 *start) {
    opt->flags = *start++;

    if (TEST_OPTION_FLAGS(opt->flags, OPTION_MAX_DELACK_MS))
        opt->max_delack_ms = *start++;

    if (TEST_OPTION_FLAGS(opt->flags, OPTION_RAND))
        opt->rand = *start++;

    return 0;
}

static int store_option(struct bpf_sock_ops *skops, const struct bpf_test_option *test_opt) {
    union {
        struct tcp_exprm_opt exprm;
        struct tcp_opt regular;
    } write_opt;
    int err;

    if (test_kind == TCPOPT_EXP) {
        write_opt.exprm.kind = TCPOPT_EXP;
        write_opt.exprm.len = option_total_len(test_opt->flags);
        write_opt.exprm.magic = __bpf_htons(test_magic);
        write_opt.exprm.data32 = 0;
        write_test_option(test_opt, write_opt.exprm.data);
        err = (int) bpf_store_hdr_opt(skops, &write_opt.exprm, sizeof(write_opt.exprm), 0);
    } else {
        write_opt.regular.kind = test_kind;
        write_opt.regular.len = option_total_len(test_opt->flags);
        write_opt.regular.data32 = 0;
        write_test_option(test_opt, write_opt.regular.data);
        err = (int) bpf_store_hdr_opt(skops, &write_opt.regular, sizeof(write_opt.regular), 0);
    }

    if (err)
        ret_cg_error(skops, err);

    return CG_OK;
}

static int load_option(struct bpf_sock_ops *skops, struct bpf_test_option *test_opt, bool from_syn) {
    union {
        struct tcp_exprm_opt exprm;
        struct tcp_opt regular;
    } search_opt;
    int ret, load_flags = from_syn ? BPF_LOAD_HDR_OPT_TCP_SYN : 0;

    if (test_kind == TCPOPT_EXP) {
        search_opt.exprm.kind = TCPOPT_EXP;
        search_opt.exprm.len = 4;
        search_opt.exprm.magic = __bpf_htons(test_magic);
        search_opt.exprm.data32 = 0;
        ret = (int) bpf_load_hdr_opt(skops, &search_opt.exprm, sizeof(search_opt.exprm), load_flags);
        if (ret < 0)
            return ret;
        return parse_test_option(test_opt, search_opt.exprm.data);
    } else {
        search_opt.regular.kind = test_kind;
        search_opt.regular.len = 0;
        search_opt.regular.data32 = 0;
        ret = (int) bpf_load_hdr_opt(skops, &search_opt.regular, sizeof(search_opt.regular), load_flags);
        if (ret < 0)
            return ret;
        return parse_test_option(test_opt, search_opt.regular.data);
    }
}

static inline __u8 skops_tcp_flags(const struct bpf_sock_ops *skops) {
    return skops->skb_tcp_flags;
}

static int handle_parse_hdr(struct bpf_sock_ops *skops) {
    struct tcphdr *tcp_header;
    struct hdr_stg *hdr_stg;

    if (!skops->sk) {
        return ret_cg_error(skops, 0);
    }

    // 判断报文长度正确性
    tcp_header = skops->skb_data;
    if ((void *) (tcp_header + 1) > skops->skb_data_end)
        ret_cg_error(skops, 0);

    hdr_stg = bpf_sk_storage_get(&hdr_stg_map, skops->sk, NULL, 0);
    if (!hdr_stg)
        ret_cg_error(skops, 0);

    if (hdr_stg->resend_syn || hdr_stg->fastopen)
        /* The PARSE_ALL_HDR cb flag was turned on
         * to ensure that the previously written
         * options have reached the peer.
         * Those previously written option includes:
         *     - Active side: resend_syn in ACK during syncookie
         *      or
         *     - Passive side: SYNACK during fastopen
         *
         * A valid packet has been received here after
         * the 3WHS, so the PARSE_ALL_HDR cb flag
         * can be cleared now.
         */
        clear_parse_all_hdr_cb_flags(skops);

    if (hdr_stg->resend_syn && !active_fin_out.flags)
        /* Active side resent the syn option in ACK
         * because the server was in syncookie mode.
         * A valid packet has been received, so
         * clear header cb flags if there is no
         * more option to send.
         */
        clear_hdr_cb_flags(skops);

    if (hdr_stg->fastopen && !passive_fin_out.flags)
        /* Passive side was in fastopen.
         * A valid packet has been received, so
         * the SYNACK has reached the peer.
         * Clear header cb flags if there is no more
         * option to send.
         */
        clear_hdr_cb_flags(skops);


    if (tcp_header->fin) {
        struct bpf_test_option *fin_opt;
        int err;

        if (hdr_stg->active)
            fin_opt = &active_fin_in;
        else
            fin_opt = &passive_fin_in;

        err = load_option(skops, fin_opt, false);
        if (err && err != -ENOMSG)
            return ret_cg_error(skops, err);
    }

    return CG_OK;
}

static int write_synack_opt(struct bpf_sock_ops *skops) {
    struct bpf_test_option opt;

    if (!passive_synack_out.flags)
        /* We should not even be called since no header
         * space has been reserved.
         */
        return ret_cg_error(skops, 0);

    opt = passive_synack_out;
    if (skops->args[0] == BPF_WRITE_HDR_TCP_SYNACK_COOKIE) { // Kernel is in syncookie mode when sending a SYN
        opt.flags |= (1 << OPTION_RESEND);
    }

    return store_option(skops, &opt);
}

static int synack_opt_len(struct bpf_sock_ops *skops) {
    struct bpf_test_option test_opt = {};
    __u8 optlen;
    int err;

    if (!passive_synack_out.flags)
        return CG_OK;

    err = load_option(skops, &test_opt, true);
    /* bpf_test_option is not found */
    if (err == -ENOMSG) {
        return CG_OK;
    }
    if (err) {
        return ret_cg_error(skops, err);
    }

    optlen = option_total_len(passive_synack_out.flags);
    if (optlen) {
        err = (int) bpf_reserve_hdr_opt(skops, optlen, 0);
        if (err) {
            return ret_cg_error(skops, err);
        }
    }

    return CG_OK;
}

static int write_syn_opt(struct bpf_sock_ops *skops) {
    if (!active_syn_out.flags)
        return ret_cg_error(skops, 0);

    struct bpf_test_option opt;
    opt = active_syn_out;
    return store_option(skops, &opt);
}

static int syn_opt_len(struct bpf_sock_ops *skops) {
    __u8 optlen;
    int err;

    if (!active_syn_out.flags)
        return CG_OK;

    optlen = option_total_len(active_syn_out.flags);
    if (optlen) {
        err = (int) bpf_reserve_hdr_opt(skops, optlen, 0);
        if (err)
            return ret_cg_error(skops, err);
    }

    return CG_OK;
}

static int resend_in_ack(struct bpf_sock_ops *skops) {
    struct hdr_stg *hdr_stg;

    if (!skops->sk)
        return -1;

    hdr_stg = bpf_sk_storage_get(&hdr_stg_map, skops->sk, NULL, 0);
    if (!hdr_stg)
        return -1;

    return !!hdr_stg->resend_syn;
}

static int write_fin_opt(struct bpf_sock_ops *skops) {
    int resend;

    resend = resend_in_ack(skops);
    if (resend < 0)
        ret_cg_error(skops, 0);

    if (resend)
        return write_syn_opt(skops);

    return CG_OK;
}

static int fin_opt_len(struct bpf_sock_ops *skops) {
    volatile const struct bpf_test_option *opt;
    struct hdr_stg *hdr_stg;
    __u8 optlen;
    int err;

    if (!skops->sk)
        ret_cg_error(skops, 0);

    hdr_stg = bpf_sk_storage_get(&hdr_stg_map, skops->sk, NULL, 0);
    if (!hdr_stg)
        ret_cg_error(skops, 0);

    if (hdr_stg->active)
        opt = &active_fin_out;
    else
        opt = &passive_fin_out;

    optlen = option_total_len(opt->flags);
    if (optlen) {
        err = (int)bpf_reserve_hdr_opt(skops, optlen, 0);
        if (err)
            ret_cg_error(skops, err);
    }

    return CG_OK;
}

static int write_nodata_opt(struct bpf_sock_ops *skops) {
    int resend;

    resend = resend_in_ack(skops);
    if (resend < 0)
        ret_cg_error(skops, 0);

    if (resend)
        return write_syn_opt(skops);

    return CG_OK;
}

static int nodata_opt_len(struct bpf_sock_ops *skops) {
    int resend;

    resend = resend_in_ack(skops);
    if (resend < 0)
        ret_cg_error(skops, 0);

    if (resend)
        return syn_opt_len(skops);

    return CG_OK;
}

static int write_data_opt(struct bpf_sock_ops *skops) {
    return write_nodata_opt(skops);
}

static int data_opt_len(struct bpf_sock_ops *skops) {
    /* Same as the nodata version.  Mostly to show
     * an example usage on skops->skb_len.
     */
    return nodata_opt_len(skops);
}

static int handle_write_hdr_opt(struct bpf_sock_ops *skops) {
    __u8 tcp_flags = skops->skb_tcp_flags; // SYN, SYNACK, FIN
    struct tcphdr *th;

    if ((tcp_flags & TCPHDR_SYNACK) == TCPHDR_SYNACK)
        return write_synack_opt(skops); // synack

    if (tcp_flags & TCPHDR_SYN)
        return write_syn_opt(skops); // syn

    if (tcp_flags & TCPHDR_FIN)
        return write_fin_opt(skops); // fin

    th = skops->skb_data;
    if ((void *) (th + 1) > skops->skb_data_end)
        ret_cg_error(skops, 0);

    // 0110 .... = Header Length: 24 bytes (6), 表示 tcp header 总共长度 24 字节, 最基础是20字节，>20说明有 tcp options
    if (skops->skb_len > (th->doff << 2))
        return write_data_opt(skops);

    return write_nodata_opt(skops);
}

static int handle_hdr_opt_len(struct bpf_sock_ops *skops) {
    __u8 tcp_flags = skops_tcp_flags(skops);

    if ((tcp_flags & TCPHDR_SYNACK) == TCPHDR_SYNACK)
        return synack_opt_len(skops);

    if (tcp_flags & TCPHDR_SYN)
        return syn_opt_len(skops);

    if (tcp_flags & TCPHDR_FIN)
        return fin_opt_len(skops);

    if (skops_current_mss(skops))
        /* The kernel is calculating the MSS */
        return current_mss_opt_len(skops);

    if (skops->skb_len)
        return data_opt_len(skops);

    return nodata_opt_len(skops);
}

static int set_delack_max(struct bpf_sock_ops *skops, __u8 max_delack_ms) {
    __u32 max_delack_us = max_delack_ms * 1000;

    return bpf_setsockopt(skops, SOL_TCP, TCP_BPF_DELACK_MAX, &max_delack_us, sizeof(max_delack_us));
}

static int set_rto_min(struct bpf_sock_ops *skops, __u8 peer_max_delack_ms) {
    __u32 min_rto_us = peer_max_delack_ms * 1000;

    return bpf_setsockopt(skops, SOL_TCP, TCP_BPF_RTO_MIN, &min_rto_us, sizeof(min_rto_us));
}

static int handle_passive_estab(struct bpf_sock_ops *skops) {
    struct hdr_stg init_stg = {};
    int err;
    struct tcphdr *th;

    inherit_cb_flags = skops->bpf_sock_ops_cb_flags;
    err = load_option(skops, &passive_estab_in, true);
    if (err == -ENOENT) {
        /* saved_syn is not found. It was in syncookie mode.
         * We have asked the active side to resend the options
         * in ACK, so try to find the bpf_test_option from ACK now.
         */
        err = load_option(skops, &passive_estab_in, false);
        init_stg.syncookie = true;
    }

    /* ENOMSG: The bpf_test_option is not found which is fine.
     * Bail out now for all other errors.
     */
    if (err && err != -ENOMSG)
        ret_cg_error(skops, err);

    th = skops->skb_data;
    if ((void *) (th + 1) > skops->skb_data_end)
        ret_cg_error(skops, 0);

    if (th->syn) {
        /* Fastopen */

        /* Cannot clear cb_flags to stop write_hdr cb.
         * synack is not sent yet for fast open.
         * Even it was, the synack may need to be retransmitted.
         *
         * PARSE_ALL_HDR cb flag is set to learn
         * if synack has reached the peer.
         * All cb_flags will be cleared in handle_parse_hdr().
         */
        set_parse_all_hdr_cb_flags(skops);
        init_stg.fastopen = true;
    } else if (!passive_fin_out.flags) {
        /* No options will be written from now */
        clear_hdr_cb_flags(skops);
    }

    if (!skops->sk ||
        !bpf_sk_storage_get(&hdr_stg_map, skops->sk, &init_stg,
                            BPF_SK_STORAGE_GET_F_CREATE))
        ret_cg_error(skops, 0);

    if (passive_synack_out.max_delack_ms) {
        err = set_delack_max(skops, passive_synack_out.max_delack_ms);
        if (err)
            ret_cg_error(skops, err);
    }

    if (passive_estab_in.max_delack_ms) {
        err = set_rto_min(skops, passive_estab_in.max_delack_ms);
        if (err)
            ret_cg_error(skops, err);
    }

    return CG_OK;
}

static int handle_active_estab(struct bpf_sock_ops *skops) {
    struct hdr_stg init_stg = {
            .active = true,
    };
    int err;

    err = load_option(skops, &active_estab_in, false);
    if (err && err != -ENOMSG)
        ret_cg_error(skops, err);

    init_stg.resend_syn = TEST_OPTION_FLAGS(active_estab_in.flags, OPTION_RESEND);
    if (!skops->sk || !bpf_sk_storage_get(&hdr_stg_map, skops->sk, &init_stg, BPF_SK_STORAGE_GET_F_CREATE)) {
        ret_cg_error(skops, 0);
    }

    if (init_stg.resend_syn)
        /* Don't clear the write_hdr cb now because
         * the ACK may get lost and retransmit may
         * be needed.
         *
         * PARSE_ALL_HDR cb flag is set to learn if this
         * resend_syn option has received by the peer.
         *
         * The header option will be resent until a valid
         * packet is received at handle_parse_hdr()
         * and all hdr cb flags will be cleared in
         * handle_parse_hdr().
         */
        set_parse_all_hdr_cb_flags(skops);
    else if (!active_fin_out.flags)
        /* No options will be written from now */
        clear_hdr_cb_flags(skops);

    if (active_syn_out.max_delack_ms) {
        err = set_delack_max(skops, active_syn_out.max_delack_ms);
        if (err)
            ret_cg_error(skops, err);
    }

    if (active_estab_in.max_delack_ms) {
        err = set_rto_min(skops, active_estab_in.max_delack_ms);
        if (err)
            ret_cg_error(skops, err);
    }

    return CG_OK;
}

static inline void set_hdr_cb_flags(struct bpf_sock_ops *skops, __u32 extra) {
    bpf_sock_ops_cb_flags_set(skops, (int) (skops->bpf_sock_ops_cb_flags | BPF_SOCK_OPS_PARSE_UNKNOWN_HDR_OPT_CB_FLAG |
                                            BPF_SOCK_OPS_WRITE_HDR_OPT_CB_FLAG | extra));
}

/**
 * 目前bpf代码还有错误，报错 "loading objects: field Estab: program estab: load program: permission denied: invalid access to packet, off=12 size=2, R8(id=0,off=12,r=0): R8 offset is outside of the packet (258 line(s) omitted)"
 */

// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tcp_hdr_options.c
SEC("sockops/estab")
int estab(struct bpf_sock_ops *skops) {
    int true_val = 1;

    switch (skops->op) {
        // server 调用 listen() 切换到 listen 状态时调用
        case BPF_SOCK_OPS_TCP_LISTEN_CB:
            /**
            TCP_SAVE_SYN is a socket option that if saves SYN packet
            具体来说，当 TCP_SAVE_SYN 标志打开时，内核会在结构体 tcp_options_received 中存储发送端的 SYN 报文信息。
            这个结构体中包括了源 IP 地址、源端口号、初始序列号等字段。在连接建立完成后，
            应用程序可以通过套接字选项来获取这些信息。
            */
            // tcpSaveSyn, err := unix.GetsockoptInt(serverFd, unix.SOL_TCP, unix.TCP_SAVE_SYN)
            bpf_setsockopt(skops, SOL_TCP, TCP_SAVE_SYN, &true_val, sizeof(true_val));
            set_hdr_cb_flags(skops, BPF_SOCK_OPS_STATE_CB_FLAG);
            bpf_printk("BPF_SOCK_OPS_TCP_LISTEN_CB");
            break;

            // client 调用 connect() 状态时调用
        case BPF_SOCK_OPS_TCP_CONNECT_CB:
            set_hdr_cb_flags(skops, 0);
            bpf_printk("BPF_SOCK_OPS_TCP_CONNECT_CB");
            break;

            // active connection，就是 client 主动建立 tcp connection
        case BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB:
            bpf_printk("BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB");
            return handle_active_estab(skops);

            // passive connection, 就是 server 被动建立 tcp connection
        case BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB:
            bpf_printk("BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB");
            return handle_passive_estab(skops);

            // client/server 写包 packet option
            // Call bpf when the kernel is writing header options for the outgoing packet.
        case BPF_SOCK_OPS_WRITE_HDR_OPT_CB:
            bpf_printk("BPF_SOCK_OPS_WRITE_HDR_OPT_CB");
            return handle_write_hdr_opt(skops);

        // server/client 收包 parse packet option
        // ->server 收到的包 parse hdr, handle the packets received at an already established connection
        case BPF_SOCK_OPS_PARSE_HDR_OPT_CB:
            bpf_printk("BPF_SOCK_OPS_PARSE_HDR_OPT_CB");
            return handle_parse_hdr(skops);

        case BPF_SOCK_OPS_HDR_OPT_LEN_CB:
            bpf_printk("BPF_SOCK_OPS_HDR_OPT_LEN_CB");
            return handle_hdr_opt_len(skops);
    }

    return CG_OK;
}


// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_misc_tcp_hdr_options.c
SEC("sockops/misc_estab")
int misc_estab(struct bpf_sock_ops *skops) {

    return CG_OK;
}


char _license[] SEC("license") = "GPL";
