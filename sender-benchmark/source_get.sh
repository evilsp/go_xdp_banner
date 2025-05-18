#!/usr/bin/env bash
# nic_pps_cpu.sh —— 每秒输出 NIC RX PPS + 每核 Busy%

IFACE="eth0"

# ---------- 函数 ----------
get_total_rx() {
  ethtool -S "$IFACE" 2>/dev/null |
    awk '/rx_queue_[0-9]+_packets/ {sum += $2}
         END {printf "%.0f\n", sum}'
}

parse_cpu_busy() {
  # 读取 mpstat 临时文件，输出 "cpu00=xx.x% cpu01=yy.y% ..."
  awk '$3 ~ /^[0-9]+$/ {
         busy = 100 - $13;                    # 100-Idle
         printf "cpu%02d=%4.1f%% ", $3, busy
       }' "$1"
}

# ---------- 主循环 ----------
while true; do
  # 1) 记录起始包计数
  prev_rx=$(get_total_rx)

  # 2) 启动 mpstat 并行采样 1 s，结果写临时文件
  TMP_MP=/tmp/mpstat_$$.log
  mpstat -P ALL 1 1 > "$TMP_MP" &   # 后台运行
  MP_PID=$!

  # 3) 同步等待 1 s 结束（保持与 mpstat 采样窗口对齐）
  sleep 1

  # 4) 计算 PPS
  curr_rx=$(get_total_rx)
  pps=$(( curr_rx - prev_rx ))

  # 5) 等待 mpstat 完全结束、解析每核 Busy%
  wait "$MP_PID" 2>/dev/null
  cpu_busy=$(parse_cpu_busy "$TMP_MP")
  rm -f "$TMP_MP"

  # 6) 打印结果
  printf '%(%H:%M:%S)T  NIC RX PPS: %-8s | %s\n' -1 "$pps" "$cpu_busy"
done