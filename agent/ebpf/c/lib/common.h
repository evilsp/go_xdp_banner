#pragma once

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include <linux/types.h>
#include <asm-generic/errno.h>
#include <asm/byteorder.h>

#include "section.h"

#define IPCACHE_MAP_SIZE 512000
#define LIBBPF_PIN_BY_NAME 1
#define CIDR_LMAP_ELEMS 1024

// Ref from include/linux/socket.h
#define ENDPOINT_KEY_IPV4 2
#define ENDPOINT_KEY_IPV6 10

// https://github.com/cilium/cilium/blob/main/bpf/lib/ipv6.h#L122
#define GET_PREFIX(PREFIX)						\
	bpf_htonl(PREFIX <= 0 ? 0 : PREFIX < 32 ? ((1<<PREFIX) - 1) << (32-PREFIX)	\
			      : 0xFFFFFFFF)

// https://github.com/cilium/cilium/blob/main/bpf/include/bpf/compiler.h#L37
#ifndef offsetof
# define offsetof(T, M)		__builtin_offsetof(T, M)
#endif

#ifndef unlikely
# define unlikely(X)		__builtin_expect(!!(X), 0)
#endif

#ifndef __packed
#define __packed __attribute__((packed))
#endif

#ifndef __section_license
# define __section_license		__section("license")
#endif

#ifndef __section_maps
# define __section_maps			__section("maps")
#endif

#ifndef __section_maps_btf
# define __section_maps_btf		__section(".maps")
#endif

#ifndef BPF_LICENSE
# define BPF_LICENSE(NAME)				\
	char ____license[] __section_license = NAME
#endif

union v6addr {
	struct {
		__u32 p1;
		__u32 p2;
		__u32 p3;
		__u32 p4;
	};
	struct {
		__u64 d1;
		__u64 d2;
	};
	__u8 addr[16];
} __packed;

struct udphdr {
	__be16	source;
	__be16	dest;
	__be16	len;
	__sum16	check;
};

struct icmphdr {
  __u8		type;
  __u8		code;
  __sum16	checksum;
  union {
	struct {
		__be16	id;
		__be16	sequence;
	} echo;
	__be32	gateway;
	struct {
		__be16	__unused;
		__be16	mtu;
	} frag;
	__u8	reserved[4];
  } un;
};

struct icmp6hdr {

	__u8		icmp6_type;
	__u8		icmp6_code;
	__sum16		icmp6_cksum;


	union {
		__be32			un_data32[1];
		__be16			un_data16[2];
		__u8			un_data8[4];

		struct icmpv6_echo {
			__be16		identifier;
			__be16		sequence;
		} u_echo;

                struct icmpv6_nd_advt {
#if defined(__LITTLE_ENDIAN_BITFIELD)
                        __u32		reserved:5,
                        		override:1,
                        		solicited:1,
                        		router:1,
					reserved2:24;
#elif defined(__BIG_ENDIAN_BITFIELD)
                        __u32		router:1,
					solicited:1,
                        		override:1,
                        		reserved:29;
#else
#error	"Please fix <asm/byteorder.h>"
#endif						
                } u_nd_advt;

                struct icmpv6_nd_ra {
			__u8		hop_limit;
#if defined(__LITTLE_ENDIAN_BITFIELD)
			__u8		reserved:3,
					router_pref:2,
					home_agent:1,
					other:1,
					managed:1;

#elif defined(__BIG_ENDIAN_BITFIELD)
			__u8		managed:1,
					other:1,
					home_agent:1,
					router_pref:2,
					reserved:3;
#else
#error	"Please fix <asm/byteorder.h>"
#endif
			__be16		rt_lifetime;
                } u_nd_ra;

	} icmp6_dataun;

#define icmp6_identifier	icmp6_dataun.u_echo.identifier
#define icmp6_sequence		icmp6_dataun.u_echo.sequence
#define icmp6_pointer		icmp6_dataun.un_data32[0]
#define icmp6_mtu		icmp6_dataun.un_data32[0]
#define icmp6_unused		icmp6_dataun.un_data32[0]
#define icmp6_maxdelay		icmp6_dataun.un_data16[0]
#define icmp6_datagram_len	icmp6_dataun.un_data8[0]
#define icmp6_router		icmp6_dataun.u_nd_advt.router
#define icmp6_solicited		icmp6_dataun.u_nd_advt.solicited
#define icmp6_override		icmp6_dataun.u_nd_advt.override
#define icmp6_ndiscreserved	icmp6_dataun.u_nd_advt.reserved
#define icmp6_hop_limit		icmp6_dataun.u_nd_ra.hop_limit
#define icmp6_addrconf_managed	icmp6_dataun.u_nd_ra.managed
#define icmp6_addrconf_other	icmp6_dataun.u_nd_ra.other
#define icmp6_rt_lifetime	icmp6_dataun.u_nd_ra.rt_lifetime
#define icmp6_router_pref	icmp6_dataun.u_nd_ra.router_pref
};

// Make Variables Aligned
// Make the struct can be used to compare easily
// Proto defination: https://github.com/torvalds/linux/blob/1e26c5e28ca5821a824e90dd359556f5e9e7b89f/include/uapi/linux/in.h#L11
// iphdr and ip6hdr comes to void in order to avoid corruption
struct banrule_key {
	struct bpf_lpm_trie_key lpm;
	__u16 pad1;
	__u8 pad2;
	__u8 protocol;
	__u32 identity; 
	__u16 sport;
	__u16 dport;
} __packed;

