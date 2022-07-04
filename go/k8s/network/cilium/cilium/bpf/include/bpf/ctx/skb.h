//
// Created by 刘祥 on 7/1/22.
//

#ifndef BPF_CTX_SKB_H
#define BPF_CTX_SKB_H

#define __ctx_buff __sk_buff

#define ctx_pull_data		skb_pull_data



// struct sk_buff: https://github.com/torvalds/linux/blob/v5.19-rc4/include/linux/skbuff.h#L850-L1212
// 是包的结构体
/* user accessible mirror of in-kernel sk_buff.
 * new fields can only be added to the end of this structure
 */
struct __sk_buff {
    __u32 len;
    __u32 pkt_type;
    __u32 mark;
    __u32 queue_mapping;
    __u32 protocol;
    __u32 vlan_present;
    __u32 vlan_tci;
    __u32 vlan_proto;
    __u32 priority;
    __u32 ingress_ifindex;
    __u32 ifindex;
    __u32 tc_index; /* traffic control index */
    __u32 cb[5];
    __u32 hash;
    __u32 tc_classid;
    __u32 data;
    __u32 data_end;
    __u32 napi_id;

    /* Accessed by BPF_PROG_TYPE_sk_skb types from here to ... */
    __u32 family;
    __u32 remote_ip4;	/* Stored in network byte order */
    __u32 local_ip4;	/* Stored in network byte order */
    __u32 remote_ip6[4];	/* Stored in network byte order */
    __u32 local_ip6[4];	/* Stored in network byte order */
    __u32 remote_port;	/* Stored in network byte order */
    __u32 local_port;	/* stored in host byte order */
    /* ... here. */

    __u32 data_meta;
};

#endif //BPF_CTX_SKB_H
