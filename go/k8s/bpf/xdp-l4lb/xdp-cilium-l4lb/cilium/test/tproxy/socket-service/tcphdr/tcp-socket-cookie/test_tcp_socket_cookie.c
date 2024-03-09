

#include <linux/bpf.h>
#include <sys/socket.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>


struct socket_cookie {
    __u64 cookie_key;
    __u32 cookie_value;
};

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, int); // 这里
    __type(value, struct socket_cookie);
} socket_cookies SEC(".maps");

// 调用 connect4() 会触发, 比 BPF_SOCK_OPS_TCP_CONNECT_CB 先触发
SEC("cgroup/connect4")
int set_cookie(struct bpf_sock_addr *ctx)
{
    struct socket_cookie *cookie;
    if (ctx->family != AF_INET || ctx->user_family != AF_INET) { // only ipv4
        return 1;
    }

    cookie = bpf_sk_storage_get(&socket_cookies, ctx->sk, 0, BPF_SK_STORAGE_GET_F_CREATE);
    if (!cookie) {
        return 1;
    }

    cookie->cookie_value = 0xFF;
    cookie->cookie_key = bpf_get_socket_cookie(ctx);

    bpf_printk("cgroup/connect4");
    return 1;
}

// BPF_SOCK_OPS_TCP_CONNECT_CB 后触发
SEC("sockops")
int update_cookie(struct bpf_sock_ops *ctx)
{
    struct bpf_sock *sk;
    struct socket_cookie *cookie;

    if (ctx->family != AF_INET) {
        return 1;
    }

    // Calls BPF program right before an active connection is initialized, 调用 connect4() 会触发
    if (ctx->op != BPF_SOCK_OPS_TCP_CONNECT_CB)
        return 1;

    if (!ctx->sk)
        return 1;

    cookie = bpf_sk_storage_get(&socket_cookies, ctx->sk, 0, 0);
    if (!cookie)
        return 1;

    if (cookie->cookie_key != bpf_get_socket_cookie(ctx))
        return 1;

    // hex((0x0101<<8) | 0xFF) = 0x0101ff
    cookie->cookie_value = (ctx->local_port << 8) | cookie->cookie_value;
    bpf_printk("sockops/BPF_SOCK_OPS_TCP_CONNECT_CB");
    return 1;
}


int _version SEC("version") = 1;
char _license[] SEC("license") = "GPL";
