

#include <sys/socket.h>

#include <linux/bpf.h>
#include <linux/in.h>
#include <linux/tcp.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#ifndef PAGE_SIZE
#define PAGE_SIZE 4096
#endif

#define SOL_CUSTOM            0xdeadbeef
/**
 * /root/linux-5.10.142/include/linux/socket.h
 */
/* Setsockoptions(2) level. Thanks to BSD these must match IPPROTO_xxx */
#define SOL_IP		0
#define SOL_TCP		6
#define SOL_UDP		17


struct sockopt_sk {
    __u8 val;
};

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, int);
    __type(value, struct sockopt_sk);
} socket_storage_map SEC(".maps");

// EPERM: error permission

/**
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/sockopt_sk.c
 * /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/sockopt_sk.c
 */


SEC("cgroup/getsockopt")
int getsockopt1(struct bpf_sockopt *ctx) {
    struct sockopt_sk *storage;
    __u8 *optval_end = ctx->optval_end;
    __u8 *optval = ctx->optval;

    if (optval + 1 > optval_end) {
        return 0; /* EPERM, bounds check */
    }

    if (ctx->level == SOL_IP && ctx->optname == IP_TOS) {
        /* Not interested in SOL_IP:IP_TOS;
         * let next BPF program in the cgroup chain or kernel
         * handle it.
         */
        ctx->optlen = 0; /* bypass optval>PAGE_SIZE */
        return 1;
    }

    if (ctx->level == SOL_SOCKET && ctx->optname == SO_SNDBUF) {
        /* Not interested in SOL_SOCKET:SO_SNDBUF;
         * let next BPF program in the cgroup chain or kernel
         * handle it.
         */
        return 1;
    }

    if (ctx->level == SOL_TCP && ctx->optname == TCP_CONGESTION) {
        /* Not interested in SOL_TCP:TCP_CONGESTION;
         * let next BPF program in the cgroup chain or kernel
         * handle it.
         */
        return 1;
    }

    if (ctx->level == SOL_IP && ctx->optname == IP_FREEBIND) {
        if (optval + 1 > optval_end)
            return 0; /* EPERM, bounds check */

        ctx->retval = 0; /* Reset system call return value to zero */

        /* Always export 0x55 */
        optval[0] = 0x55;
        ctx->optlen = 1;

        /* Userspace buffer is PAGE_SIZE * 2, but BPF
         * program can only see the first PAGE_SIZE
         * bytes of data.
         */
        if (optval_end - optval != PAGE_SIZE)
            return 0; /* EPERM, unexpected data size */

        return 1;
    }

    if (ctx->level != SOL_CUSTOM) {
        return 0; /* EPERM, deny everything except custom level */
    }

    storage = bpf_sk_storage_get(&socket_storage_map, ctx->sk, 0, BPF_SK_STORAGE_GET_F_CREATE);
    if (!storage) {
        return 0; /* EPERM, couldn't get sk storage */
    }

    if (!ctx->retval) {
        return 0; /* EPERM, kernel should not have handled SOL_CUSTOM, something is wrong! */
    }

    ctx->retval = 0; /* Reset system call return value to zero */
    optval[0] = storage->val;
    ctx->optlen = 1;

    return 1;
}

