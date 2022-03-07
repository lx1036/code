


#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <arpa/inet.h>

__attribute__((section("ingress"), used))
int drop(struct __sk_buff *skb) {
    void *data = (void*)(long)skb->data;
    void *data_end = (void*)(long)skb->data_end;

    if (data_end < data + ETH_HLEN)
        return TC_ACT_OK; // Not our packet, return it back to kernel

    struct ethhdr *eth = data;
    if (eth->h_proto != htons(ETH_P_ARP))
       return TC_ACT_OK;

    return TC_ACT_SHOT;
}
