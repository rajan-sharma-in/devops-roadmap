CTX ?= kind-ops

.PHONY: preflight apply watch load-on load-off cleanup

preflight:
	CTX=$(CTX) ./scripts/00-preflight.sh

apply:
	CTX=$(CTX) ./scripts/10-apply-k8s.sh

watch:
	CTX=$(CTX) ./scripts/30-watch.sh

load-on:
	CTX=$(CTX) ./scripts/20-load-on.sh

load-off:
	CTX=$(CTX) ./scripts/21-load-off.sh

cleanup:
	CTX=$(CTX) ./scripts/90-cleanup.sh
