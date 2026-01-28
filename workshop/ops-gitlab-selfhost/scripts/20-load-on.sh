#!/usr/bin/env bash
set -euo pipefail
CTX=${CTX:-kind-ops}
NS=ops-demo

kubectl --context "$CTX" -n "$NS" apply -f "$(cd "$(dirname "$0")/.." && pwd)/k8s/08-loadgen.yaml"
kubectl --context "$CTX" -n "$NS" scale deployment/ops-demo-loadgen --replicas=1

echo "Load generator ON (1 replica)." 