struct banrule_val {
    __u64 latest_access_timestamp;
    __u64 refuse_times;
};

struct {
	__uint(type, BPF_MAP_TYPE_LPM_TRIE);
	__type(key, struct banrule_key);
	__type(value, struct banrule_val);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, CIDR_LMAP_ELEMS);
	__uint(map_flags, BPF_F_NO_PREALLOC);
} xdp_banner_banlist __section_maps_btf;

#define PREFIX_FULL      96  /* protocol+identity+sport+dport */
#define PREFIX_SPORT     80  /* protocol+identity+sport */
#define PREFIX_DPORT     96  /* sport=0, dport=X */
#define PREFIX_NONE      64  /* protocol+identity */

// Search rules
//static inline __maybe_unused struct banrule_val *
//lpm_key_lookup(const void *map, __u8 protocol, __u32 identity, __u16 sport, __u16 dport)
//{
//	struct banrule_key key = {
//		.lpm = { BANLIST_PREFIX, {} },
//		.protocol = protocol,
//		.identity = identity,
//		.sport = sport,
//		.dport = dport
//	};
//
//	return bpf_map_lookup_elem(map, &key);
//}

//static inline __maybe_unused int // 0 pass; 1 drop;
//lpm_rule_check(const void *map, __u8 protocol, __u32 identity, __u16 sport, __u16 dport)
//{
//        int action = 1;
//        struct banrule_val * rule_metric = lpm_key_lookup(map,protocol,identity,sport,dport);
//        if (!rule_metric){
//            static const char fmt[] = "identity_ipcache map init for identity %u not exist\n";
//            bpf_trace_printk(fmt, sizeof(fmt), identity);
//            action = 0;
//        }
//        return action;
//}

static inline __maybe_unused int
lpm_rule_check(const void *map, __u8 protocol, __u32 identity, __u16 sport, __u16 dport)
{

    struct banrule_key key = {
        /* zero-init */
    };
    key.protocol = protocol;
    key.identity = identity;

    /* 1) both sport & dport */
    key.sport = sport;
    key.dport = dport;
    key.lpm.prefixlen = PREFIX_FULL;

    if (bpf_map_lookup_elem(map, &key))
        return 1;

    /* 2) only dport (忽略 sport) */
    if (dport) {
        key.sport = 0;
        key.dport = dport;
        /* 由于 dport 在结构体尾部，只能用 full-length 做查找 */
        key.lpm.prefixlen = PREFIX_DPORT;
        if (bpf_map_lookup_elem(map, &key))
            return 1;
    }

	
    /* 3) only sport (忽略 dport) */
    if (sport) {
        key.dport = 0;
        key.lpm.prefixlen = PREFIX_SPORT;
        if (bpf_map_lookup_elem(map, &key))
            return 1;
    }

    /* 4) neither (只按 protocol+identity) */
    key.sport = 0;
    key.dport = 0;
    key.lpm.prefixlen = PREFIX_NONE;
    if (bpf_map_lookup_elem(map, &key))
        return 1;

	static const char fmt[] = "identity_ipcache map init for identity %u not exist\n";
    bpf_trace_printk(fmt, sizeof(fmt), identity);

    return 0;  /* no match ⇒ pass */
}

static inline void ipv6_addr_clear_suffix(union v6addr *addr,
	int prefix)
{
	addr->p1 &= GET_PREFIX(prefix);
	prefix -= 32;
	addr->p2 &= GET_PREFIX(prefix);
	prefix -= 32;
	addr->p3 &= GET_PREFIX(prefix);
	prefix -= 32;
	addr->p4 &= GET_PREFIX(prefix);
}

// Move from libbpf-tools/maps.bpf.h#L10
static __always_inline void *
bpf_map_lookup_or_try_init(void *map, const void *key, const void *init)
{
	void *val;
	/* bpf helper functions like bpf_map_update_elem() below normally return
	 * long, but using int instead of long to store the result is a workaround
	 * to avoid incorrectly evaluating err in cases where the following criteria
	 * is met:
	 *     the architecture is 64-bit
	 *     the helper function return type is long
	 *     the helper function returns the value of a call to a bpf_map_ops func
	 *     the bpf_map_ops function return type is int
	 *     the compiler inlines the helper function
	 *     the compiler does not sign extend the result of the bpf_map_ops func
	 *
	 * if this criteria is met, at best an error can only be checked as zero or
	 * non-zero. it will not be possible to check for a negative value or a
	 * specific error value. this is because the sign bit would have been stuck
	 * at the 32nd bit of a 64-bit long int.
	 */
	int err;

	val = bpf_map_lookup_elem(map, key);
	if (val)
		return val;

	err = bpf_map_update_elem(map, key, init, BPF_NOEXIST);
	if (err && err != -EEXIST)
		return 0;

	return bpf_map_lookup_elem(map, key);
}

//static __always_inline int update_ban_and_drop(const struct banrule_key *key, struct banrule_val *old) {
//	struct banrule_val v = *old;
//	v.latest_access_timestamp = bpf_ktime_get_ns();
//	v.refuse_times += 1;
//	bpf_map_update_elem(&xdp_banner_banlist, key, &v, BPF_ANY);
//	return XDP_DROP;
//}
