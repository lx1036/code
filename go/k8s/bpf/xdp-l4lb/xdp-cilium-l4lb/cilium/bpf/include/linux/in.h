

// /root/linux-5.10.142/include/uapi/linux/in.h

#ifndef _UAPI_LINUX_IN_H
#define _UAPI_LINUX_IN_H


//#include <linux/socket.h>

/* Standard well-defined IP protocols.  */
enum {
    IPPROTO_IP = 0,        /* Dummy protocol for TCP		*/
#define IPPROTO_IP        IPPROTO_IP
    IPPROTO_ICMP = 1,        /* Internet Control Message Protocol	*/
#define IPPROTO_ICMP        IPPROTO_ICMP
    IPPROTO_IGMP = 2,        /* Internet Group Management Protocol	*/
#define IPPROTO_IGMP        IPPROTO_IGMP
    IPPROTO_IPIP = 4,        /* IPIP tunnels (older KA9Q tunnels use 94) */
#define IPPROTO_IPIP        IPPROTO_IPIP
    IPPROTO_TCP = 6,        /* Transmission Control Protocol	*/
#define IPPROTO_TCP        IPPROTO_TCP
    IPPROTO_EGP = 8,        /* Exterior Gateway Protocol		*/
#define IPPROTO_EGP        IPPROTO_EGP
    IPPROTO_PUP = 12,        /* PUP protocol				*/
#define IPPROTO_PUP        IPPROTO_PUP
    IPPROTO_UDP = 17,        /* User Datagram Protocol		*/
#define IPPROTO_UDP        IPPROTO_UDP
    IPPROTO_IDP = 22,        /* XNS IDP protocol			*/
#define IPPROTO_IDP        IPPROTO_IDP
    IPPROTO_TP = 29,        /* SO Transport Protocol Class 4	*/
#define IPPROTO_TP        IPPROTO_TP
    IPPROTO_DCCP = 33,        /* Datagram Congestion Control Protocol */
#define IPPROTO_DCCP        IPPROTO_DCCP
    IPPROTO_IPV6 = 41,        /* IPv6-in-IPv4 tunnelling		*/
#define IPPROTO_IPV6        IPPROTO_IPV6
    IPPROTO_RSVP = 46,        /* RSVP Protocol			*/
#define IPPROTO_RSVP        IPPROTO_RSVP
    IPPROTO_GRE = 47,        /* Cisco GRE tunnels (rfc 1701,1702)	*/
#define IPPROTO_GRE        IPPROTO_GRE
    IPPROTO_ESP = 50,        /* Encapsulation Security Payload protocol */
#define IPPROTO_ESP        IPPROTO_ESP
    IPPROTO_AH = 51,        /* Authentication Header protocol	*/
#define IPPROTO_AH        IPPROTO_AH
    IPPROTO_MTP = 92,        /* Multicast Transport Protocol		*/
#define IPPROTO_MTP        IPPROTO_MTP
    IPPROTO_BEETPH = 94,        /* IP option pseudo header for BEET	*/
#define IPPROTO_BEETPH        IPPROTO_BEETPH
    IPPROTO_ENCAP = 98,        /* Encapsulation Header			*/
#define IPPROTO_ENCAP        IPPROTO_ENCAP
    IPPROTO_PIM = 103,        /* Protocol Independent Multicast	*/
#define IPPROTO_PIM        IPPROTO_PIM
    IPPROTO_COMP = 108,        /* Compression Header Protocol		*/
#define IPPROTO_COMP        IPPROTO_COMP
    IPPROTO_SCTP = 132,        /* Stream Control Transport Protocol	*/
#define IPPROTO_SCTP        IPPROTO_SCTP
    IPPROTO_UDPLITE = 136,    /* UDP-Lite (RFC 3828)			*/
#define IPPROTO_UDPLITE        IPPROTO_UDPLITE
    IPPROTO_MPLS = 137,        /* MPLS in IP (RFC 4023)		*/
#define IPPROTO_MPLS        IPPROTO_MPLS
    IPPROTO_ETHERNET = 143,    /* Ethernet-within-IPv6 Encapsulation	*/
#define IPPROTO_ETHERNET    IPPROTO_ETHERNET
    IPPROTO_RAW = 255,        /* Raw IP packets			*/
#define IPPROTO_RAW        IPPROTO_RAW
    IPPROTO_MPTCP = 262,        /* Multipath TCP connection		*/
#define IPPROTO_MPTCP        IPPROTO_MPTCP
    IPPROTO_MAX
};

