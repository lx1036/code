



#ifndef __LIB_COMMON_H_
#define __LIB_COMMON_H_


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


/* Value of endpoint map */
struct endpoint_info {
	__u32		ifindex;
	__u16		unused; /* used to be sec_label, no longer used */
	__u16           lxc_id;
	__u32		flags;
	mac_t		mac;
	mac_t		node_mac;
	__u32		pad[4];
};





#include "overloadable.h"

#endif /* __LIB_COMMON_H_ */