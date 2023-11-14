



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
	__u16 count;
	__u16 rev_nat_index;	/* Reverse NAT ID in lb4_reverse_nat */
	__u8 flags;
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





#include "overloadable.h"

#endif /* __LIB_COMMON_H_ */
