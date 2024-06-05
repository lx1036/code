
#ifndef __LIB_MAPS_H_
#define __LIB_MAPS_H_

#include <stddef.h>
#include <linux/bpf.h>
#include <linux/types.h>
#include <bpf/bpf_helpers.h>
#include <linux/pkt_cls.h>

#include "common.h"

#define ENDPOINT_KEY_IPV4 1
#define ENDPOINT_KEY_IPV6 2

#define ENDPOINT_F_HOST		1 /* Special endpoint representing local host */

/* Subset of kernel's include/linux/kconfig.h */
#define __ARG_PLACEHOLDER_1 0,
#define __take_second_arg(__ignored, val, ...) val
#define ____is_defined(arg1_or_junk) __take_second_arg(arg1_or_junk 1, 0)
#define ___is_defined(val)           ____is_defined(__ARG_PLACEHOLDER_##val)
#define __is_defined(x)              ___is_defined(x)
#define is_defined(option)           __is_defined(option)

/* Magic ctx->mark identifies packets origination and encryption status.
 *
 * The upper 16 bits plus lower 8 bits (e.g. mask 0XFFFF00FF) contain the
 * packets security identity. The lower/upper halves are swapped to recover
 * the identity.
 *
 * In case of MARK_MAGIC_PROXY_EGRESS_EPID the upper 16 bits carry the Endpoint
 * ID instead of the security identity and the lower 8 bits will be zeroes.
 *
 * The 4 bits at 0X0F00 provide
 *  - the magic marker values which indicate whether the packet is coming from
 *    an ingress or egress proxy, a local process and its current encryption
 *    status.
 *
 * The 4 bits at 0xF000 provide
 *  - the key index to use for encryption when multiple keys are in-flight.
 *    In the IPsec case this becomes the SPI on the wire.
 */
#define MARK_MAGIC_HOST_MASK		0x0F00
#define MARK_MAGIC_PROXY_EGRESS_EPID	0x0900 /* mark carries source endpoint ID */
#define MARK_MAGIC_PROXY_INGRESS	0x0A00
#define MARK_MAGIC_PROXY_EGRESS		0x0B00
#define MARK_MAGIC_HOST			0x0C00
#define MARK_MAGIC_DECRYPT		0x0D00
#define MARK_MAGIC_ENCRYPT		0x0E00
#define MARK_MAGIC_IDENTITY		0x0F00 /* mark carries identity */
#define MARK_MAGIC_TO_PROXY		0x0200
#define MARK_MAGIC_KEY_ID		0xF000
#define MARK_MAGIC_KEY_MASK		0xFF00
/* IPSec cannot be configured with NodePort BPF today, hence non-conflicting
 * overlap with MARK_MAGIC_KEY_ID.
 */
#define MARK_MAGIC_SNAT_DONE		0x1500
/* MARK_MAGIC_HEALTH_IPIP_DONE can overlap with MARK_MAGIC_SNAT_DONE with both
 * being mutual exclusive given former is only under DSR. Used to push health
 * probe packets to ipip tunnel device & to avoid looping back.
 */
#define MARK_MAGIC_HEALTH_IPIP_DONE	MARK_MAGIC_SNAT_DONE

/* MARK_MAGIC_HEALTH can overlap with MARK_MAGIC_DECRYPT with both being
 * mutual exclusive. Note, MARK_MAGIC_HEALTH is user-facing UAPI for LB!
 */
#define MARK_MAGIC_HEALTH		MARK_MAGIC_DECRYPT

/*
 * ctx->tc_index uses
 *
 * cilium_host @egress
 *   bpf_host -> bpf_lxc
 */
#define TC_INDEX_F_SKIP_INGRESS_PROXY	1
#define TC_INDEX_F_SKIP_EGRESS_PROXY	2
#define TC_INDEX_F_SKIP_NODEPORT	4
#define TC_INDEX_F_SKIP_RECIRCULATION	8
#define TC_INDEX_F_SKIP_HOST_FIREWALL	16

