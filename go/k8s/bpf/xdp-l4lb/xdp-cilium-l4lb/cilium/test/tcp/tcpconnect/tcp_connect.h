

/**
 * struct sock 定义见 /root/linux-5.10.142/include/net/sock.h
 */


#ifndef XDP_CILIUM_L4LB_TCP_CONNECT_H
#define XDP_CILIUM_L4LB_TCP_CONNECT_H

typedef unsigned int u32;
typedef int s32;
typedef int8_t   __s8;
typedef uint8_t  __u8;
typedef int16_t  __s16;
typedef uint16_t __u16;
typedef int32_t  __s32;
typedef uint32_t __u32;
typedef int64_t  __s64;
typedef uint64_t __u64;

typedef __u32 __bitwise __portpair;
typedef __u64 __bitwise __addrpair;

/**
 *	struct sock_common - minimal network layer representation of sockets
 *	@skc_daddr: Foreign IPv4 addr
 *	@skc_rcv_saddr: Bound local IPv4 addr
 *	@skc_addrpair: 8-byte-aligned __u64 union of @skc_daddr & @skc_rcv_saddr
 *	@skc_hash: hash value used with various protocol lookup tables
 *	@skc_u16hashes: two u16 hash values used by UDP lookup tables
 *	@skc_dport: placeholder for inet_dport/tw_dport
 *	@skc_num: placeholder for inet_num/tw_num
 *	@skc_portpair: __u32 union of @skc_dport & @skc_num
 *	@skc_family: network address family
 *	@skc_state: Connection state
 *	@skc_reuse: %SO_REUSEADDR setting
 *	@skc_reuseport: %SO_REUSEPORT setting
 *	@skc_ipv6only: socket is IPV6 only
 *	@skc_net_refcnt: socket is using net ref counting
 *	@skc_bound_dev_if: bound device index if != 0
 *	@skc_bind_node: bind hash linkage for various protocol lookup tables
 *	@skc_portaddr_node: second hash linkage for UDP/UDP-Lite protocol
 *	@skc_prot: protocol handlers inside a network family
 *	@skc_net: reference to the network namespace of this socket
 *	@skc_v6_daddr: IPV6 destination address
 *	@skc_v6_rcv_saddr: IPV6 source address
 *	@skc_cookie: socket's cookie value
 *	@skc_node: main hash linkage for various protocol lookup tables
 *	@skc_nulls_node: main hash linkage for TCP/UDP/UDP-Lite protocol
 *	@skc_tx_queue_mapping: tx queue number for this connection
 *	@skc_rx_queue_mapping: rx queue number for this connection
 *	@skc_flags: place holder for sk_flags
 *		%SO_LINGER (l_onoff), %SO_BROADCAST, %SO_KEEPALIVE,
 *		%SO_OOBINLINE settings, %SO_TIMESTAMPING settings
 *	@skc_listener: connection request listener socket (aka rsk_listener)
 *		[union with @skc_flags]
 *	@skc_tw_dr: (aka tw_dr) ptr to &struct inet_timewait_death_row
 *		[union with @skc_flags]
 *	@skc_incoming_cpu: record/match cpu processing incoming packets
 *	@skc_rcv_wnd: (aka rsk_rcv_wnd) TCP receive window size (possibly scaled)
 *		[union with @skc_incoming_cpu]
 *	@skc_tw_rcv_nxt: (aka tw_rcv_nxt) TCP window next expected seq number
 *		[union with @skc_incoming_cpu]
 *	@skc_refcnt: reference count
 *
 *	This is the minimal network layer representation of sockets, the header
 *	for struct sock and struct inet_timewait_sock.
 */
struct sock_common {
    union {
        __addrpair	skc_addrpair;
        struct {
            __be32	skc_daddr;
            __be32	skc_rcv_saddr;
        };
    };
    union  {
        unsigned int	skc_hash;
        __u16		skc_u16hashes[2];
    };
    /* skc_dport && skc_num must be grouped as well */
    union {
        __portpair	skc_portpair;
        struct {
            __be16	skc_dport;
            __u16	skc_num;
        };
    };

    unsigned short		skc_family;
    volatile unsigned char	skc_state;
    unsigned char		skc_reuse:4;
    unsigned char		skc_reuseport:1;
    unsigned char		skc_ipv6only:1;
    unsigned char		skc_net_refcnt:1;
    int			skc_bound_dev_if;
//    union {
//        struct hlist_node	skc_bind_node;
//        struct hlist_node	skc_portaddr_node;
//    };
    struct proto		*skc_prot;
//    possible_net_t		skc_net;

//#if IS_ENABLED(CONFIG_IPV6)
//    struct in6_addr		skc_v6_daddr;
//	struct in6_addr		skc_v6_rcv_saddr;
//#endif

//    atomic64_t		skc_cookie;

