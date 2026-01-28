#!/usr/bin/env bash
set -euo pipefail
CTX=${CTX:-kind-workshop-cluster}
NS=ops-demo

kubectl --context "$CTX" -n "$NS" apply -f "$(cd "$(dirname "$0")/.." && pwd)/k8s/08-loadgen.yaml"
kubectl --context "$CTX" -n "$NS" rollout status deployment/ops-demo-loadgen --timeout=60s || true
kubectl --context "$CTX" -n "$NS" scale deployment/ops-demo-loadgen --replicas=4
kubectl --context "$CTX" -n "$NS" rollout status deployment/ops-demo-loadgen --timeout=60s || true

echo "Load generator ON (4 replicas, bursty curl loops)." 
