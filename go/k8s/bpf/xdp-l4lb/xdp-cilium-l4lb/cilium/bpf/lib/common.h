



#ifndef __LIB_COMMON_H_
#define __LIB_COMMON_H_


#include <bpf/ctx/ctx.h>
#include <bpf/api.h>

#include <linux/if_ether.h>
#include <linux/ipv6.h>
#include <linux/in.h>
#include <linux/socket.h>

#include "eth.h"
#include "endian.h"
#include "mono.h"
#include "config.h"


#define CT_EGRESS 0
#define CT_INGRESS 1
#define CT_SERVICE 2

#define TUPLE_F_OUT		0	/* Outgoing flow */
#define TUPLE_F_IN		1	/* Incoming flow */
#define TUPLE_F_RELATED		2	/* Flow represents related packets */
#define TUPLE_F_SERVICE		4	/* Flow represents packets to service */



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

/* Cilium error codes, must NOT overlap with TC return codes.
 * These also serve as drop reasons for metrics,
 * where reason > 0 corresponds to -(DROP_*)
 *
 * These are shared with pkg/monitor/api/drop.go and api/v1/flow/flow.proto.
 * When modifying any of the below, those files should also be updated.
 */
#define DROP_UNUSED1		-130 /* unused */
#define DROP_UNUSED2		-131 /* unused */
#define DROP_INVALID_SIP	-132
#define DROP_POLICY		-133
#define DROP_INVALID		-134
#define DROP_CT_INVALID_HDR	-135
#define DROP_FRAG_NEEDED	-136
#define DROP_CT_UNKNOWN_PROTO	-137
#define DROP_UNUSED4		-138 /* unused */
#define DROP_UNKNOWN_L3		-139
#define DROP_MISSED_TAIL_CALL	-140
#define DROP_WRITE_ERROR	-141
#define DROP_UNKNOWN_L4		-142
#define DROP_UNKNOWN_ICMP_CODE	-143
#define DROP_UNKNOWN_ICMP_TYPE	-144
#define DROP_UNKNOWN_ICMP6_CODE	-145
#define DROP_UNKNOWN_ICMP6_TYPE	-146
#define DROP_NO_TUNNEL_KEY	-147
#define DROP_UNUSED5		-148 /* unused */
#define DROP_UNUSED6		-149 /* unused */
#define DROP_UNKNOWN_TARGET	-150
#define DROP_UNROUTABLE		-151
#define DROP_UNUSED7		-152 /* unused */
#define DROP_CSUM_L3		-153
#define DROP_CSUM_L4		-154
#define DROP_CT_CREATE_FAILED	-155
#define DROP_INVALID_EXTHDR	-156
#define DROP_FRAG_NOSUPPORT	-157
#define DROP_NO_SERVICE		-158
#define DROP_UNUSED8		-159 /* unused */
#define DROP_NO_TUNNEL_ENDPOINT -160
#define DROP_UNUSED9		-161 /* unused */
#define DROP_EDT_HORIZON	-162
#define DROP_UNKNOWN_CT		-163
#define DROP_HOST_UNREACHABLE	-164
#define DROP_NO_CONFIG		-165
#define DROP_UNSUPPORTED_L2	-166
#define DROP_NAT_NO_MAPPING	-167
#define DROP_NAT_UNSUPP_PROTO	-168
#define DROP_NO_FIB		-169
#define DROP_ENCAP_PROHIBITED	-170
#define DROP_INVALID_IDENTITY	-171
#define DROP_UNKNOWN_SENDER	-172
#define DROP_NAT_NOT_NEEDED	-173 /* Mapped as drop code, though drop not necessary. */
#define DROP_IS_CLUSTER_IP	-174
#define DROP_FRAG_NOT_FOUND	-175
#define DROP_FORBIDDEN_ICMP6	-176
#define DROP_NOT_IN_SRC_RANGE	-177
#define DROP_PROXY_LOOKUP_FAILED	-178
#define DROP_PROXY_SET_FAILED	-179
#define DROP_PROXY_UNKNOWN_PROTO	-180
#define DROP_POLICY_DENY	-181

/* Lookup scope for externalTrafficPolicy=Local */
#define LB_LOOKUP_SCOPE_EXT	0
#define LB_LOOKUP_SCOPE_INT	1 // local

#define SRC_RANGE_STATIC_PREFIX(STRUCT)		\
	(8 * (sizeof(STRUCT) - sizeof(struct bpf_lpm_trie_key)))

#ifdef PREALLOCATE_MAPS
#define CONDITIONAL_PREALLOC 0
#else
#define CONDITIONAL_PREALLOC BPF_F_NO_PREALLOC
#endif


static __always_inline __maybe_unused bool
____revalidate_data_pull(struct __ctx_buff *ctx, void **data_, void **data_end_,
                         void **l3, const __u32 l3_len, const bool pull,
                         __u8 eth_hlen)
{
    const __u64 tot_len = eth_hlen + l3_len;
    void *data_end;
    void *data;

    /* Verifier workaround, do this unconditionally: invalid size of register spill. */
    if (pull)
        ctx_pull_data(ctx, tot_len);
    data_end = ctx_data_end(ctx);
    data = ctx_data(ctx);
    if (data + tot_len > data_end)
        return false;

    /* Verifier workaround: pointer arithmetic on pkt_end prohibited. */
    *data_ = data;
    *data_end_ = data_end;

    *l3 = data + eth_hlen;
    return true;
}