/* ctx_{load,store}_meta() usage: */
enum {
    CB_SRC_LABEL,
#define	CB_PORT			CB_SRC_LABEL	/* Alias, non-overlapping */
#define	CB_HINT			CB_SRC_LABEL	/* Alias, non-overlapping */
#define	CB_PROXY_MAGIC		CB_SRC_LABEL	/* Alias, non-overlapping */
#define	CB_ENCRYPT_MAGIC	CB_SRC_LABEL	/* Alias, non-overlapping */
#define	CB_DST_ENDPOINT_ID	CB_SRC_LABEL    /* Alias, non-overlapping */
    CB_IFINDEX,
#define	CB_ADDR_V4		CB_IFINDEX	/* Alias, non-overlapping */
#define	CB_ADDR_V6_1		CB_IFINDEX	/* Alias, non-overlapping */
#define	CB_ENCRYPT_IDENTITY	CB_IFINDEX	/* Alias, non-overlapping */
#define	CB_IPCACHE_SRC_LABEL	CB_IFINDEX	/* Alias, non-overlapping */
    CB_POLICY,
#define	CB_ADDR_V6_2		CB_POLICY	/* Alias, non-overlapping */
    CB_NAT46_STATE,
#define CB_NAT			CB_NAT46_STATE	/* Alias, non-overlapping */
#define	CB_ADDR_V6_3		CB_NAT46_STATE	/* Alias, non-overlapping */
#define	CB_FROM_HOST		CB_NAT46_STATE	/* Alias, non-overlapping */
    CB_CT_STATE,
#define	CB_ADDR_V6_4		CB_CT_STATE	/* Alias, non-overlapping */
#define	CB_ENCRYPT_DST		CB_CT_STATE	/* Alias, non-overlapping,
						 * Not used by xfrm.
						 */
#define	CB_CUSTOM_CALLS		CB_CT_STATE	/* Alias, non-overlapping */
};

#define CILIUM_MAP_POLICY	1
#define CILIUM_MAP_CALLS	2
#define CILIUM_MAP_CUSTOM_CALLS	3
#define CILIUM_MAP_EGRESSPOLICY	4

#define PIN_NONE		0
#define PIN_OBJECT_NS		1
#define PIN_GLOBAL_NS		2

/* These are shared with test/bpf/check-complexity.sh, when modifying any of
 * the below, that script should also be updated.
 */
#define CILIUM_CALL_DROP_NOTIFY			1
#define CILIUM_CALL_ERROR_NOTIFY		2
#define CILIUM_CALL_SEND_ICMP6_ECHO_REPLY	3
#define CILIUM_CALL_HANDLE_ICMP6_NS		4
#define CILIUM_CALL_SEND_ICMP6_TIME_EXCEEDED	5
#define CILIUM_CALL_ARP				6
#define CILIUM_CALL_IPV4_FROM_LXC		7
#define CILIUM_CALL_NAT64			8
#define CILIUM_CALL_NAT46			9
#define CILIUM_CALL_IPV6_FROM_LXC		10
#define CILIUM_CALL_IPV4_TO_LXC_POLICY_ONLY	11
#define CILIUM_CALL_IPV4_TO_HOST_POLICY_ONLY	CILIUM_CALL_IPV4_TO_LXC_POLICY_ONLY
#define CILIUM_CALL_IPV6_TO_LXC_POLICY_ONLY	12
#define CILIUM_CALL_IPV6_TO_HOST_POLICY_ONLY	CILIUM_CALL_IPV6_TO_LXC_POLICY_ONLY
#define CILIUM_CALL_IPV4_TO_ENDPOINT		13
#define CILIUM_CALL_IPV6_TO_ENDPOINT		14
#define CILIUM_CALL_IPV4_NODEPORT_NAT		15
#define CILIUM_CALL_IPV6_NODEPORT_NAT		16
#define CILIUM_CALL_IPV4_NODEPORT_REVNAT	17
#define CILIUM_CALL_IPV6_NODEPORT_REVNAT	18
#define CILIUM_CALL_IPV4_ENCAP_NODEPORT_NAT	19
#define CILIUM_CALL_IPV4_NODEPORT_DSR		20
#define CILIUM_CALL_IPV6_NODEPORT_DSR		21
#define CILIUM_CALL_IPV4_FROM_HOST		22
#define CILIUM_CALL_IPV6_FROM_HOST		23
#define CILIUM_CALL_IPV6_ENCAP_NODEPORT_NAT	24
#define CILIUM_CALL_SIZE			25

