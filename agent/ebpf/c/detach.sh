for dev in $(ls /sys/class/net); do
  sudo ip link set dev "$dev" xdp off 2>/dev/null || true
done
