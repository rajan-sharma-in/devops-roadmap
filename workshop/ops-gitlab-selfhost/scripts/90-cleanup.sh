#!/usr/bin/env bash
set -euo pipefail
CTX=${CTX:-kind-ops}
NS=ops-demo

kubectl --context "$CTX" delete namespace "$NS" --ignore-not-found
