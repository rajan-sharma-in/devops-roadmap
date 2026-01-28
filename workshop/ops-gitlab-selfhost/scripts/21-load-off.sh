#!/usr/bin/env bash
set -euo pipefail
CTX=${CTX:-kind-ops}
NS=ops-demo

kubectl --context "$CTX" -n "$NS" scale deployment/ops-demo-loadgen --replicas=0 || true
echo "Load generator OFF (scaled to 0)." 
