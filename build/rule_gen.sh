#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ./generate_rules.sh TOTAL_RULES [BASE_URL]
#   TOTAL_RULES: 要生成的规则总条数（正整数）
#   BASE_URL   : 可选，默认 http://az.evilsp4.ltd:6062

if [ $# -lt 1 ]; then
  echo "用法：$0 TOTAL_RULES [BASE_URL]"
  exit 1
fi

TOTAL=$1
BASE_URL=${2:-"http://az.evilsp4.ltd:6062"}
NAME="default"
DURATION="7200s"
RULE_FILE="rule.rule"
> "$RULE_FILE"   # 清空或新建输出文件

# 支持协议列表
protos=(TCP UDP ICMP)
counter=0
group_size=3

while [ $counter -lt $TOTAL ]; do
  # 1. 随机生成一个 /32 CIDR
  cidr="$((RANDOM%256)).$((RANDOM%256)).$((RANDOM%256)).$((RANDOM%256))/32"

  # 2. 生成不重复的 source ports 列表
  declare -A seen_sport=()
  sports=()
  while [ ${#sports[@]} -lt $group_size ]; do
    p=$((RANDOM % 65535 + 1))
    if [ -z "${seen_sport[$p]:-}" ]; then
      seen_sport[$p]=1
      sports+=($p)
    fi
  done

  # 3. 生成不重复的 destination ports 列表
  declare -A seen_dport=()
  dports=()
  while [ ${#dports[@]} -lt $group_size ]; do
    p=$((RANDOM % 65535 + 1))
    if [ -z "${seen_dport[$p]:-}" ]; then
      seen_dport[$p]=1
      dports+=($p)
    fi
  done

  # 4. 随机打乱协议顺序
  mapfile -t shuffled_protos < <(printf '%s
' "${protos[@]}" | shuf)

  # 5. 生成每条规则
  for (( i=0; i<group_size && counter<TOTAL; i++ )); do
    protocol=${shuffled_protos[i]}
    sport=${sports[i]}
    dport=${dports[i]}

    # 构造 JSON payload
    payload=$(cat <<EOF
{
  "name": "$NAME",
  "rule": {
    "cidr": "$cidr",
    "protocol": "$protocol",
    "sport": $sport,
    "dport": $dport,
    "comment": "auto-generated",
    "duration": "$DURATION"
  }
}
EOF
)

    # 发送给 xdp-agent
    echo "Posted [$((counter+1))]: $cidr $protocol $sport→$dport"

    # 同时生成等效 iptables 规则并追加到文件
    # 对于 TCP/UDP，使用 -m tcp/udp 模块进行端口匹配；ICMP 忽略端口
    case "$protocol" in
      TCP)
        ipt_rule="iptables -A INPUT -s ${cidr%/*} -m iprange --src-range ${cidr%/*}-${cidr%/*} -p tcp -m tcp"
        [ "$sport" -ne 0 ] && ipt_rule+=" --sport $sport"
        [ "$dport" -ne 0 ] && ipt_rule+=" --dport $dport"
        ;;
      UDP)
        ipt_rule="iptables -A INPUT -s ${cidr%/*} -m iprange --src-range ${cidr%/*}-${cidr%/*} -p udp -m udp"
        [ "$sport" -ne 0 ] && ipt_rule+=" --sport $sport"
        [ "$dport" -ne 0 ] && ipt_rule+=" --dport $dport"
        ;;
      ICMP)
        ipt_rule="iptables -A INPUT -s ${cidr%/*} -m iprange --src-range ${cidr%/*}-${cidr%/*} -p icmp"
        ;;
    esac
    ipt_rule+=" -j DROP"
    echo "$ipt_rule" >> "$RULE_FILE"

    counter=$((counter+1))
  done
done

echo "✔ 全部完成，共发送 $counter 条规则，并生成 iptables 规则到 $RULE_FILE。"