#define IS_ERR(x) (unlikely((x < 0) || (x == TC_ACT_SHOT)))

enum {
    POLICY_MATCH_NONE = 0,
    POLICY_MATCH_L3_ONLY = 1,
    POLICY_MATCH_L3_L4 = 2,
    POLICY_MATCH_L4_ONLY = 3,
    POLICY_MATCH_ALL = 4,
};

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(max_entries, 128);
//    __uint(max_entries, __NR_CPUS__);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(__u32));
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} cilium_events SEC(".maps");

struct ipv4_ct_tuple {
    /* Address fields are reversed, i.e.,
     * these field names are correct for reply direction traffic.
     */
    __be32		daddr;
    __be32		saddr;
    /* The order of dport+sport must not be changed!
     * These field names are correct for original direction traffic.
     */
    __be16		dport;
    __be16		sport;
    __u8		nexthdr;
    __u8		flags;
} __packed;

struct ct_entry {
    __u64 rx_packets;
    __u64 rx_bytes;
    __u64 tx_packets;
    __u64 tx_bytes;
    __u32 lifetime;
    __u16 rx_closing:1,
            tx_closing:1,
            nat46:1,
            lb_loopback:1,
            seen_non_syn:1,
            node_port:1,
            proxy_redirect:1, /* Connection is redirected to a proxy */
    dsr:1,
            reserved:8;
    __u16 rev_nat_index;
    /* In the kernel ifindex is u32, so we need to check in cilium-agent
     * that ifindex of a NodePort device is <= MAX(u16).
     */
    __u16 ifindex;

    /* *x_flags_seen represents the OR of all TCP flags seen for the
     * transmit/receive direction of this entry.
     */
    __u8  tx_flags_seen;
    __u8  rx_flags_seen;

    __u32 src_sec_id; /* Used from userspace proxies, do not change offset! */

    /* last_*x_report is a timestamp of the last time a monitor
     * notification was sent for the transmit/receive direction.
     */
    __u32 last_tx_report;
    __u32 last_rx_report;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 4096);
    __uint(key_size, sizeof(struct ipv4_ct_tuple));
    __uint(value_size, sizeof(struct ct_entry));
    __uint(pinning, LIBBPF_PIN_BY_NAME);
#ifndef HAVE_LRU_HASH_MAP_TYPE
    __uint(map_flags, BPF_F_NO_PREALLOC);
#endif
} cilium_ct_tcp4 SEC(".maps");

/* Structure representing an IPv4 or IPv6 address, being used for:
 *  - key as endpoints map
 *  - key for tunnel endpoint map
 *  - value for tunnel endpoint map
 */
struct endpoint_key {
    union {
        struct {
            __u32		ip4;
            __u32		pad1;
            __u32		pad2;
            __u32		pad3;
        };
        union v6addr	ip6;
    };
    __u8 family; // ipv4/ipv6
    __u8 key;
    __u16 pad5;
} __packed;

/* Value of endpoint map */
struct endpoint_info {
    __u32		ifindex;
    __u16		unused; /* used to be sec_label, no longer used */
    __u16       lxc_id;
    __u32		flags;
    mac_t		mac; // 容器侧 eth0 网卡 mac
    mac_t		node_mac; // host 侧 lxc 网卡 mac
    __u32		pad[4];
};

// `cilium bpf endpoint list`
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(key_size, sizeof(struct endpoint_key));
    __uint(value_size, sizeof(struct endpoint_info));
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(max_entries, 65536);
    __uint(map_flags, BPF_F_NO_PREALLOC);
} cilium_lxc SEC(".maps");

