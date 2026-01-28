# Local Kind Cluster — Step-by-Step (Windows, macOS, Ubuntu)

This guide walks students through creating a multi-node Kind cluster that can host GitLab, a demo app, and optional Mautic. It is intentionally verbose and copy/paste friendly.

**What you get**
- Kind cluster named `ops` (1 control-plane + 2 workers) using `deploy-local-cluster/kind-config.yaml`.
- Ingress-NGINX installed and reachable via port-forward on 8080/8443.
- Ready for GitLab CE + Runner using provided Helm values.

**Prereqs (all OS)**
- 8–12 GB RAM available to Docker/WSL.
- Stable internet for image pulls.
- Ports 8080/8443 free on your host (used by port-forward).

---

## macOS (Apple/Intel)
1) Install Docker Desktop and start it.
2) Install CLIs:
   ```sh
   brew install kind kubectl helm
   ```
3) Create the cluster:
   ```sh
   kind create cluster --name ops --config deploy-local-cluster/kind-config.yaml
   ```
4) Install ingress-nginx:
   ```sh
   helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
   helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
     -n ingress-nginx --create-namespace \
     --set controller.ingressClassResource.default=true \
     --set controller.publishService.enabled=true
   ```
5) Verify:
   ```sh
   kubectl --context kind-ops get nodes
   kubectl --context kind-ops get pods -A
   kubectl --context kind-ops get svc -n ingress-nginx
   ```
6) Port-forward for browsers (keep running):
   ```sh
   kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   ```
7) Cleanup:
   ```sh
   kind delete cluster --name ops
   ```

---

## Ubuntu / Debian (WSL or native)
1) Install Docker Engine:
   ```sh
   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER && newgrp docker
   ```
2) Install Kind, kubectl, Helm:
   ```sh
   curl -Lo kind https://kind.sigs.k8s.io/dl/v0.31.0/kind-linux-amd64 && chmod +x kind && sudo mv kind /usr/local/bin/
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
   ```
3) Create cluster + install ingress (same commands as macOS):
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
4) Port-forward (keep running):
   ```sh
   kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   ```
5) Cleanup:
   ```sh
   kind delete cluster --name ops
   ```

---

## Windows 11/10 (Docker Desktop + WSL2)
1) Enable WSL2, install Ubuntu from Microsoft Store.
2) Install Docker Desktop for Windows and enable “Use the WSL 2 based engine”.
3) Open the Ubuntu (WSL) shell and install CLIs:
   ```sh
   sudo apt-get update && sudo apt-get install -y ca-certificates curl
   curl -Lo kind https://kind.sigs.k8s.io/dl/v0.31.0/kind-windows-amd64 && chmod +x kind && sudo mv kind /usr/local/bin/
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
   ```
4) Create cluster + install ingress (run inside WSL):
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
5) Port-forward (inside WSL, keep running):
   ```sh
   kubectl --context kind-ops -n ingress-nginx port-forward svc/ingress-nginx-controller 8080:80 8443:443
   ```
6) Cleanup:
   ```sh
   kind delete cluster --name ops
   ```

---

## Hosts entries (all OS)
Add hostnames once so GitLab/registry/app ingress rules resolve:
```sh
sudo sh -c 'echo "127.0.0.1 gitlab.host.docker.internal registry.host.docker.internal ops.host.docker.internal mautic.host.docker.internal" >> /etc/hosts'
```

---

## Troubleshooting
- **Ingress pods Pending**: check Docker Desktop/WSL memory; ensure at least 8 GB allocated.
- **Port 8080/8443 in use**: stop other services (e.g., minikube, nginx) or change port-forward command.
- **Image pulls slow**: run `docker pull` for large base images before running Helm installs to warm cache.
- **Context not found**: verify cluster exists with `kind get clusters`; recreate if needed.
