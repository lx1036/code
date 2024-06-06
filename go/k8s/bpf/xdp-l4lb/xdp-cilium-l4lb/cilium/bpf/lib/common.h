



#ifndef __LIB_COMMON_H_
#define __LIB_COMMON_H_


#include <bpf/ctx/ctx.h>
#include <bpf/api.h>

#include <linux/if_ether.h>
#include <linux/ipv6.h>
#include <linux/in.h>
//#include <linux/socket.h>

#include "eth.h"
#include "endian.h"
//#include "mono.h"
#include "config.h"


#define CT_EGRESS 0
#define CT_INGRESS 1
#define CT_SERVICE 2

#define TUPLE_F_OUT		0	/* Outgoing flow */
#define TUPLE_F_IN		1	/* Incoming flow */
#define TUPLE_F_RELATED		2	/* Flow represents related packets */
#define TUPLE_F_SERVICE		4	/* Flow represents packets to service */

#ifndef AF_INET
#define AF_INET 2
#endif

#ifndef AF_INET6
#define AF_INET6 10
#endif




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

#ifndef TRACE_PAYLOAD_LEN
#define TRACE_PAYLOAD_LEN 128ULL
#endif

/* Cilium metrics direction for dropping/forwarding packet */
#define METRIC_INGRESS  1
#define METRIC_EGRESS   2
#define METRIC_SERVICE  3

typedef __u64 mac_t;



static __always_inline __maybe_unused bool
____revalidate_data_pull(struct __sk_buff *ctx, void **data_, void **data_end_,
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
__revalidate_data_pull(struct __sk_buff *ctx, void **data, void **data_end,
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

enum {
    CILIUM_NOTIFY_UNSPEC,
    CILIUM_NOTIFY_DROP,
    CILIUM_NOTIFY_DBG_MSG,
    CILIUM_NOTIFY_DBG_CAPTURE,
    CILIUM_NOTIFY_TRACE,
    CILIUM_NOTIFY_POLICY_VERDICT,
    CILIUM_NOTIFY_CAPTURE,
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



struct lb4_backend {
    __be32 address;		/* Service endpoint IPv4 address */
    __be16 port;		/* L4 port filter */
    __u8 proto;		/* L4 protocol, currently not used (set to 0) */
    __u8 pad;
};




struct lb4_src_range_key {
	struct bpf_lpm_trie_key lpm_key;
	__u16 rev_nat_id;
	__u16 pad;
	__u32 addr;
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
