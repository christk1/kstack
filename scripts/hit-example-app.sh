#!/usr/bin/env bash
# Simple demo script: hit example app every second until stopped
set -euo pipefail
HOST=${1:-http://localhost:8080}
INTERVAL=${2:-1}
echo "Hitting $HOST every $INTERVAL second(s). Ctrl-C to stop."
while true; do
  curl -sS "$HOST/" > /dev/null || true
  sleep "$INTERVAL"
done
