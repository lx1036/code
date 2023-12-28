


#ifndef _UAPI_LINUX_IP_H
#define _UAPI_LINUX_IP_H

#include <linux/types.h>

// IPv4 header
// #include <linux/ip.h>
struct iphdr {
  __u8 ihl : 4;
  __u8 version : 4;
  __u8 tos;
  __u16 tot_len;
  __u16 id;
  __u16 frag_off;
  __u8 ttl;
  __u8 protocol;
  __u16 check;
  __u32 saddr;
  __u32 daddr;
} __attribute__((packed));



#endif
