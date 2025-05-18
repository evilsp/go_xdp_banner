import matplotlib.pyplot as plt
from matplotlib import rcParams

# 设置中文字体（若系统无对应字体请自行替换），并避免负号方块
rcParams['font.sans-serif'] = ['Microsoft YaHei', 'SimHei', 'Arial']
rcParams['axes.unicode_minus'] = False
# —— 1. PPS 对比 —— 
# ========== 1. PPS 对比（新增 iptables + 10 000 条规则） ==========
counts = [100_000, 200_000, 400_000, 800_000, 1_600_000, 3_200_000]

pps_baseline = [1706352.0, 840816.5, 567994.5, 538882.0, 521369.5, 512478.5]
pps_iptables100 = [1086032.5, 782896.5, 637699.0, 537471.5, 520010.0, 511509.5]
pps_iptables1000 = [1780090.5, 1053877.5, 630562.0, 553952.0, 521361.0, 512360.0]
# 由日志计算出的 10k 规则六轮次平均 PPS
pps_iptables10000 = [1933657.5, 1420998.5, 646101.0, 533170.0, 537521.5, 514691.0]

plt.figure(figsize=(8,5))
plt.plot(counts, pps_baseline,     'o-', label='Baseline (无 iptables)')
plt.plot(counts, pps_iptables100,  's-', label='iptables + 100 条规则')
plt.plot(counts, pps_iptables1000, 'd-', label='iptables + 1 000 条规则')
plt.plot(counts, pps_iptables10000,'^-', label='iptables + 10 000 条规则')

plt.xscale('log')
plt.xlabel('单队列发送包数 (-C)')
plt.ylabel('平均发送速率 (PPS)')
plt.title('iptables 启用前后 PPS 对比')
plt.grid(True, linestyle='--', alpha=0.5)
plt.legend()
plt.tight_layout()

# ========== 2. CPU 使用率对比（新增 iptables + 10 000 条规则） ==========
cpu_data = {
    'Baseline (无 iptables)': {
        'nic_pps': [95_496, 140_582, 317_622, 542_206, 714_219, 956_059],
        'cpu'    : [10.6,   11.6,   26.55,   46.5,    60.75,   81.1],
    },
    'iptables + 100 条规则': {
        'nic_pps': [95_496, 148_669, 304_521, 525_257, 806_158, 898_752],
        'cpu'    : [10.6,   16.9,   34.35,   59.0,    90.5,   100.0],
    },
    'iptables + 1 000 条规则': {
        'nic_pps': [95_624, 275_119, 522_824, 636_526, 828_427, 896_576],
        'cpu'    : [11.45,  31.5,   58.7,    70.85,   92.5,   100.0],
    },
    'iptables + 10 000 条规则': {
        # 选取 8 个代表采样点，按 NIC PPS 升序
        'nic_pps': [647, 1_155, 2_118, 3_458, 5_700, 8_715, 10_436, 11_520],
        'cpu'    : [5.9, 10.45, 18.3, 28.85, 47.25, 75.6,  89.65, 100.0],
    },
}

plt.figure(figsize=(8,5))
for label, dat in cpu_data.items():
    xs, ys = zip(*sorted(zip(dat['nic_pps'], dat['cpu'])))
    plt.plot(xs, ys, marker='o', label=label)

plt.xscale('log')
plt.xlabel('NIC 接收 PPS')
plt.ylabel('CPU 忙碌率 (%)')
plt.title('iptables CPU 使用率对比')
plt.grid(True, linestyle='--', alpha=0.5)
plt.legend(loc='upper left')
plt.tight_layout()

plt.show()