#!/usr/bin/env bash
XDP_OBJ="xdp_banner.o"
XDP_SEC="xdp"

# 遍历所有物理/虚拟接口（如果想跳过 lo，可在 grep -v lo）
for dev in $(ls /sys/class/net); do
  echo ">> Attaching $XDP_OBJ to $dev ..."
  sudo ip link set dev "$dev" xdp obj "$XDP_OBJ" sec "$XDP_SEC" \
    && echo "   ✔ $dev" \
    || echo "   ✖ $dev (attach failed)"
done
