#!/usr/bin/env bash
# nic_pps_cpu.sh —— 每秒输出 PASS/DROP PPS 与整机 CPU 忙碌率

MAP_ID=106

# 从 BPF map 中汇总 key 对应的所有 value
get_count_for_key() {
  local key=$1
  bpftool map dump id $MAP_ID | \
    sed -n "/\"key\": $key/,/\"key\":/ { /\"value\"/p }" | \
    sed -E 's/.*"value":[ ]*([0-9]+).*/\1/' | \
    { sum=0
      while read -r num; do
        sum=$((sum + num))
      done
      echo $sum
    }
}

while true; do
  # 1) 采集上次计数
  prev_pass=$(get_count_for_key 0)
  prev_drop=$(get_count_for_key 1)

  cpu_idle=$(mpstat 1 1 | awk '/all/ {print $NF}')
  # 100 - idle = 忙碌率
  cpu_busy=$(awk -v idle="$cpu_idle" 'BEGIN{printf "%.1f", 100 - idle}')
  curr_pass=$(get_count_for_key 0)
  curr_drop=$(get_count_for_key 1)

  pps_pass=$((curr_pass - prev_pass))
  pps_drop=$((curr_drop - prev_drop))

  timestamp=$(date +'%H:%M:%S')
  printf "%s  PASS PPS: %-8s  DROP PPS: %-8s  CPU 忙碌率: %5s%%\n" \
         "$timestamp" "$pps_pass" "$pps_drop" "$cpu_busy"
done