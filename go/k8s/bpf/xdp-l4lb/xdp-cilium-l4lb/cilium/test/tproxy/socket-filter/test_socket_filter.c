


#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/tcp.h>
//#include <linux/socket.h> // /root/linux-5.10.142/include/linux/socket.h 不起作用
#include <sys/socket.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

/**
 * https://github.com/dropbox/goebpf/blob/e568275f843160ec86c497dc0d8a2cfccacc9c8c/examples/socket_filter/packet_counter/ebpf_prog/sock_filter.c
 */

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
//    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u64);
} counter_map SEC(".maps");

SEC("socket_filter")
int packet_counter(struct __sk_buff *skb) {
    bpf_printk("test1");
    __u32 idx = 0;
    __u64 *value = bpf_map_lookup_elem(&counter_map, &idx);
    if (value) {
        *value += 1;
    }

    return 1;
}


char _license[] SEC("license") = "GPL";
