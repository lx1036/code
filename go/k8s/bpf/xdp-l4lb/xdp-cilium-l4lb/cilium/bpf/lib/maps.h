
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

#define IS_ERR(x) (unlikely((x < 0) || (x == TC_ACT_SHOT)))

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
    mac_t		mac;
    mac_t		node_mac;
    __u32		pad[4];
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(key_size, sizeof(struct endpoint_key));
    __uint(value_size, sizeof(struct endpoint_info));
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(max_entries, 65536);
    __uint(map_flags, CONDITIONAL_PREALLOC);
} endpoints SEC(".maps");


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

//
//
//#ifndef SKIP_CALLS_MAP
//
//// CALLS_MAP 在 xdp.go 里定义为 "cilium_calls_xdp" map
//
///* Private per EP map for internal tail calls */
//struct bpf_elf_map __section_maps CALLS_MAP = {
//	.type		= BPF_MAP_TYPE_PROG_ARRAY,
//	.id		= CILIUM_MAP_CALLS,
//	.size_key	= sizeof(__u32),
//	.size_value	= sizeof(__u32),
//	.pinning	= PIN_GLOBAL_NS,
//	.max_elem	= CILIUM_CALL_SIZE,
//};
//#endif /* SKIP_CALLS_MAP */
//
//
//struct ipcache_key {
//	struct bpf_lpm_trie_key lpm_key;
//	__u16 pad1;
//	__u8 pad2;
//	__u8 family;
//	union {
//		struct {
//			__u32		ip4;
//			__u32		pad4;
//			__u32		pad5;
//			__u32		pad6;
//		};
//		union v6addr	ip6;
//	};
//} __packed;
//
///* Global IP -> Identity map for applying egress label-based policy */
//// 实际上在用户态里定义为 "cilium_ipcache" bpf map
//struct bpf_elf_map __section_maps IPCACHE_MAP = {
//	.type		= LPM_MAP_TYPE,
//	.size_key	= sizeof(struct ipcache_key),
//	.size_value	= sizeof(struct remote_endpoint_info),
//	.pinning	= PIN_GLOBAL_NS,
//	.max_elem	= IPCACHE_MAP_SIZE,
//	.flags		= BPF_F_NO_PREALLOC,
//};



#ifndef SKIP_CALLS_MAP
static __always_inline void 
ep_tail_call(struct __ctx_buff *ctx, const __u32 index) {
	tail_call_static(ctx, &CALLS_MAP, index);
}
#endif /* SKIP_CALLS_MAP */



#endif
