#!/usr/bin/env bash
# nic_pps.sh —— 每秒统计 eth0 全部 RX PPS

IFACE="eth0"

get_total_rx() {
  ethtool -S "$IFACE" 2>/dev/null |
    awk '/rx_queue_[0-9]+_packets/ { sum += $2 }
         END { printf "%.0f\n", sum }'     # 始终输出纯整数
}

prev=$(get_total_rx)

while true; do
  sleep 1
  curr=$(get_total_rx)
  pps=$((curr - prev))            # 现在是合法整数运算
  printf '%(%H:%M:%S)T  NIC RX PPS: %s\n' -1 "$pps"
  prev=$curr
done