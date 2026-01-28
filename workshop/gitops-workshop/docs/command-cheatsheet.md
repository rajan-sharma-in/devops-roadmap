# Command Cheatsheet (simple, append fixes here)
This is the source of truth for commands students run. Keep it updated if we add bug fixes or tweaks.

> **GitHub + Docker Hub path (preferred for this repo)**
> 1) Ensure your self-hosted GitHub runner has Docker + kubectl and can reach the `kind-mautic` cluster.
> 2) Set repo secrets `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`, `KUBE_CONFIG_B64`.
> 3) Workflow `.github/workflows/ci.yml` uses `DOCKERHUB_USERNAME` (secrets/.envrc) to envsubst `k8s/02-deployment.yaml`, builds, pushes, and deploys.
>
> The sections below are the original GitLab bootstrap notes; keep for reference but they are not required when using GitHub Actions.

## Phase 0: Fresh kind cluster (start here after cleanup)
```sh
kind delete cluster --name mautic || true
cat >/tmp/kind-config-mautic.yaml <<'EOF'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
EOF
kind create cluster --name mautic --config /tmp/kind-config-mautic.yaml
kubectl config use-context kind-mautic

# Install ingress-nginx
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx --create-namespace \
  --set controller.ingressClassResource.default=true \
  --set controller.publishService.enabled=true

# Verify
kubectl --context kind-mautic get nodes
kubectl --context kind-mautic get pods -A | head
kubectl --context kind-mautic -n ingress-nginx get svc
```

## Phase 0: Fresh GitLab bootstrap (empty instance)
> Goal: create a group + project, register runner, and wire local repo — everything stays inside the cluster; only GitLab is exposed via port-forward.

1) Get initial root password from cluster:
   ```sh
   kubectl --context kind-mautic -n gitlab get secret gitlab-gitlab-initial-root-password -ojsonpath='{.data.password}' | base64 -d; echo
   ```
2) Start/confirm ingress port-forward (keep running):
   ```sh
   kubectl --context kind-mautic -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   ```
3) Login to GitLab UI at https://gitlab.host.docker.internal:8443 as `root`, change the password.
4) Create a Personal Access Token (PAT) with scopes `api`, `read_registry`, `write_registry`; store it locally:
   ```sh
   export GITLAB_PAT=<your-token>
   export GITLAB_API="https://gitlab.host.docker.internal:8443/api/v4"
   ```
5) Create a group (example `workshop`):
   ```sh
   curl --header "PRIVATE-TOKEN: $GITLAB_PAT" \
     --data "name=workshop&path=workshop&visibility=private" \
     "$GITLAB_API/groups"
   ```
6) Create the project inside that group (name must be `ops-workshop-stateless` for the workshop scripts):
   ```sh
   curl --header "PRIVATE-TOKEN: $GITLAB_PAT" \
     --data "name=ops-workshop-stateless&path=ops-workshop-stateless&namespace_id=$(curl -s --header \"PRIVATE-TOKEN: $GITLAB_PAT\" $GITLAB_API/groups?search=workshop | jq '.[0].id')" \
     "$GITLAB_API/projects"
   ```
7) (Optional) Add an SSH key so you can push/pull over SSH instead of HTTPS+PAT:
   ```sh
   ssh-keygen -t ed25519 -C "workshop" -f ~/.ssh/gitlab_workshop -N ""
   curl --header "PRIVATE-TOKEN: $GITLAB_PAT" \
     --data-urlencode "title=workshop-key" \
     --data-urlencode "key=$(cat ~/.ssh/gitlab_workshop.pub)" \
     "$GITLAB_API/user/keys"
   # then use: git remote set-url origin git@gitlab.host.docker.internal:workshop/ops-workshop-stateless.git
   ```
   If you prefer not to add an SSH key, continue using HTTPS with your PAT when prompted for username/password.
7) Register the in-cluster runner to this GitLab (if not already registered):
   - In UI: Admin -> Runners -> New runner -> copy the registration token.
   - Apply via Helm (runner already installed namespace `gitlab-runner`):
     ```sh
     export RUNNER_REG_TOKEN=<token-from-ui>
     helm repo add gitlab https://charts.gitlab.io
     helm repo update
     helm upgrade --install gitlab-runner gitlab/gitlab-runner \
       -n gitlab-runner --create-namespace \
       -f gitlab-runner-values.yaml \
       --set runnerRegistrationToken=$RUNNER_REG_TOKEN
     ```
