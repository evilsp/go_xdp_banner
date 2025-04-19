#pragma once

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#include "common.h"

struct ipcache_key {
	struct bpf_lpm_trie_key lpm_key;
    __u16 pad1;
	__u8 pad2;
	__u8 family;   /*AF_INET / AF_INET6*/ 
	union {
		struct {
			__u32		ip4;
			__u32		pad3;
			__u32		pad4;
			__u32		pad5;
		} ip4;
		union v6addr	ip6;
	};
} __packed;

struct identity_info {
	__u32		identity; 
};

struct {
	__uint(type, BPF_MAP_TYPE_LPM_TRIE);
	__type(key, struct ipcache_key);
	__type(value, struct identity_info);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, IPCACHE_MAP_SIZE);
	__uint(map_flags, BPF_F_NO_PREALLOC);
} identity_ipcache __section_maps_btf;

/* IPCACHE_STATIC_PREFIX gets sizeof non-IP, non-prefix part of ipcache_key */
#define IPCACHE_STATIC_PREFIX							\
	(8 * (sizeof(struct ipcache_key) - sizeof(struct bpf_lpm_trie_key)	\
	      - sizeof(union v6addr)))
#define IPCACHE_PREFIX_LEN(PREFIX) (IPCACHE_STATIC_PREFIX + (PREFIX))

static inline __maybe_unused struct identity_info *
ipcache_lookup6(const void *map, const union v6addr *addr,
		__u32 prefix)
{
	struct ipcache_key key = {
		.lpm_key = { IPCACHE_PREFIX_LEN(prefix), {} },
		.family = ENDPOINT_KEY_IPV6,
		.ip6 = *addr,
	};

	ipv6_addr_clear_suffix(&key.ip6, prefix);
	return bpf_map_lookup_elem(map, &key);
}

// __be32 Big-Endian 32-bit (net)
// __u32 Unsigned 32-bit（host）

static inline __maybe_unused struct identity_info *
ipcache_lookup4(const void *map, __be32 addr, __u32 prefix)
{
	struct ipcache_key key = {
		.lpm_key      = { IPCACHE_PREFIX_LEN(prefix), {} },
		.family       = ENDPOINT_KEY_IPV4,
		.ip4.ip4      = addr,                    // 把 __be32 addr 塞到 .ip4.ip4
		/* pad1,pad2,ip4.pad3..pad5 自动归零 */
	};
	key.ip4.ip4 &= GET_PREFIX(prefix);           // 只掩码真正的 IP 字段
	return bpf_map_lookup_elem(map, &key);
}