SEC("cgroup/setsockopt")
int setsockopt1(struct bpf_sockopt *ctx) {
    struct sockopt_sk *storage;
    __u8 *optval_end = ctx->optval_end;
    __u8 *optval = ctx->optval;

    if (optval + 1 > optval_end) {
        return 0; /* EPERM, bounds check */
    }

    if (ctx->level == SOL_IP && ctx->optname == IP_TOS) {
        /* Not interested in SOL_IP:IP_TOS;
         * let next BPF program in the cgroup chain or kernel
         * handle it.
         */
        ctx->optlen = 0; /* bypass optval>PAGE_SIZE */
        return 1;
    }

    if (ctx->level == SOL_SOCKET && ctx->optname == SO_SNDBUF) {
        /* Overwrite SO_SNDBUF value */
        if (optval + sizeof(__u32) > optval_end)
            return 0; /* EPERM, bounds check */

        *(__u32 *)optval = 0x55AA;
        ctx->optlen = 4;

        return 1;
    }

    if (ctx->level == SOL_TCP && ctx->optname == TCP_CONGESTION) {
        /* Always use cubic */
        if (optval + 5 > optval_end)
            return 0; /* EPERM, bounds check */

        __builtin_memcpy(optval, "cubic", 5);
        ctx->optlen = 5;

        return 1;
    }

    if (ctx->level == SOL_IP && ctx->optname == IP_FREEBIND) {
        /* Original optlen is larger than PAGE_SIZE. */
        if (ctx->optlen != PAGE_SIZE * 2)
            return 0; /* EPERM, unexpected data size */

        if (optval + 1 > optval_end)
            return 0; /* EPERM, bounds check */

        /* Make sure we can trim the buffer. */
        optval[0] = 0;
        ctx->optlen = 1;

        /* Usepace buffer is PAGE_SIZE * 2, but BPF
         * program can only see the first PAGE_SIZE
         * bytes of data.
         */
        if (optval_end - optval != PAGE_SIZE)
            return 0; /* EPERM, unexpected data size */

        return 1;
    }

    if (ctx->level != SOL_CUSTOM) {
        return 0; /* EPERM, deny everything except custom level */
    }

    storage = bpf_sk_storage_get(&socket_storage_map, ctx->sk, 0, BPF_SK_STORAGE_GET_F_CREATE);
    if (!storage) {
        return 0; /* EPERM, couldn't get sk storage */
    }

    storage->val = optval[0];
    ctx->optlen = -1; /* BPF has consumed this option, don't call kernel setsockopt handler.*/

    return 1;
}


#define CUSTOM_INHERIT1            0
#define CUSTOM_INHERIT2            1
#define CUSTOM_LISTENER            2

struct sockopt_inherit {
    __u8 val;
};

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC | BPF_F_CLONE);
    __type(key, int);
    __type(value, struct sockopt_inherit);
} cloned1_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC | BPF_F_CLONE);
    __type(key, int);
    __type(value, struct sockopt_inherit);
} cloned2_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, int);
    __type(value, struct sockopt_inherit);
} listener_only_map SEC(".maps");

static __always_inline struct sockopt_inherit *get_storage(struct bpf_sockopt *ctx) {
    if (ctx->optname == CUSTOM_INHERIT1)
        return bpf_sk_storage_get(&cloned1_map, ctx->sk, 0, BPF_SK_STORAGE_GET_F_CREATE);
    else if (ctx->optname == CUSTOM_INHERIT2)
        return bpf_sk_storage_get(&cloned2_map, ctx->sk, 0, BPF_SK_STORAGE_GET_F_CREATE);
    else
        return bpf_sk_storage_get(&listener_only_map, ctx->sk, 0, BPF_SK_STORAGE_GET_F_CREATE);
}

/**
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/sockopt_inherit.c
 * /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/sockopt_inherit.c
 */

SEC("cgroup/getsockopt")
int getsockopt2(struct bpf_sockopt *ctx) {
    struct sockopt_inherit *storage;
    __u8 *optval_end = ctx->optval_end;
    __u8 *optval = ctx->optval;

    if (optval + 1 > optval_end) {
        return 0; /* EPERM, bounds check */
    }

    if (ctx->level != SOL_CUSTOM) {
        return 1; /* only interested in SOL_CUSTOM */
    }

    storage = get_storage(ctx);
    if (!storage) {
        return 0; /* EPERM, couldn't get sk storage */
    }

    ctx->retval = 0; /* Reset system call return value to zero, ??? */
    optval[0] = storage->val;
    ctx->optlen = 1;

    return 1;
}

SEC("cgroup/setsockopt")
int setsockopt2(struct bpf_sockopt *ctx) {
    struct sockopt_inherit *storage;
    __u8 *optval_end = ctx->optval_end;
    __u8 *optval = ctx->optval;

    if (optval + 1 > optval_end) {
        return 0; /* EPERM, bounds check */
    }

    if (ctx->level != SOL_CUSTOM) {
        return 1; /* only interested in SOL_CUSTOM */
    }

    storage = get_storage(ctx);
    if (!storage) {
        return 0; /* EPERM, couldn't get sk storage */
    }

    // INFO: 注意获取 opt val 的方式
    storage->val = optval[0];
    ctx->optlen = -1; // 注意 -1

    return 1;
}

char _license[] SEC("license") = "GPL";
