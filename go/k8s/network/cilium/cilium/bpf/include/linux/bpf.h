//
// Created by 刘祥 on 7/1/22.
//

#ifndef __LINUX_BPF_H__
#define __LINUX_BPF_H__

#include "bpf/types_mapper.h"

/* flags for BPF_MAP_UPDATE_ELEM command */
enum {
    BPF_ANY		= 0, /* create new element or update existing */
    BPF_NOEXIST	= 1, /* create new element if it didn't exist */
    BPF_EXIST	= 2, /* update existing element */
    BPF_F_LOCK	= 4, /* spin_lock-ed map_lookup/map_update */
};

enum bpf_map_type {
    BPF_MAP_TYPE_UNSPEC,
    BPF_MAP_TYPE_HASH,
    BPF_MAP_TYPE_ARRAY,
    BPF_MAP_TYPE_PROG_ARRAY,
    BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    BPF_MAP_TYPE_PERCPU_HASH,
    BPF_MAP_TYPE_PERCPU_ARRAY,
    BPF_MAP_TYPE_STACK_TRACE,
    BPF_MAP_TYPE_CGROUP_ARRAY,
    BPF_MAP_TYPE_LRU_HASH,
    BPF_MAP_TYPE_LRU_PERCPU_HASH,
    BPF_MAP_TYPE_LPM_TRIE,
    BPF_MAP_TYPE_ARRAY_OF_MAPS,
    BPF_MAP_TYPE_HASH_OF_MAPS,
    BPF_MAP_TYPE_DEVMAP,
    BPF_MAP_TYPE_SOCKMAP,
    BPF_MAP_TYPE_CPUMAP,
    BPF_MAP_TYPE_XSKMAP,
    BPF_MAP_TYPE_SOCKHASH,
    BPF_MAP_TYPE_CGROUP_STORAGE,
    BPF_MAP_TYPE_REUSEPORT_SOCKARRAY,
    BPF_MAP_TYPE_PERCPU_CGROUP_STORAGE,
    BPF_MAP_TYPE_QUEUE,
    BPF_MAP_TYPE_STACK,
    BPF_MAP_TYPE_SK_STORAGE,
    BPF_MAP_TYPE_DEVMAP_HASH,
    BPF_MAP_TYPE_STRUCT_OPS,
    BPF_MAP_TYPE_RINGBUF,
};

/* List of known BPF sock_ops operators.
 * New entries can only be added at the end
 */
enum {
    BPF_SOCK_OPS_VOID,
    BPF_SOCK_OPS_TIMEOUT_INIT,	/* Should return SYN-RTO value to use or
					 * -1 if default value should be used
					 */
    BPF_SOCK_OPS_RWND_INIT,		/* Should return initial advertized
					 * window (in packets) or -1 if default
					 * value should be used
					 */
    BPF_SOCK_OPS_TCP_CONNECT_CB,	/* Calls BPF program right before an
					 * active connection is initialized
					 */
    BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB,	/* Calls BPF program when an
						 * active connection is
						 * established
						 */
    BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB,	/* Calls BPF program when a
						 * passive connection is
						 * established
						 */
    BPF_SOCK_OPS_NEEDS_ECN,		/* If connection's congestion control
					 * needs ECN
					 */
    BPF_SOCK_OPS_BASE_RTT,		/* Get base RTT. The correct value is
					 * based on the path and may be
					 * dependent on the congestion control
					 * algorithm. In general it indicates
					 * a congestion threshold. RTTs above
					 * this indicate congestion
					 */
    BPF_SOCK_OPS_RTO_CB,		/* Called when an RTO has triggered.
					 * Arg1: value of icsk_retransmits
					 * Arg2: value of icsk_rto
					 * Arg3: whether RTO has expired
					 */
    BPF_SOCK_OPS_RETRANS_CB,	/* Called when skb is retransmitted.
					 * Arg1: sequence number of 1st byte
					 * Arg2: # segments
					 * Arg3: return value of
					 *       tcp_transmit_skb (0 => success)
					 */
    BPF_SOCK_OPS_STATE_CB,		/* Called when TCP changes state.
					 * Arg1: old_state
					 * Arg2: new_state
					 */
    BPF_SOCK_OPS_TCP_LISTEN_CB,	/* Called on listen(2), right after
					 * socket transition to LISTEN state.
					 */
    BPF_SOCK_OPS_RTT_CB,		/* Called on every RTT.
					 */
};

enum sk_action {
    SK_DROP = 0,
    SK_PASS,
};

/* BPF_FUNC_clone_redirect and BPF_FUNC_redirect flags. */
enum {
    BPF_F_INGRESS			= (1ULL << 0),
};

#define __bpf_md_ptr(type, name)	\
union {					\
	type name;			\
	__u64 :64;			\
} __attribute__((aligned(8)))

/* User bpf_sock_ops struct to access socket values and specify request ops
 * and their replies.
 * Some of this fields are in network (bigendian) byte order and may need
 * to be converted before use (bpf_ntohl() defined in samples/bpf/bpf_endian.h).
 * New fields can only be added at the end of this structure
 */
struct bpf_sock_ops {
    __u32 op;
    union {
        __u32 args[4];		/* Optionally passed to bpf program */
        __u32 reply;		/* Returned by bpf program	    */
        __u32 replylong[4];	/* Optionally returned by bpf prog  */
    };
    __u32 family;
    __u32 remote_ip4;	/* Stored in network byte order */
    __u32 local_ip4;	/* Stored in network byte order */
    __u32 remote_ip6[4];	/* Stored in network byte order */
    __u32 local_ip6[4];	/* Stored in network byte order */
    __u32 remote_port;	/* Stored in network byte order */
    __u32 local_port;	/* stored in host byte order */
    __u32 is_fullsock;	/* Some TCP fields are only valid if
				 * there is a full socket. If not, the
				 * fields read as zero.
				 */
    __u32 snd_cwnd;
    __u32 srtt_us;		/* Averaged RTT << 3 in usecs */
    __u32 bpf_sock_ops_cb_flags; /* flags defined in uapi/linux/tcp.h */
    __u32 state;
    __u32 rtt_min;
    __u32 snd_ssthresh;
    __u32 rcv_nxt;
    __u32 snd_nxt;
    __u32 snd_una;
    __u32 mss_cache;
    __u32 ecn_flags;
    __u32 rate_delivered;
    __u32 rate_interval_us;
    __u32 packets_out;
    __u32 retrans_out;
    __u32 total_retrans;
    __u32 segs_in;
    __u32 data_segs_in;
    __u32 segs_out;
    __u32 data_segs_out;
    __u32 lost_out;
    __u32 sacked_out;
    __u32 sk_txhash;
    __u64 bytes_received;
    __u64 bytes_acked;
    __bpf_md_ptr(struct bpf_sock *, sk);
};

/* user accessible metadata for SK_MSG packet hook, new fields must
 * be added to the end of this structure
 */
struct sk_msg_md {
    __bpf_md_ptr(void *, data);
    __bpf_md_ptr(void *, data_end);

    __u32 family;
    __u32 remote_ip4;	/* Stored in network byte order */
    __u32 local_ip4;	/* Stored in network byte order */
    __u32 remote_ip6[4];	/* Stored in network byte order */
    __u32 local_ip6[4];	/* Stored in network byte order */
    __u32 remote_port;	/* Stored in network byte order */
    __u32 local_port;	/* stored in host byte order */
    __u32 size;		/* Total size of sk_msg */

    __bpf_md_ptr(struct bpf_sock *, sk); /* current socket */
};





#endif //__LINUX_BPF_H__
