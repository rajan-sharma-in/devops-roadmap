# Mautic Local Cluster Workshop Plan

This plan matches the manifests and values currently in the repo. It walks students through spinning up a kind-based Kubernetes cluster, installing GitLab CE with an in-cluster Kubernetes executor runner, and deploying Mautic (official image) with MariaDB. It intentionally avoids extras (custom images, observability stack) that are not present in the codebase.

## 0) Outcomes
- kind cluster (multi-node) with ingress-nginx ready and host ports 10080/10443 reserved.
- GitLab CE installed in-cluster via Helm, exposed through ingress (browser access via port-forward 8080/8443).
- GitLab Runner installed in-cluster, Kubernetes executor, privileged, wired to GitLab internal service.
- Mautic deployed using official `mautic/mautic:latest`, backed by Bitnami MariaDB, with PVCs for media/logs/config and ingress at `mautic.host.docker.internal`.

## 1) Architecture (as-coded)
- Cluster: kind, 1 control-plane + 2 workers (`kind-config.yaml`); ingress-ready label on control-plane; host ports 10080/10443 mapped to 80/443.
- Ingress: `ingress-nginx` installed via Helm; port-forward used for safe browser ports (8080/8443).
- GitLab: chart values in `ops-workshop-gitlab-selfhost/gitlab-values.yaml` (HTTP only, TLS disabled, nginx ingress class, registry enabled, external host `gitlab.host.docker.internal:8080`, shell SSH port 2222).
- Runner: chart values in `ops-workshop-gitlab-selfhost/gitlab-runner-values.yaml` (Kubernetes executor, privileged, namespace `gitlab-runner`, connects to `gitlab-webservice-default.gitlab.svc:8181`).
- Mautic app: `mautic-app.yaml` (PVCs for media/logs/config, Deployment using official image, Service, Ingress host `mautic.host.docker.internal`).
- Database: Bitnami MariaDB installed at deploy time (not checked into repo) with persistence enabled, credentials supplied via Helm values (env var in commands).

## 2) Prerequisites for students
- Hardware: ≥4 vCPU, 8–12 GB RAM allocated to Docker Desktop.
- Tools: Docker, kind, kubectl, Helm; optional k9s.
- Hosts file entry: `127.0.0.1 gitlab.host.docker.internal registry.host.docker.internal mautic.host.docker.internal`.
- macOS/Linux/Windows (WSL) install paths covered in commands cheat sheet.

## 3) Phases and commands (source of truth: `commands.md`)

### Phase 1: Cluster + ingress
- Create cluster: `kind create cluster --name mautic --config kind-config.yaml`.
- Install ingress-nginx:
  ```sh
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
  helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
    -n ingress-nginx --create-namespace \
    --set controller.ingressClassResource.default=true \
    --set controller.publishService.enabled=true
  ```
- Verify: `kubectl --context kind-mautic get nodes,pods -A`.

### Phase 2: GitLab CE (in-cluster)
- Port-forward ingress for browser-safe ports (keep running):
  ```sh
  kubectl --context kind-mautic -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
  ```
- Install GitLab with repo values: `helm upgrade --install gitlab gitlab/gitlab -n gitlab --create-namespace -f ops-workshop-gitlab-selfhost/gitlab-values.yaml`.
- Get root password: `kubectl --context kind-mautic -n gitlab get secret gitlab-gitlab-initial-root-password -ojsonpath='{.data.password}' | base64 -d; echo`.
- Access UI: https://gitlab.host.docker.internal:8443 (HTTP internally; browser hits via forwarded HTTPS). Change root password, create project/group, enable Container Registry.

### Phase 3: GitLab Runner (Kubernetes executor)
- Export runner registration token from GitLab UI: `export RUNNER_REG_TOKEN=<token>`.
- Install runner (values now live alongside the app workshop): `helm upgrade --install gitlab-runner gitlab/gitlab-runner -n gitlab-runner --create-namespace -f ops-workshop-gitlab-selfhost/gitlab-runner-values.yaml --set runnerRegistrationToken=$RUNNER_REG_TOKEN`.
- Verify: `kubectl --context kind-mautic -n gitlab-runner get pods`.

### Phase 4: Mautic + MariaDB
- Create namespace and deploy Bitnami MariaDB (set strong password):
  ```sh
  kubectl create namespace mautic
  helm repo add bitnami https://charts.bitnami.com/bitnami
  helm repo update
  export MAUTIC_DB_PASS=<strong-password>
  helm upgrade --install mautic-mariadb bitnami/mariadb -n mautic \
    --set auth.rootPassword=$MAUTIC_DB_PASS \
    --set auth.password=$MAUTIC_DB_PASS \
    --set auth.username=mautic \
    --set auth.database=mautic \
    --set primary.persistence.enabled=true \
    --set primary.persistence.size=8Gi
  ```
- Apply Mautic manifests: `kubectl --context kind-mautic apply -f mautic-app.yaml`.
- Access app (reuse ingress port-forward): http://mautic.host.docker.internal:8080.

### Phase 5: Optional local image iteration
- Build locally: `docker build -t mautic:dev .`
- Load into kind: `kind load docker-image mautic:dev --name mautic`
- Point Deployment to the new tag: `kubectl --context kind-mautic -n mautic set image deployment/mautic mautic=mautic:dev`

## 4) Deliverables checklist (tracked in repo)
- Cluster configs: `kind-config.yaml`, `kind-config-simple.yaml`.
- GitLab Helm values: `ops-workshop-gitlab-selfhost/gitlab-values.yaml`; Runner values: `ops-workshop-gitlab-selfhost/gitlab-runner-values.yaml`.
- App manifests: `mautic-app.yaml`.
- Student commands reference: `commands.md`.

## 5) Notes and caveats
- GitLab deploy is HTTP-only (TLS disabled) but ingress is accessed via HTTPS because of the port-forward path; certs are self-signed by ingress-nginx—use incognito if you see 422 errors.
- Runner uses privileged pods for Docker-in-Docker style builds; keep cluster resources in mind (≥8–12 GB RAM).
- Registry endpoint inside cluster: `http://gitlab-registry.gitlab.svc:5000/<group>/<project>/<image>:tag`. Host-side docker login not required unless pushing from your laptop.
