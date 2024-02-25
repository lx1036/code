


// /root/linux-5.10.142/include/uapi/linux/ip.h


#ifndef _UAPI_LINUX_IP_H
#define _UAPI_LINUX_IP_H

#include <linux/types.h>



/*
 *     0                   1                   2                   3
    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |Version|  IHL  |Type of Service|          Total Length         |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |         Identification        |Flags|      Fragment Offset    |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |  Time to Live |    Protocol   |         Header Checksum       |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                       Source Address                          |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                    Destination Address                        |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
   |                    Options                    |    Padding    |
   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 *
 */

// https://en.wikipedia.org/wiki/IP_in_IP
// https://en.wikipedia.org/wiki/Internet_Protocol_version_4#IHL

// https://datatracker.ietf.org/doc/html/rfc791#section-3.1

// IPv4 header
// #include <linux/ip.h>
// ipv4 头, 字节大小: 1+1+2+2+2+1+1+2+4+4=20，也就是说，ipip 会多 20 字节大小，所以 MTU 原来 1500，使用 ipip 后只能搬运 1480 字节数据
struct iphdr {
    // ihl 值为: .... 0101 = Header Length: 20 bytes (5)
    __u8 ihl: 4; // Internet Header Length,
    __u8 version: 4;

    // Type of Service (TOS): 8 bits. This field is copied from the inner IP header
    __u8 tos;
    __u16 tot_len;

    // Identification: 16 bits
    // This field is used to identify the fragments of a datagram which will be helpful while reassembling the datagram as the encapsulator might fragment the datagram.
    // For the Outer IP Header, a new number is generated.
    __u16 id;

    __u16 frag_off; // fragment offset
    __u8 ttl;
    __u8 protocol;
    __u16 check;
    __u32 saddr;
    __u32 daddr;
} __attribute__((packed));


#endif
