#!/usr/bin/env bash
set -euo pipefail
CTX=${CTX:-kind-ops}
NS=ops-demo

if command -v watch >/dev/null 2>&1; then
  watch -n2 "kubectl --context $CTX -n $NS get pods,deploy,svc,hpa,ingress"
else
  echo "watch not found; falling back to loop (Ctrl+C to exit)" >&2
  while true; do
    date
    kubectl --context "$CTX" -n "$NS" get pods,deploy,svc,hpa,ingress
    sleep 2
    echo "------"
  done
fi
