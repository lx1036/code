

/**
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/connect_force_port4.c
 */



#include <sys/socket.h>

#include <linux/bpf.h>
#include <linux/in.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

struct svc_addr {
    __be32 addr;
    __be16 port;
};

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, int);
    __type(value, struct svc_addr);
} service_mapping SEC(".maps");

/**
 * connect4() 在 socket 层面完成 ip:port 转换, 只需要一次转换
 * 1. 不需要逐包的 dnat 行为
 * 2. 不需要逐包的查找 svc 的行为
 */
SEC("cgroup/connect4")
int connect4(struct bpf_sock_addr *ctx) {
    struct sockaddr_in sa = {};
    struct svc_addr *orig;

    // INFO: 在 connect bpf 里去 bind(sockAddr)，当然也可以在 bind bpf 里做

    /* Force local address to 127.0.0.1:22222. */
    sa.sin_family = AF_INET;
    sa.sin_port = bpf_htons(22222); // 注意 port 赋值，__be16=bpf_htons(port)
    sa.sin_addr.s_addr = bpf_htonl(0x7f000001); // 注意 ip 赋值，__be32=bpf_htonl(ip)
    /**
     * This allows for making outgoing connection from the desired IP address, which can be useful for
     * example when all processes inside a cgroup should use one single IP address on a host that has multiple IP configured.
     */
    if (bpf_bind(ctx, (struct sockaddr *) &sa, sizeof(sa)) != 0) {
        return 0; // false
    }

    // 简化 service_map lookup 逻辑和 select backend 逻辑
    /* Rewrite service 1.2.3.4:60000 to backend 127.0.0.1:60123 */
    if (ctx->user_port == bpf_htons(60000)) { // dst port
        // bpf 专门存储 socket storage 数据，存储 dstIP:dstPort 数据，getpeername4() 需要使用
        orig = bpf_sk_storage_get(&service_mapping, ctx->sk, 0, BPF_SK_STORAGE_GET_F_CREATE);
        if (!orig) {
            return 0;
        }

        orig->addr = ctx->user_ip4; // dst ip, 1.2.3.4
        orig->port = ctx->user_port; // dst port, 60000

        ctx->user_ip4 = bpf_htonl(0x7f000001);
        ctx->user_port = bpf_htons(60123);
    }

    return 1; // true
}

SEC("cgroup/getsockname4")
int getsockname4(struct bpf_sock_addr *ctx) {
    /* Expose local server as service 1.2.3.4:60000 to client. */
    if (ctx->user_port == bpf_htons(60123)) {
        ctx->user_ip4 = bpf_htonl(0x01020304);
        ctx->user_port = bpf_htons(60000);
    }

    return 1;
}

SEC("cgroup/getpeername4")
int getpeername4(struct bpf_sock_addr *ctx) {
    struct svc_addr *orig;

    /* Expose service 1.2.3.4:60000 as peer instead of backend 127.0.0.1:60123 */
    if (ctx->user_port == bpf_htons(60123)) {
        orig = bpf_sk_storage_get(&service_mapping, ctx->sk, 0, 0); // 注意这里不用 BPF_SK_STORAGE_GET_F_CREATE
        if (orig) {
            ctx->user_ip4 = orig->addr;
            ctx->user_port = orig->port;
        }
    }

    return 1; // true
}


char _license[] SEC("license") = "GPL";
