watch -n1 'R1=$(cat /sys/class/net/eth0/statistics/rx_packets); sleep 1; R2=$(cat /sys/class/net/eth0/statistics/rx_packets); echo RX PPS: $((R2-R1))'