    /* following fields are padding to force
     * offset(struct sock, sk_refcnt) == 128 on 64bit arches
     * assuming IPV6 is enabled. We use this padding differently
     * for different kind of 'sockets'
     */
    union {
        unsigned long	skc_flags;
        struct sock	*skc_listener; /* request_sock */
        struct inet_timewait_death_row *skc_tw_dr; /* inet_timewait_sock */
    };
    /*
     * fields between dontcopy_begin/dontcopy_end
     * are not copied in sock_copy()
     */
    /* private: */
    int			skc_dontcopy_begin[0];
    /* public: */
//    union {
//        struct hlist_node	skc_node;
//        struct hlist_nulls_node skc_nulls_node;
//    };
    unsigned short		skc_tx_queue_mapping;
#ifdef CONFIG_XPS
    unsigned short		skc_rx_queue_mapping;
#endif
    union {
        int		skc_incoming_cpu;
        u32		skc_rcv_wnd;
        u32		skc_tw_rcv_nxt; /* struct tcp_timewait_sock  */
    };

//    refcount_t		skc_refcnt;
    /* private: */
    int                     skc_dontcopy_end[0];
    union {
        u32		skc_rxhash;
        u32		skc_window_clamp;
        u32		skc_tw_snd_nxt; /* struct tcp_timewait_sock */
    };
    /* public: */
};

struct bpf_local_storage;

/**
  *	struct sock - network layer representation of sockets
  *	@__sk_common: shared layout with inet_timewait_sock
  *	@sk_shutdown: mask of %SEND_SHUTDOWN and/or %RCV_SHUTDOWN
  *	@sk_userlocks: %SO_SNDBUF and %SO_RCVBUF settings
  *	@sk_lock:	synchronizer
  *	@sk_kern_sock: True if sock is using kernel lock classes
  *	@sk_rcvbuf: size of receive buffer in bytes
  *	@sk_wq: sock wait queue and async head
  *	@sk_rx_dst: receive input route used by early demux
  *	@sk_dst_cache: destination cache
  *	@sk_dst_pending_confirm: need to confirm neighbour
  *	@sk_policy: flow policy
  *	@sk_rx_skb_cache: cache copy of recently accessed RX skb
  *	@sk_receive_queue: incoming packets
  *	@sk_wmem_alloc: transmit queue bytes committed
  *	@sk_tsq_flags: TCP Small Queues flags
  *	@sk_write_queue: Packet sending queue
  *	@sk_omem_alloc: "o" is "option" or "other"
  *	@sk_wmem_queued: persistent queue size
  *	@sk_forward_alloc: space allocated forward
  *	@sk_napi_id: id of the last napi context to receive data for sk
  *	@sk_ll_usec: usecs to busypoll when there is no data
  *	@sk_allocation: allocation mode
  *	@sk_pacing_rate: Pacing rate (if supported by transport/packet scheduler)
  *	@sk_pacing_status: Pacing status (requested, handled by sch_fq)
  *	@sk_max_pacing_rate: Maximum pacing rate (%SO_MAX_PACING_RATE)
  *	@sk_sndbuf: size of send buffer in bytes
  *	@__sk_flags_offset: empty field used to determine location of bitfield
  *	@sk_padding: unused element for alignment
  *	@sk_no_check_tx: %SO_NO_CHECK setting, set checksum in TX packets
  *	@sk_no_check_rx: allow zero checksum in RX packets
  *	@sk_route_caps: route capabilities (e.g. %NETIF_F_TSO)
  *	@sk_route_nocaps: forbidden route capabilities (e.g NETIF_F_GSO_MASK)
  *	@sk_route_forced_caps: static, forced route capabilities
  *		(set in tcp_init_sock())
  *	@sk_gso_type: GSO type (e.g. %SKB_GSO_TCPV4)
  *	@sk_gso_max_size: Maximum GSO segment size to build
  *	@sk_gso_max_segs: Maximum number of GSO segments
  *	@sk_pacing_shift: scaling factor for TCP Small Queues
  *	@sk_lingertime: %SO_LINGER l_linger setting
  *	@sk_backlog: always used with the per-socket spinlock held
  *	@sk_callback_lock: used with the callbacks in the end of this struct
  *	@sk_error_queue: rarely used
  *	@sk_prot_creator: sk_prot of original sock creator (see ipv6_setsockopt,
  *			  IPV6_ADDRFORM for instance)
  *	@sk_err: last error
  *	@sk_err_soft: errors that don't cause failure but are the cause of a
  *		      persistent failure not just 'timed out'
  *	@sk_drops: raw/udp drops counter
  *	@sk_ack_backlog: current listen backlog
  *	@sk_max_ack_backlog: listen backlog set in listen()
  *	@sk_uid: user id of owner
  *	@sk_priority: %SO_PRIORITY setting
  *	@sk_type: socket type (%SOCK_STREAM, etc)
  *	@sk_protocol: which protocol this socket belongs in this network family
  *	@sk_peer_pid: &struct pid for this socket's peer
  *	@sk_peer_cred: %SO_PEERCRED setting
  *	@sk_rcvlowat: %SO_RCVLOWAT setting
  *	@sk_rcvtimeo: %SO_RCVTIMEO setting
  *	@sk_sndtimeo: %SO_SNDTIMEO setting
  *	@sk_txhash: computed flow hash for use on transmit
  *	@sk_filter: socket filtering instructions
  *	@sk_timer: sock cleanup timer
  *	@sk_stamp: time stamp of last packet received
  *	@sk_stamp_seq: lock for accessing sk_stamp on 32 bit architectures only
  *	@sk_tsflags: SO_TIMESTAMPING socket options
  *	@sk_tskey: counter to disambiguate concurrent tstamp requests
  *	@sk_zckey: counter to order MSG_ZEROCOPY notifications
  *	@sk_socket: Identd and reporting IO signals
  *	@sk_user_data: RPC layer private data
  *	@sk_frag: cached page frag
  *	@sk_peek_off: current peek_offset value
  *	@sk_send_head: front of stuff to transmit
  *	@tcp_rtx_queue: TCP re-transmit queue [union with @sk_send_head]
  *	@sk_tx_skb_cache: cache copy of recently accessed TX skb
  *	@sk_security: used by security modules
  *	@sk_mark: generic packet mark
  *	@sk_cgrp_data: cgroup data for this cgroup
  *	@sk_memcg: this socket's memory cgroup association
  *	@sk_write_pending: a write to stream socket waits to start
  *	@sk_state_change: callback to indicate change in the state of the sock
  *	@sk_data_ready: callback to indicate there is data to be processed
  *	@sk_write_space: callback to indicate there is bf sending space available
  *	@sk_error_report: callback to indicate errors (e.g. %MSG_ERRQUEUE)
  *	@sk_backlog_rcv: callback to process the backlog
  *	@sk_validate_xmit_skb: ptr to an optional validate function
  *	@sk_destruct: called at sock freeing time, i.e. when all refcnt == 0
  *	@sk_reuseport_cb: reuseport group container
  *	@sk_bpf_storage: ptr to cache and control for bpf_sk_storage
  *	@sk_rcu: used during RCU grace period
  *	@sk_clockid: clockid used by time-based scheduling (SO_TXTIME)
  *	@sk_txtime_deadline_mode: set deadline mode for SO_TXTIME
  *	@sk_txtime_report_errors: set report errors mode for SO_TXTIME
  *	@sk_txtime_unused: unused txtime flags
  */
