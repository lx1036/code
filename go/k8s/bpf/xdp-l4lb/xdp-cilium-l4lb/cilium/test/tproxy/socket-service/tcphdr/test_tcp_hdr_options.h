



#ifndef SOL_TCP
#define SOL_TCP 6
#endif

#define TCPHDR_FIN 0x01
#define TCPHDR_SYN 0x02
#define TCPHDR_RST 0x04
#define TCPHDR_PSH 0x08
#define TCPHDR_ACK 0x10
#define TCPHDR_URG 0x20
#define TCPHDR_ECE 0x40
#define TCPHDR_CWR 0x80
#define TCPHDR_SYNACK (TCPHDR_SYN | TCPHDR_ACK)

#define TCPOPT_EOL		0
#define TCPOPT_NOP		1
#define TCPOPT_WINDOW		3
#define TCPOPT_EXP		254

#define CG_OK	1
#define CG_ERR	0

#define TCP_BPF_EXPOPT_BASE_LEN 4

struct bpf_test_option {
	__u8 flags;
	__u8 max_delack_ms;
	__u8 rand;
} __attribute__((packed));

struct linum_err {
	unsigned int linum;
	int err;
};

// key->value
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__type(key, int);
	__type(value, struct linum_err);
	__uint(max_entries, 2);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} lport_linum_map SEC(".maps");

/* Store in bpf_sk_storage */
struct hdr_stg {
	bool active;
	bool resend_syn; /* active side only */
	bool syncookie;  /* passive side only */
	bool fastopen;	/* passive side only */
};
struct {
	__uint(type, BPF_MAP_TYPE_SK_STORAGE);
	__uint(map_flags, BPF_F_NO_PREALLOC);
	__type(key, int);
	__type(value, struct hdr_stg);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} hdr_stg_map SEC(".maps");

static inline void clear_hdr_cb_flags(struct bpf_sock_ops *skops)
{
	bpf_sock_ops_cb_flags_set(skops,
				  skops->bpf_sock_ops_cb_flags &
				  ~(BPF_SOCK_OPS_PARSE_UNKNOWN_HDR_OPT_CB_FLAG |
				    BPF_SOCK_OPS_WRITE_HDR_OPT_CB_FLAG));
}

static inline void clear_parse_all_hdr_cb_flags(struct bpf_sock_ops *skops)
{
	bpf_sock_ops_cb_flags_set(skops,
				  skops->bpf_sock_ops_cb_flags &
				  ~BPF_SOCK_OPS_PARSE_ALL_HDR_OPT_CB_FLAG);
}

static inline void set_parse_all_hdr_cb_flags(struct bpf_sock_ops *skops)
{
	bpf_sock_ops_cb_flags_set(skops,
				  skops->bpf_sock_ops_cb_flags |
				  BPF_SOCK_OPS_PARSE_ALL_HDR_OPT_CB_FLAG);
}


#define RET_CG_ERR(__err) ({			\
	struct linum_err __linum_err;		\
	int __lport;				\
						\
	__linum_err.linum = __LINE__;		\
	__linum_err.err = __err;		\
	__lport = skops->local_port;		\
	bpf_map_update_elem(&lport_linum_map, &__lport, &__linum_err, BPF_NOEXIST); \
	clear_hdr_cb_flags(skops);					\
	clear_parse_all_hdr_cb_flags(skops);				\
	return CG_ERR;							\
})


struct tcp_exprm_opt {
	__u8 kind;
	__u8 len;
	__u16 magic;
	union {
		__u8 data[4];
		__u32 data32;
	};
} __attribute__((packed));

struct tcp_opt {
	__u8 kind;
	__u8 len;
	union {
		__u8 data[4];
		__u32 data32;
	};
} __attribute__((packed));


__u8 test_kind = TCPOPT_EXP;
__u16 test_magic = 0xeB9F;

#define TEST_OPTION_FLAGS(flags, option) (1 & ((flags) >> (option)))
#define SET_OPTION_FLAGS(flags, option)	((flags) |= (1 << (option)))

enum {
	OPTION_RESEND,
	OPTION_MAX_DELACK_MS,
	OPTION_RAND,
	__NR_OPTION_FLAGS,
};

static int parse_test_option(struct bpf_test_option *opt, const __u8 *start)
{
	opt->flags = *start++;

	if (TEST_OPTION_FLAGS(opt->flags, OPTION_MAX_DELACK_MS))
		opt->max_delack_ms = *start++;

	if (TEST_OPTION_FLAGS(opt->flags, OPTION_RAND))
		opt->rand = *start++;

	return 0;
}

static int load_option(struct bpf_sock_ops *skops,
		       struct bpf_test_option *test_opt, bool from_syn)
{
	union {
		struct tcp_exprm_opt exprm;
		struct tcp_opt regular;
	} search_opt;
	int ret, load_flags = from_syn ? BPF_LOAD_HDR_OPT_TCP_SYN : 0;

	if (test_kind == TCPOPT_EXP) {
		search_opt.exprm.kind = TCPOPT_EXP;
		search_opt.exprm.len = 4;
		search_opt.exprm.magic = __bpf_htons(test_magic);
		search_opt.exprm.data32 = 0;
		ret = bpf_load_hdr_opt(skops, &search_opt.exprm, sizeof(search_opt.exprm), load_flags);
		if (ret < 0)
			return ret;
		return parse_test_option(test_opt, search_opt.exprm.data);
	} else {
		search_opt.regular.kind = test_kind;
		search_opt.regular.len = 0;
		search_opt.regular.data32 = 0;
		ret = bpf_load_hdr_opt(skops, &search_opt.regular, sizeof(search_opt.regular), load_flags);
		if (ret < 0)
			return ret;
		return parse_test_option(test_opt, search_opt.regular.data);
	}
}

static inline __u8 skops_tcp_flags(const struct bpf_sock_ops *skops)
{
	return skops->skb_tcp_flags;
}

static inline unsigned int tcp_hdrlen(const struct tcphdr *th)
{
	return th->doff << 2; // ???
}