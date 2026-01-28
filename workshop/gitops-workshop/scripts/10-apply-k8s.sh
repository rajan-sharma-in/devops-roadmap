#!/usr/bin/env bash
set -euo pipefail
CTX=${CTX:-kind-workshop-cluster}
BASE_DIR=$(cd "$(dirname "$0")/.." && pwd)

msg() { echo "[apply] $*"; }

kubectl --context "$CTX" apply -f "$BASE_DIR/k8s/00-namespace.yaml"
kubectl --context "$CTX" apply -f "$BASE_DIR/k8s/01-app-config.yaml"
kubectl --context "$CTX" apply -f "$BASE_DIR/k8s/03-service.yaml"
kubectl --context "$CTX" apply -f "$BASE_DIR/k8s/04-ingress.yaml"
kubectl --context "$CTX" apply -f "$BASE_DIR/k8s/05-hpa.yaml"
kubectl --context "$CTX" apply -f "$BASE_DIR/k8s/06-pdb.yaml"
kubectl --context "$CTX" apply -f "$BASE_DIR/k8s/02-deployment.yaml"
kubectl --context "$CTX" apply -f "$BASE_DIR/k8s/07-runner-rbac.yaml"

msg "If the pipeline already built an image, set it now with: kubectl --context $CTX -n ops-demo set image deployment/ops-demo ops-demo=<IMAGE>"
msg "Access app via http://ops.host.docker.internal:8080 after port-forward is up."
