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
# 预定义 10 种不同协议
protos=(TCP UDP ICMP)

counter=0
group_size=3

while [ $counter -lt $TOTAL ]; do
  # 1. 随机生成一个 /32 CIDR
  cidr="$((RANDOM%256)).$((RANDOM%256)).$((RANDOM%256)).$((RANDOM%256))/32"

  # 2. 为本组生成不重复的 sport 列表
  declare -A seen_sport=()
  sports=()
  while [ ${#sports[@]} -lt $group_size ]; do
    p=$((RANDOM % 65535 + 1))
    if [ -z "${seen_sport[$p]:-}" ]; then
      seen_sport[$p]=1
      sports+=($p)
    fi
  done

  # 3. 为本组生成不重复的 dport 列表
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
  mapfile -t shuffled_protos < <(printf '%s\n' "${protos[@]}" | shuf)

  # 5. 生成本组内规则
  for (( i=0; i<group_size && counter<TOTAL; i++ )); do
    protocol=${shuffled_protos[i]}
    sport=${sports[i]}
    dport=${dports[i]}

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

    curl -s -X POST "$BASE_URL/v1/rules" \
         -H "Content-Type: application/json" \
         -d "$payload" \
      && echo "Posted [$((counter+1))]: $cidr $protocol $sport→$dport"

    counter=$((counter+1))
  done
done

echo "✔ 全部完成，共发送 $counter 条规则。"