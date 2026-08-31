#!/bin/sh

set -eu

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <leader-http-addr> <node-id=node-addr> ..." >&2
    exit 2
fi

leader_addr=$1
shift

for member in "$@"; do
    node_id=${member%%=*}
    node_addr=${member#*=}

    echo "Joining $node_id at $node_addr..."
    until hyprctl --addr "$leader_addr" --timeout 2s join \
        --node-id "$node_id" \
        --node-addr "$node_addr"; do
        sleep 1
    done
done
