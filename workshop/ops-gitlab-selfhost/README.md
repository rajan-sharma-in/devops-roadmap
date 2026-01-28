# DevOps in Practice: Building & Scaling Systems

A 90-minute DevOps-ready workshop using a stateless Go app on a local kind cluster (`kind-ops`). CI builds with Kaniko (no DinD), pushes to the in-cluster GitLab registry, and deploys to Kubernetes with rolling updates, probes, HPA, and PDB.

## Quickstart
0) Fresh GitLab? bootstrap it (group, project, runner) via `docs/command-cheatsheet.md` Phase 0, then come back.
1) Add hosts (once): `127.0.0.1 gitlab.host.docker.internal registry.host.docker.internal ops.host.docker.internal`.
2) Start ingress port-forward (keep open):
   ```bash
   kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   ```
3) Clone your GitLab project `ops-workshop-stateless` and enter the repo (or run the Phase 0 curl-based creation then `git clone ...`).
4) Run `make preflight` to validate cluster + context.
5) Push to `main` to trigger CI (`build` -> `deploy`).
6) Access the app at http://ops.host.docker.internal:8080.

## Make targets
- `make preflight` – verify context, ingress, runner, namespace.
- `make apply` – apply Kubernetes manifests locally (optional; CI also does this).
- `make watch` – watch pods/service/HPA/ingress.
- `make load-on` / `make load-off` – toggle load generator to observe scaling.
- `make cleanup` – remove `ops-demo` namespace.

## What’s inside
- Go app with JSON logging, health/readiness, Prometheus metrics, graceful shutdown.
- Kubernetes manifests: namespace, ConfigMap, Deployment (rolling), Service, Ingress, HPA, PDB, RBAC, load generator.
- GitLab CI: Kaniko build to internal registry (insecure HTTP) + kubectl deploy.
- Docs: architecture diagram, student handout, instructor runbook.

## Docs
- Architecture: `docs/architecture.md`
- Student commands: `docs/student-handout.md`
- Instructor runbook: `docs/instructor-runbook.md`

## Changing the message
Edit `k8s/01-app-config.yaml` or update the ConfigMap directly:
```bash
kubectl --context kind-ops -n ops-demo set data configmap/ops-demo-config MESSAGE="Hello from <name>"
kubectl --context kind-ops -n ops-demo rollout restart deploy/ops-demo
```

Happy shipping!