8) Wire local repo to the new project (run inside workshop folder):
   ```sh
   git init
   git checkout -b main
   git remote add origin https://gitlab.host.docker.internal:8443/workshop/ops-workshop-stateless.git
   git add .
   git commit -m "bootstrap workshop"
   git push -u origin main
   ```
   (Use username `root` or a GitLab user; when prompted for password, use the PAT.)

## Phase 1: Cluster + Ingress (host ports 10080/10443)
Note: We removed the containerd registry mirror because containerd v2 in kind 0.31 breaks when inline mirrors are set. For local images, use `kind load docker-image ...`; for CI, pull from GitLab Registry. The cluster still reserves host ports 10080/10443 via `kind-config.yaml`, but we rely on port-forward 8080/8443 for GitLab access to avoid browser port blocks.
Hosts entry (on your laptop) to resolve the ingress hosts:
```sh
sudo sh -c 'echo "127.0.0.1 gitlab.host.docker.internal registry.host.docker.internal mautic.host.docker.internal ops.host.docker.internal" >> /etc/hosts'
```
Local image workflow:
- Build locally: `docker build -t mautic:dev .`
- Load into cluster: `kind load docker-image mautic:dev --name mautic`
- Set chart values to use that tag (e.g., `image.tag=dev`).
- Cluster sizing: use the multi-node config in `kind-config.yaml` (1 control-plane + 2 workers) to give GitLab enough capacity. Make sure Docker Desktop has ≥8–12 GB RAM allocated.

### macOS (Apple/Intel)
1) Install prerequisites:
   - Docker Desktop (start it).
   - `brew install kind kubectl helm`
2) Create kind cluster (uses `kind-config.yaml` with 10080/10443 host ports and 3 nodes):
   ```sh
   kind create cluster --name mautic --config kind-config.yaml
   ```
3) Install ingress controller:
   ```sh
   helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
   helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
     -n ingress-nginx --create-namespace \
     --set controller.ingressClassResource.default=true \
     --set controller.publishService.enabled=true
   ```
4) Verify:
   ```sh
   kubectl --context kind-mautic get nodes
   kubectl --context kind-mautic get pods -A
   kubectl --context kind-mautic get svc -n ingress-nginx
   ```
5) Cleanup:
   ```sh
   kind delete cluster --name mautic
   ```

### Linux (Ubuntu/Debian example)
1) Install prerequisites:
   ```sh
   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER && newgrp docker
   sudo apt-get update && sudo apt-get install -y ca-certificates curl
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.31.0/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
   ```
2) Create cluster (same as macOS):
   ```sh
   kind create cluster --name mautic --config kind-config.yaml
   ```
3) Install ingress and verify (same as macOS):
   ```sh
   helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
   helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
     -n ingress-nginx --create-namespace \
     --set controller.ingressClassResource.default=true \
     --set controller.publishService.enabled=true
   kubectl --context kind-mautic get nodes
   kubectl --context kind-mautic get pods -A
   kubectl --context kind-mautic get svc -n ingress-nginx
   ```
4) Cleanup:
   ```sh
   kind delete cluster --name mautic
   ```

### Windows (simplest: Docker Desktop + WSL2; run commands inside WSL)
1) Install prerequisites:
   - Enable WSL2 and install Ubuntu from Microsoft Store.
   - Install Docker Desktop for Windows and enable “Use the WSL 2 based engine”.
   - Inside WSL shell:
     ```sh
     sudo apt-get update && sudo apt-get install -y ca-certificates curl
     curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.31.0/kind-windows-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/
     curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
     curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
     ```
2) Run the same cluster/ingress/verify/cleanup commands as macOS/Linux inside the WSL shell:
   ```sh
   kind create cluster --name mautic --config kind-config.yaml
   helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
   helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
     -n ingress-nginx --create-namespace \
     --set controller.ingressClassResource.default=true \
     --set controller.publishService.enabled=true
   kubectl --context kind-mautic get nodes
   kubectl --context kind-mautic get pods -A
   kubectl --context kind-mautic get svc -n ingress-nginx
   ```

