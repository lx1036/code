//
// Created by 刘祥 on 7/5/22.
//


#include "section.h"
#include "compiler.h"
#include "../lib/common.h"
#include "../node_config.h"

#include "bpf_sockops.h"


#ifdef ENABLE_IPV4
static __always_inline void sk_extract4_key(const struct bpf_sock_ops *ops, struct sock_key *key)
{
    key->dip4 = ops->remote_ip4;
	key->sip4 = ops->local_ip4;
	key->family = ENDPOINT_KEY_IPV4;
    key->sport = (bpf_ntohl(ops->local_port) >> 16);
	/* clang-7.1 or higher seems to think it can do a 16-bit read here
	 * which unfortunately most kernels (as of October 2019) do not
	 * support, which leads to verifier failures. Insert a READ_ONCE
	 * to make sure that a 32-bit read followed by shift is generated.
	 */
	key->dport = READ_ONCE(ops->remote_port) >> 16;
}

static __always_inline void sk_lb4_key(struct lb4_key *lb4, const struct sock_key *key)
{
	/* SK MSG is always egress, so use daddr */
	lb4->address = key->dip4;
	lb4->dport = key->dport;
}

static inline void bpf_sock_ops_ipv4(struct bpf_sock_ops *skops)
{
    struct lb4_key lb4_key = {};
	__u32 dip4, dport, dstID = 0;
	struct endpoint_info *exists;
	struct lb4_service *svc;
	struct sock_key key = {};
	int verdict;

    sk_extract4_key(skops, &key);

    // 如果目的地址 ip:port 是 service ip，则跳过进入下一个 hook 处理包
    sk_lb4_key(&lb4_key, &key);
	svc = lb4_lookup_service(&lb4_key, true);
	if (svc)
		return;

    dip4 = key.dip4;
	dport = key.dport;
	key.dip4 = key.sip4;
	key.dport = key.sport;
	key.sip4 = dip4;
	key.sport = dport;

    sock_hash_update(skops, &SOCK_OPS_MAP, &key, BPF_NOEXIST);
}

#endif /* ENABLE_IPV4 */


#ifdef ENABLE_IPV6
static inline void bpf_sock_ops_ipv6(struct bpf_sock_ops *skops)
{
	if (skops->remote_ip4)
		bpf_sock_ops_ipv4(skops);
}
#endif /* ENABLE_IPV6 */


// INFO: 监听 socket 事件，然后更新 socket map
// BPF 程序一：监听 socket 事件，更新 sockmap

__section("sockops")
int bpf_sockmap(struct bpf_sock_ops *skops)
{
    __u32 family, op;

    family = skops->family;
    op = skops->op;

    // 对于两端都在本节点的 socket 来说，这段代码会执行两次：
    // (1)源端发送 SYN 时会产生一个事件，命中 case 2
    // (2)目的端发送 SYN+ACK 时会产生一个事件，命中 case 1

    switch (op) {
    case BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB: // 被动建连
    case BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB: // 主动建连
#ifdef ENABLE_IPV6
        if (family == AF_INET6)
            bpf_sock_ops_ipv6(skops);
#endif
#ifdef ENABLE_IPV4
        if (family == AF_INET) // AF_INET 是 ipv4 包
            bpf_sock_ops_ipv4(skops);
#endif
        break;
    default:
        break;
    }

    return 0;
}


BPF_LICENSE("GPL");
int _version __section("version") = 1;

