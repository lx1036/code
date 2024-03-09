



#include <sys/socket.h>
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

static volatile int invocations = 0, in_use = 0;

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, int);
    __type(value, int);
} sk_map SEC(".maps");


SEC("cgroup/sock_create")
int sock(struct bpf_sock *ctx)
{
    int *orig;
    if (ctx->type != SOCK_DGRAM)
        return 1;

    orig = bpf_sk_storage_get(&sk_map, ctx, 0, BPF_SK_STORAGE_GET_F_CREATE);
    if (!orig)
        return 0;
    *orig = (int)0xdeadbeef; // 3735928559

    __sync_fetch_and_add(&invocations, 1);

    if (in_use > 0) {
        /* BPF_CGROUP_INET_SOCK_RELEASE is _not_ called
         * when we return an error from the BPF
         * program!
         */
        return 0;
    }

    __sync_fetch_and_add(&in_use, 1);

    bpf_printk("create socket in_use: %d, invocations: %d", in_use, invocations);
    return 1;
}

SEC("cgroup/sock_release")
int sock_release(struct bpf_sock *ctx)
{
    int *orig;
    if (ctx->type != SOCK_DGRAM)
        return 1;

    orig = bpf_sk_storage_get(&sk_map, ctx, 0, 0);
    if (!orig || *orig != 0xdeadbeef)
        return 0;

    __sync_fetch_and_add(&invocations, 1);
    __sync_fetch_and_add(&in_use, -1);

    bpf_printk("release socket in_use: %d, invocations: %d", in_use, invocations);
    return 1;
}

char _license[] SEC("license") = "GPL";
