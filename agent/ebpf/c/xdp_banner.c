#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/tcp.h>
#include <linux/in.h>
#include <linux/in6.h>
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#include "lib/common.h"
#include "lib/ctx.h"
#include "lib/eps.h"
#include "lib/eth.h"

int check_v4(struct xdp_md *ctx){
    void *data_end = ctx_data_end(ctx);
    void *data = ctx_data(ctx);
    struct iphdr *ipv4_hdr = data + sizeof(struct ethhdr);
    // ipv6_hdr + 1 == (void *)ipv6_hdr + sizeof(struct ipv6hdr), add an element
    if ((void*)(ipv4_hdr + 1) > data_end)    // 先保证能读整个最小 iphdr
        return XDP_DROP;
    __u8 ihl = ipv4_hdr->ihl;
    if (ihl != 5)                      // verifier 能识别的常量比较
        return XDP_DROP;

    __u8 hdr_protocol = ipv4_hdr->protocol;
    __u32 saddr = ipv4_hdr->saddr;
    void *l4 = ipv4_hdr + 1;

    struct identity_info *identity = ipcache_lookup4(&identity_ipcache,saddr,32);

    if (!identity) {
        goto pass;
    }

    static const char identity_message[] = "Get package from ip %x.Identity: %u\n";
    bpf_trace_printk(identity_message, sizeof(identity_message), saddr,identity->identity);

    int action = 0;

    switch (hdr_protocol){
    case IPPROTO_ICMP:
        action = lpm_rule_check(&xdp_banner_banlist, hdr_protocol, identity->identity, \
            0, 0);
        if (action == 0){
            goto pass;
        }
        struct icmphdr *icmp = (struct icmphdr *)(l4);
        if (ctx_no_room(icmp + 1, data_end)){
            return XDP_DROP;
        }
        goto drop;
    case IPPROTO_TCP:
        struct tcphdr *tcp = (struct tcphdr *)(l4);
        if (ctx_no_room(tcp + 1, data_end))
            return XDP_DROP;
        // Trans source pointer type to void* in order to avoid action undefined
        action = lpm_rule_check(&xdp_banner_banlist, hdr_protocol, identity->identity, \
            tcp -> source, tcp -> dest);
        if (action == 0) {
            return XDP_PASS;
        }
        goto drop;
    case IPPROTO_UDP:
        struct udphdr *udp = (struct udphdr *)(l4);
        if (ctx_no_room(udp + 1, data_end))
            return XDP_DROP;
        action = lpm_rule_check(&xdp_banner_banlist, hdr_protocol, identity->identity, \
            udp -> source, udp -> dest);
        if (action == 0) {
            return XDP_PASS;
        }
        goto drop;
    default:
        goto drop;
    }
drop:
    return XDP_DROP;
pass:
    return XDP_PASS;
}


int check_v6(struct xdp_md *ctx){
    void *data_end = ctx_data_end(ctx);
    void *data = ctx_data(ctx);
    struct ipv6hdr *ipv6_hdr = data + sizeof(struct ethhdr);

    if (ctx_no_room(ipv6_hdr + 1, data_end))
        return XDP_DROP;

    __u8 hdr_protocol = ipv6_hdr->nexthdr;
    __u64 ipv6_fore_data = ((__u64)ipv6_hdr->saddr.s6_addr32[0] << 32) | ipv6_hdr->saddr.s6_addr32[1];
    __u64 ipv6_after_data = ((__u64)ipv6_hdr->saddr.s6_addr32[2] << 32) | ipv6_hdr->saddr.s6_addr32[3];
    void *l4 = ipv6_hdr + 1;

    struct identity_info *identity = ipcache_lookup6(&identity_ipcache,(union v6addr *)&(ipv6_hdr->saddr),128);

    if (!identity) {
        goto pass;
    }

    static const char identity_message[] = "Get package from ip %llx %llx.Identity: %u\n";
    bpf_trace_printk(identity_message, sizeof(identity_message), ipv6_fore_data, ipv6_after_data,identity->identity);

    int action = 0;

    switch (hdr_protocol){
    case IPPROTO_ICMPV6:
        action = lpm_rule_check(&xdp_banner_banlist, hdr_protocol, identity->identity, \
            0, 0);
        if (action == 0){
            goto pass;
        }
        struct icmp6hdr *icmp6 = (struct icmp6hdr *)(l4);
        if (ctx_no_room(icmp6 + 1, data_end))
            return XDP_DROP;
        goto drop;
    case IPPROTO_TCP:
        struct tcphdr *tcp6 = (struct tcphdr *)(l4);
        if (ctx_no_room(tcp6 + 1, data_end))
            return XDP_DROP;
        action = lpm_rule_check(&xdp_banner_banlist,hdr_protocol,identity->identity, \
            tcp6->source,tcp6->dest);
        if (action == 0)
            return XDP_PASS;
        goto drop;
    case IPPROTO_UDP:
        struct udphdr *udp6 = (struct udphdr *)(l4);
        if (ctx_no_room(udp6 + 1, data_end))
            return XDP_DROP;
        action = lpm_rule_check(&xdp_banner_banlist, hdr_protocol, identity->identity, \
            udp6 -> source, udp6 -> dest);
        if (action == 0) {
            return XDP_PASS;
        }
        goto drop;
    default:
        goto drop;
    }
drop:
    return XDP_DROP;
pass:
    return XDP_PASS;
}


// If hook needed
#ifndef xdp_early_hook
#define xdp_early_hook(ctx, proto) XDP_PASS
#endif

static __always_inline int check_filters(struct xdp_md *ctx){

    int ret = XDP_PASS;
        __u16 proto;

        if (!validate_ethertype(ctx, &proto))
                return XDP_PASS;

        ret = xdp_early_hook(ctx, proto);
        if (ret != XDP_PASS)
                return ret;

        switch (proto) {
        case bpf_htons(ETH_P_IP):
                ret = check_v4(ctx);
                break;
        case bpf_htons(ETH_P_IPV6):
                ret = check_v6(ctx);
                break;
        default:
                break;
        }

        return ret;
}

SEC("xdp")
int cil_xdp_entry(struct xdp_md *ctx)
{
        return check_filters(ctx);
}

char ____license[] __section("license") = "GPL";
