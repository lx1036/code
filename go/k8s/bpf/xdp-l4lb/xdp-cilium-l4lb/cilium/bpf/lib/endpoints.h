



#ifndef __LIB_EPS_H_
#define __LIB_EPS_H_






static __always_inline __maybe_unused struct endpoint_info * __lookup_ip4_endpoint(__u32 ip) {
	struct endpoint_key key = {};

	key.ip4 = ip;
	key.family = ENDPOINT_KEY_IPV4;

	return map_lookup_elem(&ENDPOINTS_MAP, &key);
}



#endif /* __LIB_EPS_H_ */