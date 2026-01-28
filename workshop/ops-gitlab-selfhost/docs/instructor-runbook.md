# DevOps in Practice: Building & Scaling Systems — Instructor Runbook

## Prework checklist (complete before students join)
- [ ] Verify hosts entry exists on your laptop: `127.0.0.1 gitlab.host.docker.internal registry.host.docker.internal ops.host.docker.internal`.
- [ ] Start ingress port-forward (keep running): `kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443`.
- [ ] GitLab bootstrap (fresh instance): retrieve root password from secret, login, set new password.
- [ ] Create PAT with scopes `api, read_registry, write_registry`.
- [ ] Create group `workshop` and project `ops-workshop-stateless` (curl or UI).
- [ ] Register runner with the new GitLab using its registration token (Helm upgrade for gitlab-runner).
- [ ] Confirm GitLab UI reachable at https://gitlab.host.docker.internal:8443.
- [ ] Confirm GitLab Runner pods are Ready in namespace `gitlab-runner` and show as online in UI.
- [ ] Confirm registry DNS works in cluster: `kubectl --context kind-ops -n ops-demo run dnscheck --image=busybox:1.36 --restart=Never -- nslookup gitlab-registry.gitlab.svc`.
- [ ] Ensure metrics-server installed for HPA.
- [ ] Clone repo template into the freshly created GitLab project: name `ops-workshop-stateless`.

## 90-minute flow (approximate)
- 00:00–00:05: Welcome, goals, architecture overview (show `docs/architecture.md`).
- 00:05–00:10: Confirm port-forward + hosts entry; students open GitLab UI.
- 00:10–00:20: Students fork/clone repo into their GitLab group and locally.
- 00:20–00:25: Run `make preflight`; explain context switching to `kind-ops`.
- 00:25–00:35: Walk through `app/main.go`, probes, metrics, graceful shutdown.
- 00:35–00:45: Explain `.gitlab-ci.yml` (Kaniko, insecure registry, deploy order).
- 00:45–00:55: Students push to `main`, watch `build` then `deploy` job.
- 00:55–01:05: Verify service via curl and browser at `http://ops.host.docker.internal:8080`.
- 01:05–01:15: Turn load on with `make load-on`; observe HPA scale using `make watch`.
- 01:15–01:20: Turn load off; observe scale down.
- 01:20–01:25: Q&A; review common failure modes.
- 01:25–01:30: Cleanup with `make cleanup`; stop port-forward.

## Common issues and fixes
- Ingress host resolution fails: re-add `/etc/hosts` entry; ensure port-forward is active.
- Port-forward not running: restart command and keep terminal open; verify `kubectl --context kind-ops get svc -n ingress-nginx` works.
- Registry auth errors in build stage: confirm CI variables `CI_REGISTRY_USER/CI_REGISTRY_PASSWORD` exist (GitLab provides by default); ensure Kaniko flags `--insecure --skip-tls-verify` present.
- Deploy job RBAC denied: make sure `k8s/07-runner-rbac.yaml` applied and runner service account is `gitlab-runner` in namespace `gitlab-runner`.
- HPA not scaling: install metrics-server; verify CPU load exists (`kubectl top pods -n ops-demo`); ensure load generator scaled >0.
- Image pull errors: double-check registry host `gitlab-registry.gitlab.svc:5000` reachable from cluster; retry deploy after build completes.
