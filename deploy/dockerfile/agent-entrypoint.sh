#! /bin/sh

AGENT_CERT_DIR=/etc/xdp-banner

# Check if the agent has joined the cluster
if [ -z "$(ls -A $AGENT_CERT_DIR 2>/dev/null)" ]; then
    echo "Agent has not joined the cluster yet. join the agent to cluster..."
    HOST=$(hostname)

    echo "Register this agent first..."
    tokenJson=$(curl -X POST -d "{\"name\": \"$HOST\"}" "http://$xdp-banner_ORCH_ENDPOINT:6062/v1/agent/control/register")
    token=$(echo $tokenJson | jq .token )
    echo Register finished. Token: $token

    echo "Joining the agent to the cluster..."
    agent -e "$xdp-banner_ORCH_ENDPOINT:6061" join --token $token
    echo "Agent has joined the cluster. Starting the agent..."
fi

exec agent -e "$xdp-banner_ORCH_ENDPOINT:6061" server