<<<<<<< HEAD
# gitops-workshop
gitops workshop
=======
# GitOps Workshop (GitHub + Docker Hub)

Stateless Go app on a local kind cluster (`kind-mautic`). GitHub Actions builds with Buildx, pushes to Docker Hub, and deploys to Kubernetes with rolling updates, probes, HPA, and PDB.

## Quickstart
1) Prereqs on your laptop:
   - kind cluster `kind-mautic` with ingress-nginx installed and reachable (same as existing workshop).
   - Docker daemon running (for the self-hosted GitHub runner).
   - kubectl pointing at `kind-mautic`.
2) Create secrets in your GitHub repo:
   - `DOCKERHUB_USERNAME`
   - `DOCKERHUB_TOKEN` (PAT with write perms)
   - `KUBE_CONFIG_B64` (base64 of `~/.kube/config` that can reach kind)
3) (Recommended) Register a self-hosted runner on your laptop and label it `self-hosted`.
4) Commit and push to GitHub `main` — the workflow builds/pushes the image and deploys to the cluster (it uses `DOCKERHUB_USERNAME` via envsubst).
6) Access the app via your ingress host (e.g., http://ops.host.docker.internal:8080).

## Make targets
- `make preflight` – verify context and namespace.
- `make apply` – apply Kubernetes manifests locally.
- `make watch` – watch pods/service/HPA/ingress.
- `make load-on` / `make load-off` – toggle load generator to observe scaling.
- `make cleanup` – remove `ops-demo` namespace.

## CI/CD (GitHub Actions)
- `.github/workflows/ci.yml` builds with Buildx, pushes to Docker Hub, then deploys with kubectl using your provided kubeconfig.
- Images are tagged `latest` and `${GITHUB_SHA}`; deploy stage pins the deployment to `${GITHUB_SHA}`.

## What’s inside
- Go app with JSON logging, health/readiness, Prometheus metrics, graceful shutdown.
- Kubernetes manifests: namespace, ConfigMap, Deployment (rolling), Service, Ingress, HPA, PDB, RBAC, load generator.
- Docs: architecture diagram, student handout, instructor runbook (GitLab references can be ignored for this GitHub/Docker Hub path).

## Changing the message
Edit `k8s/01-app-config.yaml` or update the ConfigMap directly:
```bash
kubectl --context kind-mautic -n ops-demo set data configmap/ops-demo-config MESSAGE="Hello from <name>"
kubectl --context kind-mautic -n ops-demo rollout restart deploy/ops-demo
```

Happy shipping!
>>>>>>> 33a42f9 (initial comment)
