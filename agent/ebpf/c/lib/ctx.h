#pragma once

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/ip.h>
#include <linux/ipv6.h>

#include "common.h"

// https://github.com/cilium/cilium/blob/main/bpf/include/bpf/ctx/xdp.h#L134
#define DEFINE_FUNC_CTX_POINTER(FIELD)						\
static __always_inline void *							\
ctx_ ## FIELD(const struct xdp_md *ctx)						\
{										\
	void *ptr;								\
										\
	/* LLVM may generate u32 assignments of ctx->{data,data_end,data_meta}.	\
	 * With this inline asm, LLVM loses track of the fact this field is on	\
	 * 32 bits.								\
	 */									\
	asm volatile("%0 = *(u32 *)(%1 + %2)"					\
		     : "=r"(ptr)						\
		     : "r"(ctx), "i"(offsetof(struct xdp_md, FIELD)));		\
	return ptr;								\
}

/* This defines ctx_data(). */
DEFINE_FUNC_CTX_POINTER(data)
/* This defines ctx_data_end(). */
DEFINE_FUNC_CTX_POINTER(data_end)
/* This defines ctx_data_meta(). */
DEFINE_FUNC_CTX_POINTER(data_meta)
#undef DEFINE_FUNC_CTX_POINTER

static __always_inline bool ctx_no_room(const void *needed, const void *limit)
{
	return unlikely(needed > limit);
}

static __always_inline __maybe_unused __u16
ctx_get_protocol(const struct xdp_md *ctx)
{
	void *data_end = ctx_data_end(ctx);
	struct ethhdr *eth = ctx_data(ctx);

	if (ctx_no_room(eth + 1, data_end))
		return 0;

	return eth->h_proto;
}

static inline bool ipv4_ctx_fault(struct iphdr *ip_hdr, void *data_end){
	// 1. 首先检查 IPv4 头部是否完整（至少 20 字节）
	if (ctx_no_room(ip_hdr + 1, data_end)) {
		return 1;
	}
	
	// 2. 检查是否是无 option 的 iphdr（因为 ebpf 的静态编译器，检查长度必须是一个静态常量）
	if (ip_hdr->ihl != 5) {
		return 1;
	}

	return 0;
}

