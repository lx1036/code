

#include <linux/bpf.h>
#include <sys/socket.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#ifndef barrier
# define barrier()		asm volatile("": : :"memory")
#endif

static __always_inline void bpf_barrier(void)
{
    /* Workaround to avoid verifier complaint:
     * "dereference of modified ctx ptr R5 off=48+0, ctx+const is allowed,
     *        ctx+const+const is not"
     */
    barrier();
}

#ifndef __READ_ONCE
# define __READ_ONCE(X)		(*(volatile typeof(X) *)&X)
#endif

#ifndef READ_ONCE
# define READ_ONCE(X)						\
			({ typeof(X) __val = __READ_ONCE(X);	\
			   bpf_barrier();			\
			   __val; })
#endif

struct sock_key {
    __u32 sip4;
    __u32 dip4;
    __u8 family;
    __u8  pad1;
    __u16 pad2;
//    // this padding required for 64bit alignment
//    // else ebpf kernel verifier rejects loading
//    // of the program
    __u32 pad3;
    __u32 sport;
    __u32 dport;
//};
} __attribute__((packed));

// `bpftool map dump name sock_ops_map -j | jq`
struct {
    __uint(type, BPF_MAP_TYPE_SOCKHASH);
    __uint(max_entries, 65535);
    __type(key, struct sock_key);
    __type(value, int); // 应该是 __64 才对啊，sock_fd 直接用 int 会不会有问题?
//    __uint(pinning, LIBBPF_PIN_BY_NAME); // 方便调试，先不用 pin map
} sock_ops_map SEC(".maps");

static __always_inline void sk_msg_extract4_key(struct sk_msg_md *msg, struct sock_key *key) {
    key->family = 1;
//    key->sip4 = msg->remote_ip4;
//    key->dip4 = msg->local_ip4;
//    key->sport = msg->remote_port >> 16;
//    key->dport = bpf_htonl(msg->local_port) >> 16;

    key->sip4 = msg->local_ip4;
    key->dip4 = msg->remote_ip4;
    key->sport = bpf_htonl(msg->local_port) >> 16;
    key->dport = (msg->remote_port) >> 16;
}

// hook sendmsg call on a socket, @see SEC("cgroup/sendmsg4")
// INFO: 只有当 tcp 已经建联时，才会调用 sendmsg 发数据报文，然后经过该 ebpf 直接 redirect 对应的 sk，来 bypass TCP/IP netfilter。
//  attach 到 sock_ops_map bpf map, @see /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/sockmap_listen.c
SEC("sk_msg")
int bpf_tcpip_bypass(struct sk_msg_md *msg)
{
    struct sock_key key = {};
    sk_msg_extract4_key(msg, &key);
    // bpf_msg_redirect_map()
    bpf_printk("total size of sk_msg is %d, port %d --> %d", msg->size, bpf_ntohl(msg->remote_port), msg->local_port);
    return (int)bpf_msg_redirect_hash(msg, &sock_ops_map, &key, BPF_F_INGRESS);
//    return SK_PASS;
}

static __always_inline void bpf_sock_ops_ipv4(struct bpf_sock_ops *skops) {
    struct sock_key key = {};
    int ret;

    // keep ip and port in network byte order
    key.family = 1; // 只有指针才是 key->family, @see sk_msg_extract4_key(), 这里为何是 1???
//    key.sip4 = skops->local_ip4; // 为何这里互换???
//    key.dip4 = skops->remote_ip4;
//    // local_port is in host byte order, and remote_port is in network byte order
//    key.sport = (bpf_htonl(skops->local_port) >> 16); // ???
//    /* clang-7.1 or higher seems to think it can do a 16-bit read here
//	 * which unfortunately most kernels (as of October 2019) do not
//	 * support, which leads to verifier failures. Insert a READ_ONCE
//	 * to make sure that a 32-bit read followed by shift is generated.
//	 */
//    key.dport = (skops->remote_port) >> 16;


    key.dip4 = skops->local_ip4;
    key.dport = (bpf_htonl(skops->local_port) >> 16);
    key.sip4 = skops->remote_ip4;
    key.sport = (skops->remote_port) >> 16;

    /**
     * 这里没有报错，但是 map 里为空: `bpftool map dump name sock_ops_map -j | jq`
     */
    ret = (int)bpf_sock_hash_update(skops, &sock_ops_map, &key, BPF_NOEXIST);
    if (ret != 0) {
        bpf_printk("sock_hash_update() failed, ret: %d", ret);
    }

//    __u32 remote_port;    /* Stored in network byte order */ 因为是 network byte order，且是 u32，所以必须 bpf_ntohl()
//    __u32 local_port;    /* stored in host byte order */
    // sockmap: op 4, port 5432 --> 7007, client 端是 5432, server 端是 7007
    // sockmap: op 5, port 7007 --> 5432
    bpf_printk("sockmap: op %d, port %d --> %d", skops->op, skops->local_port, bpf_ntohl(skops->remote_port));
}

SEC("sockops")
int bpf_sockops_v4(struct bpf_sock_ops *skops)
{
    __u32 family, op;
    family = skops->family;
    op = skops->op;
    switch (op) {
    /**
     * active: source socket sending SYN
     * passive: destination socket responding with ACK for the SYN
     */
    case BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB: // 4
    case BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB: // 5
        if (family == AF_INET) { // only ipv4
            bpf_sock_ops_ipv4(skops);
        }
        break;
    default:
        break;
    }

    return SK_PASS;
}



int _version SEC("version") = 1;
char _license[] SEC("license") = "GPL";
