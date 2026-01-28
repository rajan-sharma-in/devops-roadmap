#!/usr/bin/env bash
set -euo pipefail
CTX=${CTX:-kind-ops}
NS=ops-demo

msg() { echo "[preflight] $*"; }

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl not found in PATH" >&2
  exit 1
fi

current_ctx=$(kubectl config current-context 2>/dev/null || true)
if [[ "$current_ctx" != "$CTX" ]]; then
  msg "Switching kubectl context to $CTX (was: ${current_ctx:-unset})"
  kubectl config use-context "$CTX"
fi

msg "Verifying cluster connectivity"
kubectl --context "$CTX" get nodes -o wide

msg "Checking ingress-nginx service"
kubectl --context "$CTX" -n ingress-nginx get svc ingress-nginx-controller -o wide

msg "Checking GitLab Runner namespace"
kubectl --context "$CTX" -n gitlab-runner get pods -l app=gitlab-runner

msg "K8s API server version"
kubectl --context "$CTX" version --output=yaml

msg "Creating namespace if missing"
kubectl --context "$CTX" get ns "$NS" >/dev/null 2>&1 || kubectl --context "$CTX" create ns "$NS"

msg "Reminder: ensure port-forward running: kubectl --context $CTX -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443"
