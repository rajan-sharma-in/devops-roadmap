#!/usr/bin/env bash
set -euo pipefail
CTX=${CTX:-kind-workshop-cluster}
NS=ops-demo

kubectl --context "$CTX" -n "$NS" scale deployment/ops-demo-loadgen --replicas=0 || true
kubectl --context "$CTX" -n "$NS" rollout status deployment/ops-demo-loadgen --timeout=60s || true
echo "Load generator OFF (scaled to 0)." 