## GitHub Actions + Docker Hub (secrets and first push)
1) Set secrets on the GitHub repo `rajan-sharma-in/gitops-workshop` (requires `gh` CLI login):
   ```sh
   export DOCKERHUB_USERNAME=<your-dh-username>
   export DOCKERHUB_TOKEN=<your-dh-token>
   export KUBE_CONFIG_B64=$(cat ~/.kube/config | base64 | tr -d '\n')

   gh secret set DOCKERHUB_USERNAME --repo rajan-sharma-in/gitops-workshop --body "$DOCKERHUB_USERNAME"
   gh secret set DOCKERHUB_TOKEN    --repo rajan-sharma-in/gitops-workshop --body "$DOCKERHUB_TOKEN"
   gh secret set KUBE_CONFIG_B64    --repo rajan-sharma-in/gitops-workshop --body "$KUBE_CONFIG_B64"
   ```
2) Commit and push to trigger CI:
   ```sh
   git add .
   git commit -m "bootstrap github + dockerhub"
   git push origin main
   ```
3) Cleanup (inside WSL):
   ```sh
   kind delete cluster --name mautic
   ```

## Phase 2: GitLab CE + Runner + Registry (inside cluster, Helm)
Goal: deploy GitLab CE in-cluster via Helm, expose over ingress (port-forward 8080/8443), and register a Kubernetes executor runner.
Notes:
- GitLab needs resources. Use the multi-node kind config (control-plane + 2 workers) in `kind-config.yaml`. Ensure Docker Desktop has enough RAM (≥8–12 GB).
- Allow ingress-nginx admission webhook to become Ready before installing GitLab (wait for controller pod Ready).

### GitLab CE in-cluster (Helm)
1) Pre-req: keep ingress-nginx running. Start a port-forward in a separate terminal (maps ingress svc to host). Use browser-safe ports (8080/8443):
   ```sh
   kubectl --context kind-mautic -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   ```
2) Add repo and install with provided values (HTTP only, uses ingress, disables built-in nginx/cert-manager):
   ```sh
   helm repo add gitlab https://charts.gitlab.io
   helm repo update
   helm upgrade --install gitlab gitlab/gitlab -n gitlab --create-namespace -f ops-workshop-gitlab-selfhost/gitlab-values.yaml
   ```
   Notes: requires ≥4 CPU / 8–10GB RAM; waits a few minutes to stabilize.
3) Get initial root password:
   ```sh
   kubectl --context kind-mautic -n gitlab get secret gitlab-gitlab-initial-root-password -ojsonpath='{.data.password}' | base64 -d; echo
   ```
4) Access UI at https://gitlab.host.docker.internal:8443 (HTTPS only; avoids 422). Use incognito/cleared cookies if you see 422. Change root password, create group/project, enable Container Registry (Admin → Settings → Packages and registries).

### GitLab Runner (in-cluster, Kubernetes executor)
1) Get runner registration token from GitLab UI: Admin → Runners → New runner. Export it:
   ```sh
   export RUNNER_REG_TOKEN=<YOUR_TOKEN_HERE>
   ```
2) Install the runner via Helm (uses `gitlab-runner-values.yaml`, namespace `gitlab-runner`):
   ```sh
   helm repo add gitlab https://charts.gitlab.io
   helm repo update
   helm upgrade --install gitlab-runner gitlab/gitlab-runner \
     -n gitlab-runner --create-namespace \
     -f gitlab-runner-values.yaml \
     --set runnerRegistrationToken=$RUNNER_REG_TOKEN
   ```
3) Verify runner is registered and pod is running:
   ```sh
   kubectl --context kind-mautic -n gitlab-runner get pods
   # Check runners in GitLab UI under Admin → Runners.
   ```

### Verify registry (CLI)
Use internal endpoints from within the cluster/runners:
- GitLab API/UI: `http://gitlab-webservice-default.gitlab.svc:8181`
- Registry: `http://gitlab-registry.gitlab.svc:5000/<group>/<project>/<image>:tag`
No host-side docker login is required unless you want to push from your laptop (not needed for this workshop).