#define IP_TOS        1
#define IP_TTL        2
#define IP_HDRINCL    3
#define IP_OPTIONS    4
#define IP_ROUTER_ALERT    5
#define IP_RECVOPTS    6
#define IP_RETOPTS    7
#define IP_PKTINFO    8
#define IP_PKTOPTIONS    9
#define IP_MTU_DISCOVER    10
#define IP_RECVERR    11
#define IP_RECVTTL    12
#define    IP_RECVTOS    13
#define IP_MTU        14
#define IP_FREEBIND    15
#define IP_IPSEC_POLICY    16
#define IP_XFRM_POLICY    17
#define IP_PASSSEC    18
#define IP_TRANSPARENT    19

/* BSD compatibility */
#define IP_RECVRETOPTS    IP_RETOPTS

/* TProxy original addresses */
#define IP_ORIGDSTADDR       20
#define IP_RECVORIGDSTADDR   IP_ORIGDSTADDR

#define IP_MINTTL       21
#define IP_NODEFRAG     22
#define IP_CHECKSUM    23
#define IP_BIND_ADDRESS_NO_PORT    24
#define IP_RECVFRAGSIZE    25
#define IP_RECVERR_RFC4884    26

/* IP_MTU_DISCOVER values */
#define IP_PMTUDISC_DONT        0    /* Never send DF frames */
#define IP_PMTUDISC_WANT        1    /* Use per route hints	*/
#define IP_PMTUDISC_DO            2    /* Always DF		*/
#define IP_PMTUDISC_PROBE        3       /* Ignore dst pmtu      */
/* Always use interface mtu (ignores dst pmtu) but don't set DF flag.
 * Also incoming ICMP frag_needed notifications will be ignored on
 * this socket to prevent accepting spoofed ones.
 */
#define IP_PMTUDISC_INTERFACE        4
/* weaker version of IP_PMTUDISC_INTERFACE, which allows packets to get
 * fragmented if they exeed the interface mtu
 */
#define IP_PMTUDISC_OMIT        5

#define IP_MULTICAST_IF            32
#define IP_MULTICAST_TTL        33
#define IP_MULTICAST_LOOP        34
#define IP_ADD_MEMBERSHIP        35
#define IP_DROP_MEMBERSHIP        36
#define IP_UNBLOCK_SOURCE        37
#define IP_BLOCK_SOURCE            38
#define IP_ADD_SOURCE_MEMBERSHIP    39
#define IP_DROP_SOURCE_MEMBERSHIP    40
#define IP_MSFILTER            41
#define MCAST_JOIN_GROUP        42
#define MCAST_BLOCK_SOURCE        43
#define MCAST_UNBLOCK_SOURCE        44
#define MCAST_LEAVE_GROUP        45
#define MCAST_JOIN_SOURCE_GROUP        46
#define MCAST_LEAVE_SOURCE_GROUP    47
#define MCAST_MSFILTER            48
#define IP_MULTICAST_ALL        49
#define IP_UNICAST_IF            50

#define MCAST_EXCLUDE    0
#define MCAST_INCLUDE    1


// 来自于 #include <linux/socket.h>
typedef unsigned short __kernel_sa_family_t;

// 来自于 #include <linux/in.h>
/* Internet address. */
struct in_addr {
    __be32 s_addr;
};

#define __SOCK_SIZE__    16        /* sizeof(struct sockaddr)	*/
struct sockaddr_in {
    __kernel_sa_family_t sin_family;    /* Address family		*/
    __be16 sin_port;    /* Port number			*/
    struct in_addr sin_addr;    /* Internet address		*/

    /* Pad to size of `struct sockaddr'. */
    unsigned char __pad[__SOCK_SIZE__ - sizeof(short int) -
                        sizeof(unsigned short int) - sizeof(struct in_addr)];
};
#define sin_zero    __pad        /* for BSD UNIX comp. -FvK	*/

#endif