struct ipcache_key {
    struct bpf_lpm_trie_key lpm_key;
    __u16 pad1;
    __u8 pad2;
    __u8 family;
    union {
        struct {
            __u32		ip4;
            __u32		pad4;
            __u32		pad5;
            __u32		pad6;
        };
        union v6addr	ip6;
    };
} __packed;

struct remote_endpoint_info {
    __u32		sec_label;
    __u32		tunnel_endpoint;
    __u8		key;
};

struct {
    __uint(type, BPF_MAP_TYPE_LPM_TRIE);
    __uint(key_size, sizeof(struct ipcache_key));
    __uint(value_size, sizeof(struct remote_endpoint_info));
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(max_entries, 512000);
    __uint(map_flags, BPF_F_NO_PREALLOC);
} cilium_ipcache SEC(".maps");

struct policy_key {
    __u32		sec_label;
    __u16		dport;
    __u8		protocol;
    __u8		egress:1,
            pad:7;
};

struct policy_entry {
    __be16		proxy_port;
    __u8		deny:1,
            pad:7;
    __u8		pad0;
    __u16		pad1;
    __u16		pad2;
    __u64		packets;
    __u64		bytes;
};

/* Private per EP map for internal tail calls */
#define POLICY_MAP_SIZE 16384
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(key_size, sizeof(struct policy_key));
    __uint(value_size, sizeof(struct policy_entry));
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(max_entries, POLICY_MAP_SIZE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
} cilium_policy SEC(".maps");

struct ct_state {
    __u16 rev_nat_index;
    __u16 loopback:1,
            node_port:1,
            proxy_redirect:1, /* Connection is redirected to a proxy */
    dsr:1,
            reserved:12;
    __be32 addr;
    __be32 svc_addr;
    __u32 src_sec_id;
    __u16 ifindex;
    __u16 backend_id;	/* Backend ID in lb4_backends */
};


static __always_inline int redirect_ep(struct __sk_buff *ctx __maybe_unused,
                                       int ifindex __maybe_unused,
                                       bool needs_backlog __maybe_unused)
{
    /* Going via CPU backlog queue (aka needs_backlog) is required
     * whenever we cannot do a fast ingress -> ingress switch but
     * instead need an ingress -> egress netns traversal or vice
     * versa.
     */
//    if (needs_backlog || !is_defined(ENABLE_HOST_ROUTING)) {
    if (needs_backlog) {
        return (int) bpf_redirect(ifindex, 0);
    } else {
//# ifdef ENCAP_IFINDEX
        /* When coming from overlay, we need to set packet type
		 * to HOST as otherwise we might get dropped in IP layer.
		 */
        bpf_skb_change_type(ctx, PACKET_HOST);

//# endif /* ENCAP_IFINDEX */

        return (int) bpf_redirect_peer(ifindex, 0);
    }
}

/**
 * set_identity_mark - pushes 24 bit identity into ctx mark value.
 */
static __always_inline __maybe_unused void
set_identity_mark(struct __sk_buff *ctx, __u32 identity)
{
    ctx->mark = ctx->mark & MARK_MAGIC_KEY_MASK;
    ctx->mark |= ((identity & 0xFFFF) << 16) | ((identity & 0xFF0000) >> 16);
}

static __always_inline __maybe_unused void
skb_store_meta(struct __sk_buff *ctx, const __u32 off, __u32 data)
{
    ctx->cb[off] = data;
}

static __always_inline __maybe_unused __u32
skb_load_meta(const struct __sk_buff *ctx, const __u32 off)
{
    return ctx->cb[off];
}

static __always_inline bool ctx_skip_host_fw(struct __sk_buff *ctx) {
    volatile __u32 tc_index = ctx->tc_index;
    ctx->tc_index &= ~TC_INDEX_F_SKIP_HOST_FIREWALL;
    return tc_index & TC_INDEX_F_SKIP_HOST_FIREWALL;
}




#ifndef SKIP_CALLS_MAP
static __always_inline void 
ep_tail_call(struct __ctx_buff *ctx, const __u32 index) {
	tail_call_static(ctx, &CALLS_MAP, index);
}
#endif /* SKIP_CALLS_MAP */



#endif