static __always_inline __maybe_unused bool
__revalidate_data_pull(struct __ctx_buff *ctx, void **data, void **data_end,
                       void **l3, const __u32 l3_len, const bool pull)
{
    return ____revalidate_data_pull(ctx, data, data_end, l3, l3_len, pull, ETH_HLEN);
}

/* revalidate_data() initializes the provided pointers from the ctx.
 * Returns true if 'ctx' is long enough for an IP header of the provided type,
 * false otherwise.
 */
#define revalidate_data(ctx, data, data_end, ip)			\
	__revalidate_data_pull(ctx, data, data_end, (void **)ip, sizeof(**ip), false)



/* Service flags (lb{4,6}_service->flags) */
enum {
	SVC_FLAG_EXTERNAL_IP  = (1 << 0),  /* External IPs */
	SVC_FLAG_NODEPORT     = (1 << 1),  /* NodePort service */
	SVC_FLAG_LOCAL_SCOPE  = (1 << 2),  /* externalTrafficPolicy=Local */
	SVC_FLAG_HOSTPORT     = (1 << 3),  /* hostPort forwarding */
	SVC_FLAG_AFFINITY     = (1 << 4),  /* sessionAffinity=clientIP */
	SVC_FLAG_LOADBALANCER = (1 << 5),  /* LoadBalancer service */
	SVC_FLAG_ROUTABLE     = (1 << 6),  /* Not a surrogate/ClusterIP entry */
	SVC_FLAG_SOURCE_RANGE = (1 << 7),  /* Check LoadBalancer source range */
};

enum {
    CT_NEW,
    CT_ESTABLISHED,
    CT_REPLY,
    CT_RELATED,
    CT_REOPENED,
};

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


union v6addr {
	struct {
		__u32 p1;
		__u32 p2;
		__u32 p3;
		__u32 p4;
	};
	struct {
		__u64 d1;
		__u64 d2;
	};
	__u8 addr[16];
} __packed;

struct lb4_key {
	__be32 address;		/* Service virtual IPv4 address */
	__be16 dport;		/* L4 port filter, if unset, all ports apply */
	__u16 backend_slot;	/* Backend iterator, 0 indicates the svc frontend */
	__u8 proto;		/* L4 protocol, currently not used (set to 0) */
	__u8 scope;		/* LB_LOOKUP_SCOPE_* for externalTrafficPolicy=Local */
	__u8 pad[2];
};

struct lb4_service {
	union {
		__u32 backend_id;		/* Backend ID in lb4_backends */
		__u32 affinity_timeout;		/* In seconds, only for svc frontend */
	};
	/* For the service frontend, count denotes number of service backend
	 * slots (otherwise zero).
	 */
	__u16 count; // 判断 svc 有没有 backend
	__u16 rev_nat_index;	/* Reverse NAT ID in lb4_reverse_nat */
	__u8 flags; // 实现 k8s service 是不是 external 还是 local
	__u8 flags2;
	__u8  pad[2];
};

union lb4_affinity_client_id {
	__u32 client_ip;
	__net_cookie client_cookie;
} __packed;

struct lb4_affinity_key {
	union lb4_affinity_client_id client_id;
	__u16 rev_nat_id;
	__u8 netns_cookie:1,
	     reserved:7;
	__u8 pad1;
	__u32 pad2;
} __packed;

struct lb_affinity_val {
	__u64 last_used;
	__u32 backend_id;
	__u32 pad;
} __packed;

struct lb_affinity_match {
	__u32 backend_id;
	__u16 rev_nat_id;
	__u16 pad;
} __packed;

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

struct lb4_backend {
    __be32 address;		/* Service endpoint IPv4 address */
    __be16 port;		/* L4 port filter */
    __u8 proto;		/* L4 protocol, currently not used (set to 0) */
    __u8 pad;
};

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

struct lb4_src_range_key {
	struct bpf_lpm_trie_key lpm_key;
	__u16 rev_nat_id;
	__u16 pad;
	__u32 addr;
};

struct remote_endpoint_info {
	__u32		sec_label;
	__u32		tunnel_endpoint;
	__u8		key;
};

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

// 从二层头 ethernet header 中获取 __u16 *protocol，并验证符合二层头协议的包
static __always_inline bool 
validate_ethertype(struct xdp_md *ctx, __u16 *proto) {
	void *data = ctx_data(ctx);
	void *data_end = ctx_data_end(ctx);
	struct ethhdr *eth = data; // 转换成二层头

	if (ETH_HLEN == 0) {
		/* The packet is received on L2-less device. Determine L3
		 * protocol from skb->protocol.
		 */
		*proto = ctx_get_protocol(ctx);
		return true;
	}

	if (data + ETH_HLEN > data_end) // 如果不符合二层头协议的包
		return false;

	*proto = eth->h_proto;
	if (bpf_ntohs(*proto) < ETH_P_802_3_MIN) // bpf_ntohs: 把 __u16->0xXXXX
		return false; /* non-Ethernet II unsupported */
	
	return true;
}

#define IS_ERR(x) (unlikely((x < 0) || (x == CTX_ACT_DROP)))


#include "overloadable.h"

#endif /* __LIB_COMMON_H_ */
