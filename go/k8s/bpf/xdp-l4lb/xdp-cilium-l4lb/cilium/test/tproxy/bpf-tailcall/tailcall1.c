

/**
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/tailcall1.c
 * https://docs.cilium.io/en/latest/bpf/architecture/#tail-calls
 */

#include <stddef.h>

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

static const int cfg_port = 1234;

struct {
    __uint(type, BPF_MAP_TYPE_PROG_ARRAY); // tail_call 专属的 map type: BPF_MAP_TYPE_PROG_ARRAY
    __uint(max_entries, 3);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(__u32));
} jmp_table SEC(".maps");


// cat /sys/kernel/debug/tracing/trace_pipe
// tcpdump -i lo -nneevv port 1234

SEC("classifier/0")
int bpf_func_0(struct __sk_buff *skb) {
    bpf_printk("bpf_func_0");
    return TC_ACT_OK;
}

SEC("classifier/1")
int bpf_func_1(struct __sk_buff *skb) {
    bpf_printk("bpf_func_1");
    return TC_ACT_OK;
}

SEC("classifier/2")
int bpf_func_2(struct __sk_buff *skb) {
    bpf_printk("bpf_func_2");
    return TC_ACT_SHOT; // 如果 jmp_table[2] 有值，这里是 TC_ACT_SHOT, server 会一直拒绝 syn 包，client 就会一直发 syn 包来 tcp connection
}

SEC("classifier")
int entry(struct __sk_buff *skb) {
    void *data_end = (void *) (long) skb->data_end;
    void *data = (void *) (long) skb->data;
    struct ethhdr *eth = data;
    __u32 ethdr_off;

    ethdr_off = sizeof(struct ethhdr);
    if (data + ethdr_off > data_end) {
        return TC_ACT_SHOT;
    }

    if (eth->h_proto != bpf_htons(ETH_P_IP)) { // ipv4
        return TC_ACT_OK;
    }

    struct iphdr *iph;
    iph = data + ethdr_off;
    if ((void *) (iph + 1) > data_end) {
        return TC_ACT_OK;
    }
    if (iph->ihl != 5) { // 5<=ip4->ihl<=15, ihl(ip header length) 必须是 5(0101)
        return TC_ACT_OK;
    }

    __u8 protocol;
    protocol = iph->protocol;
    if (protocol != IPPROTO_TCP) {
//        bpf_printk("not tcp");
        return TC_ACT_OK;
    }

    struct tcphdr *tcp = (struct tcphdr *) (iph + 1);
    if ((void *) (tcp + 1) > data_end) {
        return TC_ACT_OK;
    }

    // 这个赋值方式不行!!!
//    tcp = (struct tcphdr *) data;

//    if (!tcp->syn) {
//        return TC_ACT_OK;
//    }

    if (tcp->dest != bpf_htons(cfg_port)) {
//        bpf_printk("not dst port %d", tcp->dest); // "not dst port 65221"
        return TC_ACT_OK;
    }

    /**
     * bpf_tail_call_static_before
     * bpf_tail_call_static 意思是，jmp_table[i] 没有对应的 program，则继续找，有则进入 jmp_table[i] bpf 程序，后续的程序逻辑不管
     * 可以参考文档查看 tail_call 逻辑: https://docs.cilium.io/en/latest/bpf/architecture/#tail-calls
     *
     * 每一个 ->1234 包，都会打印，只运行进入第一个 bpf_tail_call_static()，如果没有则进入第二个 bpf_tail_call_static(), 而不是依次运行 bpf_tail_call_static()
     * bpf_trace_printk: bpf_tail_call_static_before
       bpf_trace_printk: bpf_func_0
     *
     */

    bpf_printk("bpf_tail_call_static_before");

    bpf_tail_call_static(skb, &jmp_table, 0);
    bpf_tail_call_static(skb, &jmp_table, 0);
    bpf_tail_call_static(skb, &jmp_table, 0);
    bpf_tail_call_static(skb, &jmp_table, 0);

    bpf_tail_call_static(skb, &jmp_table, 1);
    bpf_tail_call_static(skb, &jmp_table, 1);
    bpf_tail_call_static(skb, &jmp_table, 1);
    bpf_tail_call_static(skb, &jmp_table, 1);

    bpf_tail_call_static(skb, &jmp_table, 2);
    bpf_tail_call_static(skb, &jmp_table, 2);
    bpf_tail_call_static(skb, &jmp_table, 2);
    bpf_tail_call_static(skb, &jmp_table, 2);

    bpf_printk("bpf_tail_call_static_after");
    return TC_ACT_OK;
}


char __license[] SEC("license") = "GPL";
int _version SEC("version") = 1;
