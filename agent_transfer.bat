#ssh root@47.110.41.67 mkdir -p /root/xdp-banner/build/xdp-agent
#ssh root@47.110.32.185 mkdir -p /root/xdp-banner/build/xdp-agent
#ssh root@47.110.47.247 mkdir -p /root/xdp-banner/build/xdp-agent

ssh root@47.110.41.67 rm -rf /root/xdp-banner/build/xdp-agent
ssh root@47.110.32.185 rm -rf /root/xdp-banner/build/xdp-agent
ssh root@47.110.47.247 rm -rf /root/xdp-banner/build/xdp-agent

scp ./build/xdp-agent root@47.110.41.67:/root/xdp-banner/build/
scp ./build/xdp-agent root@47.110.32.185:/root/xdp-banner/build/
scp ./build/xdp-agent root@47.110.47.247:/root/xdp-banner/build/

scp ./systemd-services/xdp-agent.service root@47.110.41.67:/etc/systemd/system/xdp-agent.service
scp ./systemd-services/xdp-agent.service root@47.110.32.185:/etc/systemd/system/xdp-agent.service
scp ./systemd-services/xdp-agent.service root@47.110.47.247:/etc/systemd/system/xdp-agent.service
