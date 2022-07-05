//
// Created by 刘祥 on 7/5/22.
//

#include "section.h"

#include "bpf_sockops.h"
#include "../lib/endian.h"
#include "../lib/common.h"
#include "../lib/eps.h"
#include "bpf/compiler.h"
#include "bpf/helpers.h"
#include "bpf/stddef.h"
#include "../node_config.h"

static __always_inline void sk_msg_extract4_key(const struct sk_msg_md *msg,
                                                struct sock_key *key)
{
    key->dip4 = msg->remote_ip4;
    key->sip4 = msg->local_ip4;
    key->family = ENDPOINT_KEY_IPV4;

    key->sport = (bpf_ntohl(msg->local_port) >> 16);
    /* clang-7.1 or higher seems to think it can do a 16-bit read here
     * which unfortunately most kernels (as of October 2019) do not
     * support, which leads to verifier failures. Insert a READ_ONCE
     * to make sure that a 32-bit read followed by shift is generated.
     */
    key->dport = READ_ONCE(msg->remote_port) >> 16;
}

// BPF 程序二：拦截 sendmsg 系统调用，socket 重定向

__section("sk_msg")
int bpf_redir_proxy(struct sk_msg_md *msg)
{
    struct sock_key key = {};
    struct remote_endpoint_info *info;
    __u32 dstID = 0;
    int verdict;
    __u64 flags = BPF_F_INGRESS;

    sk_msg_extract4_key(msg, &key);

    /* Currently, pulling dstIP out of endpoint
	 * tables. This can be simplified by caching this information with the
	 * socket to avoid extra overhead. This would require the agent though
	 * to flush the sock ops map on policy changes.
	 */
    info = lookup_ip4_remote_endpoint(key.dip4);
    if (info != NULL && info->sec_label)
        dstID = info->sec_label;
    else
        dstID = WORLD_ID;

    // 检查 egress network policy
    verdict = policy_sk_egress(dstID, key.sip4, key.dport);
    if (verdict >= 0)
        msg_redirect_hash(msg, &SOCK_OPS_MAP, &key, flags);
    return SK_PASS;
}

BPF_LICENSE("GPL");
int _version __section("version") = 1;