## Phase 3: Mautic (official image, simple install)
Use the official `mautic/mautic` image (no KISS/monolithic image). Deploy the app plus MariaDB in-cluster with ingress on the same controller.
1) Namespace + DB (Bitnami MariaDB) — set your own strong passwords:
   ```sh
   kubectl create namespace mautic
   helm repo add bitnami https://charts.bitnami.com/bitnami
   helm repo update
   export MAUTIC_DB_PASS=<set-strong-password>
   helm upgrade --install mautic-mariadb bitnami/mariadb -n mautic \
     --set auth.rootPassword=$MAUTIC_DB_PASS \
     --set auth.password=$MAUTIC_DB_PASS \
     --set auth.username=mautic \
     --set auth.database=mautic \
     --set primary.persistence.enabled=true \
     --set primary.persistence.size=8Gi
   ```
2) Deploy Mautic (official image) with ingress and PVCs for media/logs:
   ```sh
   kubectl --context kind-mautic apply -f mautic-app.yaml
   ```
   If you need to tweak values, edit `mautic-app.yaml` and re-apply:
   ```sh
   # file lives at repo root
   nano mautic-app.yaml
   kubectl --context kind-mautic apply -f mautic-app.yaml
   ```
   The file includes:
   - PVCs for media (5Gi) and logs (1Gi)
   - PVC for config (500Mi) mounted at `/var/www/html/config` so config survives image rebuilds
   - Deployment with correct DB env vars, trusted proxies JSON, and fsGroup 33
   - Service + Ingress on `mautic.host.docker.internal`
   ```yaml
   # See mautic-app.yaml for full manifest
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: mautic-media-pvc
   spec:
     accessModes:
       - ReadWriteOnce
     resources:
       requests:
         storage: 5Gi
   ---
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: mautic-logs-pvc
   spec:
     accessModes:
       - ReadWriteOnce
     resources:
         storage: 1Gi
   ---
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: mautic
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: mautic
     template:
       metadata:
         labels:
           app: mautic
       spec:
         securityContext:
           fsGroup: 33
         containers:
           - name: mautic
             image: mautic/mautic:latest
             ports:
               - containerPort: 80
             env:
               - name: MAUTIC_DB_HOST
                 value: mautic-mariadb.mautic.svc.cluster.local
               - name: MAUTIC_DB_PORT
                 value: "3306"
               - name: MAUTIC_DB_USER
                 value: mautic
               - name: MAUTIC_DB_PASSWORD
                 valueFrom:
                   secretKeyRef:
                     name: mautic-mariadb
                     key: mariadb-password
               - name: MAUTIC_DB_DATABASE
                 value: mautic
               - name: MAUTIC_TRUSTED_PROXIES
                 value: '["0.0.0.0/0","::/0"]'
               - name: MAUTIC_RUN_CRON_JOBS
                 value: "false"
               - name: MAUTIC_SITE_URL
                 value: http://mautic.host.docker.internal:8080
         volumeMounts:
           - name: media
             mountPath: /var/www/html/app/media
           - name: logs
             mountPath: /var/www/html/var/logs
         volumes:
           - name: media
             persistentVolumeClaim:
               claimName: mautic-media-pvc
           - name: logs
             persistentVolumeClaim:
               claimName: mautic-logs-pvc
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: mautic
   spec:
     selector:
       app: mautic
     ports:
       - port: 80
         targetPort: 80
         name: http
   ---
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: mautic
     annotations:
       kubernetes.io/ingress.class: nginx
   spec:
     rules:
       - host: mautic.host.docker.internal
         http:
           paths:
             - path: /
               pathType: Prefix
               backend:
                 service:
                   name: mautic
                   port:
                     number: 80
   EOF
   ```
3) Access Mautic via ingress (reuse the ingress-nginx port-forward):
   ```sh
   kubectl --context kind-mautic -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   # then open http://mautic.host.docker.internal:8080
   ```
4) (Optional) Load local Mautic image build into the cluster:
   ```sh
   docker build -t mautic:dev .
   kind load docker-image mautic:dev --name mautic
   kubectl --context kind-mautic -n mautic set image deployment/mautic mautic=mautic:dev
   ```
