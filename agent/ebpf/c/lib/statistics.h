#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include <linux/types.h>

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 2);
    __type(key, __u32);
    __type(value, __u64);
} pkg_count_metrics __section_maps_btf;

static inline void record_pass_count_metrics() {

    __u32 key = 0;
    __u64 *cnt = bpf_map_lookup_elem(&pkg_count_metrics, &key);
    if (cnt) __sync_fetch_and_add(cnt, 1);
}

static inline void record_drop_count_metrics() {

    __u32 key = 1;
    __u64 *cnt = bpf_map_lookup_elem(&pkg_count_metrics, &key);
    if (cnt) __sync_fetch_and_add(cnt, 1);
}