struct sock {
    /*
     * Now struct inet_timewait_sock also uses sock_common, so please just
     * don't add nothing before this first member (__sk_common) --acme
     */
    struct sock_common __sk_common;
#define sk_node            __sk_common.skc_node
#define sk_nulls_node        __sk_common.skc_nulls_node
#define sk_refcnt        __sk_common.skc_refcnt
#define sk_tx_queue_mapping    __sk_common.skc_tx_queue_mapping
#ifdef CONFIG_XPS
#define sk_rx_queue_mapping	__sk_common.skc_rx_queue_mapping
#endif

#define sk_dontcopy_begin    __sk_common.skc_dontcopy_begin
#define sk_dontcopy_end        __sk_common.skc_dontcopy_end
#define sk_hash            __sk_common.skc_hash
#define sk_portpair        __sk_common.skc_portpair
#define sk_num            __sk_common.skc_num
#define sk_dport        __sk_common.skc_dport
#define sk_addrpair        __sk_common.skc_addrpair
#define sk_daddr        __sk_common.skc_daddr
#define sk_rcv_saddr        __sk_common.skc_rcv_saddr
#define sk_family        __sk_common.skc_family
#define sk_state        __sk_common.skc_state
#define sk_reuse        __sk_common.skc_reuse
#define sk_reuseport        __sk_common.skc_reuseport
#define sk_ipv6only        __sk_common.skc_ipv6only
#define sk_net_refcnt        __sk_common.skc_net_refcnt
#define sk_bound_dev_if        __sk_common.skc_bound_dev_if
#define sk_bind_node        __sk_common.skc_bind_node
#define sk_prot            __sk_common.skc_prot
#define sk_net            __sk_common.skc_net
#define sk_v6_daddr        __sk_common.skc_v6_daddr
#define sk_v6_rcv_saddr    __sk_common.skc_v6_rcv_saddr
#define sk_cookie        __sk_common.skc_cookie
#define sk_incoming_cpu        __sk_common.skc_incoming_cpu
#define sk_flags        __sk_common.skc_flags
#define sk_rxhash        __sk_common.skc_rxhash

}











#endif //XDP_CILIUM_L4LB_TCP_CONNECT_H
