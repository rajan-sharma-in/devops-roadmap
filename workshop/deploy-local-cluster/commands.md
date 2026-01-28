# Command Cheatsheet (simple, append fixes here)
For step-by-step, student-friendly OS-specific setup, see `deploy-local-cluster/README.md`. This file stays as a concise command reference and place to append quick fixes.

## Phase 1: Cluster + Ingress (host ports 10080/10443)
Note: We removed the containerd registry mirror because containerd v2 in kind 0.31 breaks when inline mirrors are set. For local images, use `kind load docker-image ...`; for CI, pull from GitLab Registry. The cluster still reserves host ports 10080/10443 via `deploy-local-cluster/kind-config.yaml`, but we rely on port-forward 8080/8443 for GitLab access to avoid browser port blocks.
Hosts entry (on your laptop) to resolve the ingress hosts:
```sh
sudo sh -c 'echo "127.0.0.1 gitlab.host.docker.internal registry.host.docker.internal mautic.host.docker.internal" >> /etc/hosts'
```
Local image workflow:
- Build locally: `docker build -t mautic:dev .`
- Load into cluster: `kind load docker-image mautic:dev --name ops`
- Set chart values to use that tag (e.g., `image.tag=dev`).
- Cluster sizing: use the multi-node config in `kind-config.yaml` (1 control-plane + 2 workers) to give GitLab enough capacity. Make sure Docker Desktop has ≥8–12 GB RAM allocated.

### macOS (Apple/Intel)
1) Install prerequisites:
   - Docker Desktop (start it).
   - `brew install kind kubectl helm`
2) Create kind cluster (uses `deploy-local-cluster/kind-config.yaml` with 10080/10443 host ports and 3 nodes):
   ```sh
   kind create cluster --name ops --config deploy-local-cluster/kind-config.yaml
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
   kubectl --context kind-ops get nodes
   kubectl --context kind-ops get pods -A
   kubectl --context kind-ops get svc -n ingress-nginx
   ```
5) Cleanup:
   ```sh
   kind delete cluster --name ops
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
   kind create cluster --name ops --config deploy-local-cluster/kind-config.yaml
   ```
3) Install ingress and verify (same as macOS):
   ```sh
   helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
   helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
     -n ingress-nginx --create-namespace \
     --set controller.ingressClassResource.default=true \
     --set controller.publishService.enabled=true
   kubectl --context kind-ops get nodes
   kubectl --context kind-ops get pods -A
   kubectl --context kind-ops get svc -n ingress-nginx
   ```
4) Cleanup:
   ```sh
   kind delete cluster --name ops
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
   kind create cluster --name ops --config deploy-local-cluster/kind-config.yaml
   helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
   helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
     -n ingress-nginx --create-namespace \
     --set controller.ingressClassResource.default=true \
     --set controller.publishService.enabled=true
   kubectl --context kind-ops get nodes
   kubectl --context kind-ops get pods -A
   kubectl --context kind-ops get svc -n ingress-nginx
   ```
3) Cleanup (inside WSL):
   ```sh
   kind delete cluster --name ops
   ```

## Phase 2: GitLab CE + Registry (inside cluster, Helm)
Goal: deploy GitLab CE in-cluster via Helm and expose it over ingress (port-forward 8080/8443). GitLab Runner setup now lives in `ops-workshop-gitlab-selfhost` (see its docs/command-cheatsheet).
Notes:
- GitLab needs resources. Use the multi-node kind config (control-plane + 2 workers) in `kind-config.yaml`. Ensure Docker Desktop has enough RAM (≥8–12 GB).
- Allow ingress-nginx admission webhook to become Ready before installing GitLab (wait for controller pod Ready).

### GitLab CE in-cluster (Helm)
1) Pre-req: keep ingress-nginx running. Start a port-forward in a separate terminal (maps ingress svc to host). Use browser-safe ports (8080/8443):
   ```sh
   kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
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
   kubectl --context kind-ops -n gitlab get secret gitlab-gitlab-initial-root-password -ojsonpath='{.data.password}' | base64 -d; echo
   ```
4) Access UI at https://gitlab.host.docker.internal:8443 (HTTPS only; avoids 422). Use incognito/cleared cookies if you see 422. Change root password, create group/project, enable Container Registry (Admin → Settings → Packages and registries).

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
   kubectl --context kind-ops apply -f deploy-mautic/mautic-app.yaml
   ```
   If you need to tweak values, edit `mautic-app.yaml` and re-apply:
   ```sh
   nano deploy-mautic/mautic-app.yaml
   kubectl --context kind-ops apply -f deploy-mautic/mautic-app.yaml
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
       requests:
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
   kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   # then open http://mautic.host.docker.internal:8080
   ```
4) (Optional) Load local Mautic image build into the cluster:
   ```sh
   docker build -t mautic:dev .
   kind load docker-image mautic:dev --name ops
   kubectl --context kind-ops -n mautic set image deployment/mautic mautic=mautic:dev
   ```
