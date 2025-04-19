#include <linux/types.h>
#include <linux/if_ether.h>
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "ctx.h"


/* __ETH_HLEN is out of sync with the kernel's if_ether.h. In Cilium datapath
 * we use ETH_HLEN which can be loaded via static data, and for L2-less devs
 * it's 0. To avoid replacing every occurrence of ETH_HLEN in the datapath,
 * we prefixed the kernel's ETH_HLEN instead.
 */

#define __ETH_HLEN	14		/* Total octets in header.	 */

// https://github.com/cilium/cilium/blob/d5c67e664a32bde51471178b62dacda16e990253/bpf/lib/eth.h#L8
#ifndef ETH_HLEN
#define ETH_HLEN __ETH_HLEN
#endif

static __always_inline bool eth_is_supported_ethertype(__be16 proto)
{
	/* non-Ethernet II unsupported */
	return proto >= bpf_htons(ETH_P_802_3_MIN);
}

static __always_inline bool validate_ethertype_l2_off(struct xdp_md *ctx,
    int l2_off, __u16 *proto)
{
    const __u64 tot_len = l2_off + ETH_HLEN;
    void *data_end = ctx_data_end(ctx);
    void *data = ctx_data(ctx);
    struct ethhdr *eth;

    if (ETH_HLEN == 0) {
        /* The packet is received on L2-less device. Determine L3
        * protocol from skb->protocol.
        */
        *proto = ctx_get_protocol(ctx);
        return true;
    }

    if (data + tot_len > data_end)
    return false;

    eth = data + l2_off;

    *proto = eth->h_proto;

    return eth_is_supported_ethertype(*proto);
}

static __always_inline bool validate_ethertype(struct xdp_md *ctx,
    __u16 *proto)
{
    return validate_ethertype_l2_off(ctx, 0, proto);
}

